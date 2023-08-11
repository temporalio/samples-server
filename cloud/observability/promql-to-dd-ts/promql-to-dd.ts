import { client, v2 } from '@datadog/datadog-api-client'
import axios from 'axios'
import { Agent } from 'https'
import { z } from 'zod'

// This automatically pulls API keys from env vars DD_API_KEY
const configuration = client.createConfiguration()

client.setServerVariables(configuration, {
    site: 'us5.datadoghq.com',
})
const datadogMetricsApi = new v2.MetricsApi(configuration)

const requireEnvVar = (variableName: string): string => {
    const value = process.env[variableName]

    if (!value) {
        throw new Error(`Missing environment variable ${variableName}`)
    }

    return value
}

const setTimeoutAsync = async (millis: number): Promise<void> => {
    return new Promise(resolve => setTimeout(resolve, millis))
}

const httpsAgent = new Agent({
    cert: Buffer.from(
        Buffer.from(requireEnvVar('TEMPORAL_OBSERVABILITY_CERT'), 'base64').toString(),
    ),
    key: Buffer.from(
        Buffer.from(requireEnvVar('TEMPORAL_OBSERVABILITY_KEY'), 'base64').toString(),
    ),
})

// This allows people to do local development on this service without polluting the metrics which
// are critical to ongoing LXB production observability.
const DATADOG_METRIC_PREFIX = process.env.TEST_METRIC_PREFIX || ''

const TEMPORAL_CLOUD_BASE_URL = 'https://lvpqy.tmprl.cloud'
const PROM_LABELS_URL = `${TEMPORAL_CLOUD_BASE_URL}/prometheus/api/v1/label/__name__/values`
const PROM_QUERY_URL = `${TEMPORAL_CLOUD_BASE_URL}/prometheus/api/v1/query_range`

// We're going to query Prometheus with a resolution of 1 minute
const PROMETHEUS_STEP_SECONDS = 60

// On an ongoing basis, query only for the last 10 minutes of data.
const QUERY_WINDOW_SECONDS = 10 * 60

const HISTOGRAM_QUANTILES = [0.5, 0.9, 0.95, 0.99]

const basePrometheusQueryParams = {
    step: PROMETHEUS_STEP_SECONDS.toFixed(0),
    format: 'json',
}

const labelsResponseDataSchema = z.object({
    status: z.literal('success'),
    data: z.string().array(),
})

const getMetricNames = async (): Promise<{
    countMetricNames: string[]
    histogramMetricNames: string[]
}> => {
    const metricNamesResponse = await axios.get(PROM_LABELS_URL, {
        httpsAgent,
    })
    const { data: metricNames } = labelsResponseDataSchema.parse(metricNamesResponse.data)

    const temporalCloudMetricNames = metricNames.filter(
        metricName => metricName.startsWith('temporal_cloud')
    )

    const countMetricNames = temporalCloudMetricNames.filter(
        metricName => metricName.endsWith('_count')
    )
    const histogramMetricNames = temporalCloudMetricNames.filter(
        metricName => metricName.endsWith('_bucket')
    )

    return { countMetricNames, histogramMetricNames }
}

const queryResponseDataSchema = z.object({
    status: z.literal('success'),
    data: z.object({
        resultType: z.literal('matrix'),
        result: z.object({
            metric: z.record(z.string()),
            values: z.tuple([z.number(), z.string()]).array(),
        }).array(),
    }),
})

type MetricData = z.infer<typeof queryResponseDataSchema>['data']

type QueryWindow = {
    startSecondsSinceEpoch: number
    endSecondsSinceEpoch: number
}

const generateQueryWindow = (): QueryWindow => {
    const endSecondsSinceEpoch = Date.now() / 1000

    const windowInSeconds = QUERY_WINDOW_SECONDS
    const startSecondsSinceEpoch = (endSecondsSinceEpoch - windowInSeconds)

    return alignQueryWindowOnPrometheusStep({ startSecondsSinceEpoch, endSecondsSinceEpoch })
}

// I'm not exactly sure why this is important, but I think that without it, we may inaccurately
// report some metrics.
const alignQueryWindowOnPrometheusStep = (queryWindow: QueryWindow): QueryWindow => {
    const startSecondsSinceEpoch = Math.floor(
        queryWindow.startSecondsSinceEpoch / PROMETHEUS_STEP_SECONDS - 1
    ) * PROMETHEUS_STEP_SECONDS
    const endSecondsSinceEpoch = Math.floor(
        queryWindow.endSecondsSinceEpoch / PROMETHEUS_STEP_SECONDS + 1
    ) * PROMETHEUS_STEP_SECONDS

    return { startSecondsSinceEpoch, endSecondsSinceEpoch }
}

const queryPrometheusCount = async (
    metricName: string,
    queryWindow: QueryWindow
): Promise<MetricData> => {
    const response = await axios.get(PROM_QUERY_URL, {
        httpsAgent,
        params: {
            ...basePrometheusQueryParams,
            query: `rate(${metricName}[1m])`,
            start: queryWindow.startSecondsSinceEpoch.toFixed(0),
            end: queryWindow.endSecondsSinceEpoch.toFixed(0),
        },
    })

    return queryResponseDataSchema.parse(response.data).data
}

const convertPrometheusCountToDatadogRateSeries = (
    metricName: string,
    metricData: MetricData,
): v2.MetricSeries[] =>
    metricData.result.map(prometheusMetric => ({
        // We need to inform datadog of the interval for this metric
        interval: PROMETHEUS_STEP_SECONDS,
        // Make it easier for the datadog user to understand what this metric is
        metric: DATADOG_METRIC_PREFIX + metricName.split('_count')[0] + '_rate1m',
        // Type 2 is a "rate" metric
        type: 2,
        points: prometheusMetric.values.map(([timestamp, value]) => {
            return {
                timestamp: timestamp,
                value: parseFloat(value),
            }
        }),
        tags: Object.entries(prometheusMetric.metric)
            .filter(([key]) => key !== '__rollup__')
            // Datadog tags can't be longer than 200 characters
            .map(([key, value]) => `${key}:${value.substring(0, 200)}`),
    }))

const queryPrometheusHistogram = async (
    metricName: string,
    quantile: number,
    queryWindow: QueryWindow,
): Promise<MetricData> => {
    const response = await axios.get(PROM_QUERY_URL, {
        httpsAgent,
        params: {
            ...basePrometheusQueryParams,
            query: `histogram_quantile(${quantile}, sum(rate(${metricName}[1m])) by (temporal_account,temporal_namespace,operation,le))`,
            start: queryWindow.startSecondsSinceEpoch.toFixed(0),
            end: queryWindow.endSecondsSinceEpoch.toFixed(0),
        },
    })

    return queryResponseDataSchema.parse(response.data).data
}

const convertPrometheusHistogramToDatadogGuageSeries = (
    metricName: string,
    quantile: number,
    metricData: MetricData,
): v2.MetricSeries[] =>
    metricData.result.map(prometheusMetric => ({
        // Make it easier for the datadog user to understand what this metric is
        metric: DATADOG_METRIC_PREFIX + metricName.split('_bucket')[0] + '_P' + quantile * 100,
        // Type 2 is a "guage" metric
        type: 3,
        points: prometheusMetric.values.map(([timestamp, value]) => {
            return {
                timestamp: timestamp,
                value: parseFloat(value),
            }
        }),
        tags: Object.entries(prometheusMetric.metric)
            .filter(([key]) => key !== '__rollup__')
            // Datadog tags can't be longer than 200 characters
            .map(([key, value]) => `${key}:${value.substring(0, 200)}`),
    }))


const main = async () => {
    const {
        countMetricNames,
        histogramMetricNames,
    } = await getMetricNames()

    console.log({
        level: 'info',
        message: 'Polling metrics',
        countMetricNames,
        histogramMetricNames,
    })

    while (true) {
        const queryWindow = generateQueryWindow()

        console.log({
            level: 'info',
            message: 'Collecting metrics from temporal cloud.',
            startDate: new Date(queryWindow.startSecondsSinceEpoch * 1000).toISOString(),
            endDate: new Date(queryWindow.endSecondsSinceEpoch * 1000).toISOString(),
        })

        const countSeries = (await Promise.all(
            countMetricNames.map(async metricName =>
                convertPrometheusCountToDatadogRateSeries(
                    metricName,
                    await queryPrometheusCount(metricName, generateQueryWindow())
                )
            ),
        )).flat()

        const guageSeries = (await Promise.all(histogramMetricNames.map(async metricName =>
            Promise.all(HISTOGRAM_QUANTILES.map(async quantile =>
                convertPrometheusHistogramToDatadogGuageSeries(
                    metricName,
                    quantile,
                    await queryPrometheusHistogram(metricName, quantile, generateQueryWindow())
                )
            ))
        ))).flat(2)

        console.log({ level: 'info', message: 'Submitting metrics to Datadog' })
        await datadogMetricsApi.submitMetrics({ body: { series: [
            ...countSeries,
            ...guageSeries,
        ]}})

        console.log({ level: 'info', message: 'Pausing for 20s' })
        await setTimeoutAsync(20 * 1000)
    }
}

main().catch(error => {
    console.log({
        level: 'error',
        message: 'Error in main loop. Closing healthcheck server.',
        error,
    })

    server.close(() => {
        console.log({
            level: 'info',
            message: 'Healthcheck server closed',
        })

        process.exit(1)
    })
})


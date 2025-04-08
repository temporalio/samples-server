import _ from "lodash";
import { Agent } from "https";
import { z } from "zod";
import axios from "axios";

const requireEnvVar = (variableName: string): string => {
  const value = process.env[variableName];
  if (!value) {
    throw new Error(`Missing environment variable ${variableName}`);
  }
  return value;
};

const newRelicMetricsApi = axios.create({
  baseURL: "https://metric-api.newrelic.com",
  timeout: 3000,
  headers: { "Api-Key": requireEnvVar("NEW_RELIC_API_KEY") },
});

const setTimeoutAsync = async (millis: number): Promise<void> => {
  return new Promise((resolve) => setTimeout(resolve, millis));
};

const httpsAgent = new Agent({
  cert: Buffer.from(
    Buffer.from(
      requireEnvVar("TEMPORAL_OBSERVABILITY_CERT"),
      "base64"
    ).toString()
  ),
  key: Buffer.from(
    Buffer.from(
      requireEnvVar("TEMPORAL_OBSERVABILITY_KEY"),
      "base64"
    ).toString()
  ),
});

// This allows people to do local development on this service without polluting the metrics which
// are critical to ongoing LXB production observability.
const NEWRELIC_METRIC_PREFIX = process.env.TEST_METRIC_PREFIX || "";

const TEMPORAL_CLOUD_BASE_URL = process.env.TEMPORAL_CLOUD_BASE_URL ?? "";
const PROM_LABELS_URL = `${TEMPORAL_CLOUD_BASE_URL}/prometheus/api/v1/label/__name__/values`;
const PROM_QUERY_URL = `${TEMPORAL_CLOUD_BASE_URL}/prometheus/api/v1/query_range`;

// We're going to query Prometheus with a resolution of 1 minute
const PROMETHEUS_STEP_SECONDS = 60;

// On an ongoing basis, query only for the last 10 minutes of data.
const QUERY_WINDOW_SECONDS = 10 * 60;

const HISTOGRAM_QUANTILES = [0.5, 0.9, 0.95, 0.99];

const basePrometheusQueryParams = {
  step: PROMETHEUS_STEP_SECONDS.toFixed(0),
  format: "json",
};

const labelsResponseDataSchema = z.object({
  status: z.literal("success"),
  data: z.string().array(),
});

const getMetricNames = async (): Promise<{
  countMetricNames: string[];
  histogramMetricNames: string[];
}> => {
  const metricNamesResponse = await axios.get(PROM_LABELS_URL, {
    httpsAgent,
  });
  const { data: metricNames } = labelsResponseDataSchema.parse(
    metricNamesResponse.data
  );

  const temporalCloudMetricNames = metricNames.filter((metricName) =>
    metricName.startsWith("temporal_cloud")
  );

  const countMetricNames = temporalCloudMetricNames.filter((metricName) =>
    metricName.endsWith("_count")
  );
  const histogramMetricNames = temporalCloudMetricNames.filter((metricName) =>
    metricName.endsWith("_bucket")
  );

  return { countMetricNames, histogramMetricNames };
};

const queryResponseDataSchema = z.object({
  status: z.literal("success"),
  data: z.object({
    resultType: z.literal("matrix"),
    result: z
      .object({
        metric: z.record(z.string()),
        values: z.tuple([z.number(), z.string()]).array(),
      })
      .array(),
  }),
});

type MetricData = z.infer<typeof queryResponseDataSchema>["data"];

type QueryWindow = {
  startSecondsSinceEpoch: number;
  endSecondsSinceEpoch: number;
};

const generateQueryWindow = (): QueryWindow => {
  const endSecondsSinceEpoch = Date.now() / 1000;

  const windowInSeconds = QUERY_WINDOW_SECONDS;
  const startSecondsSinceEpoch = endSecondsSinceEpoch - windowInSeconds;

  return alignQueryWindowOnPrometheusStep({
    startSecondsSinceEpoch,
    endSecondsSinceEpoch,
  });
};

// I'm not exactly sure why this is important, but I think that without it, we may inaccurately
// report some metrics.
const alignQueryWindowOnPrometheusStep = (
  queryWindow: QueryWindow
): QueryWindow => {
  const startSecondsSinceEpoch =
    Math.floor(
      queryWindow.startSecondsSinceEpoch / PROMETHEUS_STEP_SECONDS - 1
    ) * PROMETHEUS_STEP_SECONDS;
  const endSecondsSinceEpoch =
    Math.floor(queryWindow.endSecondsSinceEpoch / PROMETHEUS_STEP_SECONDS + 1) *
    PROMETHEUS_STEP_SECONDS;

  return { startSecondsSinceEpoch, endSecondsSinceEpoch };
};

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
  });

  return queryResponseDataSchema.parse(response.data).data;
};

type NewRelicMetric = {
  name: string;
  type: "count" | "gauge";
  value: number;
  timestamp: number;
  attributes: { [key: string]: number | string };
};

const convertPrometheusCountToNewRelicCountMetrics = (
  metricName: string,
  metricData: MetricData
): NewRelicMetric[][] =>
  metricData.result.map((prometheusMetric) =>
    prometheusMetric.values.map((v) => ({
      name: NEWRELIC_METRIC_PREFIX + metricName.split("_count")[0] + "_rate1m",
      type: "count",
      value: parseFloat(v[1]),
      timestamp: v[0],
      "interval.ms": PROMETHEUS_STEP_SECONDS * 1000,
      attributes: _.omit(prometheusMetric.metric, "__rollup__"),
    }))
  );

const queryPrometheusHistogram = async (
  metricName: string,
  quantile: number,
  queryWindow: QueryWindow
): Promise<MetricData> => {
  const response = await axios.get(PROM_QUERY_URL, {
    httpsAgent,
    params: {
      ...basePrometheusQueryParams,
      query: `histogram_quantile(${quantile}, sum(rate(${metricName}[1m])) by (temporal_account,temporal_namespace,operation,le))`,
      start: queryWindow.startSecondsSinceEpoch.toFixed(0),
      end: queryWindow.endSecondsSinceEpoch.toFixed(0),
    },
  });

  return queryResponseDataSchema.parse(response.data).data;
};

const convertPrometheusHistogramToNewRelicGaugeMetrics = (
  metricName: string,
  quantile: number,
  metricData: MetricData
): NewRelicMetric[][] =>
  metricData.result.map((prometheusMetric) =>
    prometheusMetric.values
      // need to filter out NaN's since NewRelic will reject them anyways
      .filter((v) => v[1] !== "NaN")
      .map((v) => ({
        name:
          NEWRELIC_METRIC_PREFIX +
          metricName.split("_bucket")[0] +
          "_P" +
          quantile * 100,
        type: "gauge",
        value: parseFloat(v[1]),
        timestamp: v[0],
        attributes: _.omit(prometheusMetric.metric, "__rollup__"),
      }))
  );

const main = async () => {
  const { countMetricNames, histogramMetricNames } = await getMetricNames();

  console.log({
    level: "info",
    message: "Polling metrics",
    countMetricNames,
    histogramMetricNames,
  });

  while (true) {
    const queryWindow = generateQueryWindow();

    console.log({
      level: "info",
      message: "Collecting metrics from temporal cloud.",
      startDate: new Date(
        queryWindow.startSecondsSinceEpoch * 1000
      ).toISOString(),
      endDate: new Date(queryWindow.endSecondsSinceEpoch * 1000).toISOString(),
    });

    const countMetrics = (
      await Promise.all(
        countMetricNames.map(async (metricName) =>
          convertPrometheusCountToNewRelicCountMetrics(
            metricName,
            await queryPrometheusCount(metricName, generateQueryWindow())
          )
        )
      )
    ).flat(4);

    const gaugeMetrics = (
      await Promise.all(
        histogramMetricNames.map(async (metricName) =>
          Promise.all(
            HISTOGRAM_QUANTILES.map(async (quantile) =>
              convertPrometheusHistogramToNewRelicGaugeMetrics(
                metricName,
                quantile,
                await queryPrometheusHistogram(
                  metricName,
                  quantile,
                  generateQueryWindow()
                )
              )
            )
          )
        )
      )
    ).flat(4);

    console.log({ level: "info", message: "Submitting metrics to Newrelic" });

    const res = await newRelicMetricsApi.post("/metric/v1", [
      { metrics: [...countMetrics, ...gaugeMetrics] },
    ]);

    console.log(`NewRelic Result: ${res.status}`);

    console.log({ level: "info", message: "Pausing for 20s" });
    await setTimeoutAsync(20 * 1000);
  }
};

main().catch((error) => {
  console.log({
    level: "error",
    message: "Error in main loop. Shutting down metrics service.",
    error,
  });
  process.exit(1);
});

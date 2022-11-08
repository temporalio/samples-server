#!/usr/bin/env python3
"""
promql-to-dd.py - Import counters and histograms from prometheus api endpoint into datadog

While this demonstrates how to import prometheus api data using the datadog metrics API,
there is a lot of room for improvement in terms of efficency and error handling.

To view this data in DataDog Metrics:
* use sum and as_rate for rate metrics
* use sum for gauge metrics

This script depends on these packages:
https://pypi.org/project/tenacity/
https://pypi.org/project/requests/
https://pypi.org/project/datadog-api-client/

For documentation on the prometheus API used see:
    https://prometheus.io/docs/prometheus/latest/querying/api/#range-queries

For documentation on PromQL see:
    https://prometheus.io/docs/prometheus/latest/querying/basics/

For documentation on DataDog Metrics API see:
    https://docs.datadoghq.com/api/latest/metrics/#submit-metrics
"""
import os
import time
import argparse
from itertools import zip_longest
from collections.abc import Iterable, Mapping
from datetime import datetime, timedelta

# quick and dirty
try:
    import requests
    import tenacity
    import datadog_api_client
except ImportError:
    print(
        "To run this script, please first install dependencies.\n"
        "Preferably in a virtual environment, run these commands:\n\n"
        "python3 -m pip install requests\n"
        "python3 -m pip install tenacity\n"
        "python3 -m pip install datadog_api_client\n"
    )
    exit(2)


from datadog_api_client import ApiClient, Configuration
from datadog_api_client.v2.api.metrics_api import MetricsApi
from datadog_api_client.v2.model.metric_content_encoding import MetricContentEncoding
from datadog_api_client.v2.model.metric_intake_type import MetricIntakeType
from datadog_api_client.v2.model.metric_payload import MetricPayload
from datadog_api_client.v2.model.metric_point import MetricPoint
from datadog_api_client.v2.model.metric_series import MetricSeries

# initial range to query on startup
initial_window_minutes = 240
# subsequent time windows to query
window_minutes = 10
# seconds to sleep before inserting more data
sleep_seconds = 20
# step seconds between samples
step_seconds = 60
# histogram quantiles to pre-compute for datadog
quantiles = (0.5, 0.75, 0.9, 0.95, 0.99)
# format and step used for all queries
base_params = {"format": "json", "step": step_seconds}
# retry profile for promql and datadog api calls
retry = tenacity.retry(
    wait=tenacity.wait_random_exponential(multiplier=2, min=5, max=60),
    stop=tenacity.stop_after_attempt(5),
    before_sleep=lambda s: print(
        f"{datetime.now()}: Retry {s.fn.__name__} for {type(s.outcome.exception()).__name__}: {s.outcome.exception()}"
    ),
)


def histogram_promql(metric_name: str, quantile: float):
    """Generate promql query for a histogram quantile for a bucketed metric"""
    return f"histogram_quantile({quantile}, sum(rate({metric_name}[1m])) by (temporal_account,temporal_namespace,operation,le))"


def counter_promql(metric_name: str):
    """Generate promql query for a counter rate over 1m (minimum for 30s samples)"""
    return f"rate({metric_name}[1m])"


@retry
def retryable_promql_query_results(params: Mapping):
    """Pull result out of promql query result and retry on failure"""
    response = requests.get(query_endpoint, params=params, cert=cert)
    return response.json()["data"]["result"]


@retry
def retryable_submit_metrics(datadog_api: MetricsApi, body: MetricPayload):
    """Submit gzipped metrics to datadog and retry on failure"""
    response = datadog_api.submit_metrics(
        body=body,
        content_encoding=MetricContentEncoding("gzip"),
    )
    return response


def submit_datadog_series(datadog_api: MetricsApi, series: Iterable):
    print(f"{datetime.now()}: Ingesting {len(series)} series into DataDog")
    # submit 200 series at a time as a naieve optimization
    # this could be tuned to submit upto 5MB of metrics
    # data compressed to a size of upto 512KB
    non_empty_responses = []
    for group in zip_longest(*([iter(series)] * 200)):
        mini_series = list(filter(None, group))
        datadog_response = retryable_submit_metrics(
            datadog_api,
            MetricPayload(series=mini_series),
        )
        # successful response body is an object with errors as an empty list
        if str(datadog_response) != "{'errors': []}":
            non_empty_responses.append(datadog_response)

    # currently nothing is done with these if there are any
    # none were seen during development of this script
    return non_empty_responses


def create_datadog_rate_series(metric_name: str, result: Mapping):
    """Upload a MetricSeries with MetricPoints as rates to datadog for a result from promql"""
    series = MetricSeries(
        # match interval to step size
        interval=step_seconds,
        # note that this metric is based on a 1m rate query
        metric=metric_name + "_rate1m",
        # type 2 indicates a rate
        type=MetricIntakeType(2),
        # each value from the prometheus api is a (timestamp, rate) pair
        points=[
            MetricPoint(timestamp=sample[0], value=float(sample[1]))
            for sample in result["values"]
            # don't insert anything if value is NaN
            if sample[1] != "NaN"
        ],
        # datadog expects ":" separated key value pairs as tags
        tags=[
            # datadog tags have a maximum length of 200 unicode characters
            # some characters may be modified automatically by the api
            ":".join(pair)[:200]
            for pair in result["metric"].items()
            # __rollup__ is always true so we ignore it
            if pair[0] != "__rollup__"
        ],
    )
    return series


def create_datadog_gauge_quantile(metric_name: str, quantile: float, result: Mapping):
    """Upload a MetricSeries with MetricPoints as gauges to datadog for a result from promql"""
    series = MetricSeries(
        # name metrics with quantile for clarity
        metric=metric_name.removesuffix("_bucket") + f"_P{str(int(quantile * 100))}",
        # type 3 indicates a gauge which does not take an interval
        type=MetricIntakeType(3),
        points=[
            MetricPoint(timestamp=sample[0], value=float(sample[1]))
            for sample in result["values"]
            # the way histogram queries work can result in NaN responses - skip
            if sample[1] != "NaN"
        ],
        tags=[":".join(pair)[:200] for pair in result["metric"].items()],
    )
    return series


def gather_counters(names: Iterable, start: datetime, end: datetime):
    """Gather Prometheus counters as rates into DataDog Metrics"""
    print(f"{datetime.now()}: Gathering {start} to {end} counters")
    params = base_params | {"start": start.timestamp(), "end": end.timestamp()}
    # loop over counter metrics, query prometheus api, and upload to datadog
    series = []
    for name in names:
        # setup prometheus query for a counter for this metric
        params |= {"query": counter_promql(name)}
        # prometheus api returns a list of (timestamp, rate)
        # for each resulting set of unique tags which make up a series
        for result in retryable_promql_query_results(params):
            series.append(create_datadog_rate_series(name, result))

    return series


def gather_histograms(names: Iterable, start: datetime, end: datetime):
    """Gather Prometheus histograms at quantiles as gauges into DataDog Metrics"""
    print(f"{datetime.now()}: Gathering {start} to {end} histograms")
    params = base_params | {"start": start.timestamp(), "end": end.timestamp()}
    # loop over histogram metrics, and, for each desired quantile query prometheus api and upload to datadog
    series = []
    for name in names:
        for quantile in quantiles:
            # setup prometheus query for a histogram for this metric at a quantile
            params |= {"query": histogram_promql(name, quantile)}
            for result in retryable_promql_query_results(params):
                series.append(create_datadog_gauge_quantile(name, quantile, result))

    return series


def env_or_arg(key):
    value = os.environ.get(key)
    return {"default": value} if value else {"required": True}


def configure_environment():
    global query_endpoint, labels_endpoint, cert, base_params, configuration
    parser = argparse.ArgumentParser(
        description="Export Prometheus API metrics and import into DataDog"
    )
    parser.add_argument("--temporal-account", **env_or_arg("TEMPORAL_ACCOUNT"))
    parser.add_argument("--metrics-client-cert", **env_or_arg("METRICS_CLIENT_CERT"))
    parser.add_argument("--metrics-client-key", **env_or_arg("METRICS_CLIENT_KEY"))
    parser.add_argument("--dd-site", **env_or_arg("DD_SITE"))
    parser.add_argument("--dd-api-key", **env_or_arg("DD_API_KEY"))
    args = parser.parse_args()
    # Prometheus API query_range is used for gathering rates and histograms
    query_endpoint = (
        f"https://{args.temporal_account}.tmprl.cloud/prometheus/api/v1/query_range"
    )
    # Prometheus API label endpoint is used to get names of all metrics available
    labels_endpoint = f"https://{args.temporal_account}.tmprl.cloud/prometheus/api/v1/label/__name__/values"
    # Temporal Observability mTLS Client Cert and Key used for api requests
    cert = (args.metrics_client_cert, args.metrics_client_key)
    # datadog api configuration from environment
    os.environ["DD_SITE"] = args.dd_site
    os.environ["DD_API_KEY"] = args.dd_api_key
    configuration = Configuration()


def main():
    print(f"{datetime.now()}: Starting Temporal Metrics DataDog Ingestion")

    # get a list of histogram and counter metrics from temporal prometheus api endpoint
    metric_names = [n for n in requests.get(labels_endpoint, cert=cert).json()["data"]]
    histogram_names = [n for n in metric_names if n.endswith("_bucket")]
    counter_names = [
        n
        for n in metric_names
        # exclude _bucket _sum _count for histogram metrics
        if n.rpartition("_")[0] + "_bucket" not in histogram_names
    ]

    # set initial range and a window to loop over to cover missing values
    end = datetime.now()
    start = end - timedelta(minutes=initial_window_minutes)
    window = timedelta(minutes=window_minutes)

    with ApiClient(configuration) as api_client:
        # create datadog metrics api instance
        datadog_api = MetricsApi(api_client)
        # loop forever
        while True:
            # align start and end on step boundaries for consistent sample time
            start = datetime.fromtimestamp(
                int(start.timestamp() / step_seconds - 1) * step_seconds
            )
            end = datetime.fromtimestamp(
                int(end.timestamp() / step_seconds + 1) * step_seconds
            )
            # import to datadog from prometheus
            counter_series = gather_counters(counter_names, start, end)
            histogram_series = gather_histograms(histogram_names, start, end)
            submit_datadog_series(datadog_api, counter_series + histogram_series)
            # temporal prometheus endpoint only updates every 30-60 seconds
            # so sleep a little while before polling again
            print(f"{datetime.now()} Pausing for {sleep_seconds}s")
            time.sleep(sleep_seconds)
            # look back over a window of time that overlaps the previous query
            start = end - window
            # query up to the current time
            end = datetime.now()


if __name__ == "__main__":
    configure_environment()
    try:
        main()
    except KeyboardInterrupt:
        exit(130)

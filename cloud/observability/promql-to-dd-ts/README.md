PromQL to Datadog Typescript
=================

The goal of this TS implementation is to demonstrate the minimum work necessary to read recently generated metrics from a Temporal Cloud account using the Prometheus API and import them into DataDog while handling some common edge and error cases.

Destination metrics could be modified to match a DataDog environment's naming conventions and metrics types by modifying the DataDog API calls as needed.

**These examples are provided as-is, without support. They are intended as reference material only.**

# Example dashboard

Download the dashboard from [here](examples/datadog_dashboard.json).

# Running locally

## Prerequisites

* Node v20.0.0+
* A Datadog API key exported as `DD_API_KEY` in your shell
* You have [configured your observability endpoint](https://docs.temporal.io/cloud/how-to-monitor-temporal-cloud-metrics)
* You have a datadog account and access to an api key
* You have client certs signed by the CA uploaded to temporal


Usage
-----

Here's an example, of setting up an environment and running the script:

```bash

# install dependencies
npm install

# set up required environment var
export TEMPORAL_ACCOUNT=acctcode
export METRICS_CLIENT_CERT=/path/to/cert
export METRICS_CLIENT_KEY=/path/to/key
export DD_SITE=datadog.api.endpoint.domain
export DD_API_KEY=123apikey321

# run the script with nodemon
 npm run start.watch

```


Should output:

```
[nodemon] starting `ts-node src/index.ts src/worker.ts`
{
  level: 'info',
  message: 'Polling metrics',
  countMetricNames: [
    'temporal_cloud_v0_frontend_service_error_count',
    ...
  ],
  histogramMetricNames: [ 
  'temporal_cloud_v0_service_latency_bucket',
  ...
  ]
}
{
  level: 'info',
  message: 'Collecting metrics from temporal cloud.',
  startDate: '2023-08-11T15:01:00.000Z',
  endDate: '2023-08-11T15:04:00.000Z'
}
{ level: 'info', message: 'Submitting metrics to Datadog' }
{ level: 'info', message: 'Pausing for 20s' }

```

## License
MIT License, please see [LICENSE](LICENSE) for details.

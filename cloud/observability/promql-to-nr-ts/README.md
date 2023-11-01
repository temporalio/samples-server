PromQL to Newrelic
=================

The goal of this example script is to demonstrate the minimum work necessary to read recently generated metrics from a Temporal Cloud account using the Prometheus API and import them into NewRelic while handling some common edge and error cases. This TypeScript example was adapted from the DataDog TypeScript observability example.

Destination metrics could be modified to match a NewRelic environment's naming conventions and metrics types by modifying the NewRelic API calls as needed.

**These examples are provided as-is, without support. They are intended as reference material only.**

Usage
-----
The script can be run using `ts-node` (for development/local testing), or compiled to JS and run with `node`.

```
usage: TEMPORAL_OBSERVABILITY_CERT=<b64-cert> TEMPORAL_OBSERVABILITY_KEY=<b64-key> TEMPORAL_CLOUD_BASE_URL=<your-temporal-cloud-base-url> NEW_RELIC_API_KEY=<key> ts-node promql-to-nr.ts
```

You can optionally pass a `TEST_METRIC_PREFIX=testing_` environment variable to prefix metric names for local testing.
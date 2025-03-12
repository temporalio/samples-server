# promql-to-scrape

This basic application is meant to provide an example for how one could use the Temporal Cloud Observability endpoint to expose a typical Prometheus `/metrics` endpoint.

**This example is provided as-is, without support. It is intended as reference material only.**

## How to Use

Grab your client cert and key and place them at `client.crt`, `tls.key`, and your Temporal Cloud account number that has the observability endpoint enabled.

```
go mod tidy
go build -o promql-to-scrape cmd/promql-to-scrape/main.go
./promql-to-scrape -client-cert client.crt -client-
key tls.key -prom-endpoint https://<account>.tmprl.cloud/prometheus --config-file examples/config.yaml --debug
~~~
time=2023-11-16T17:43:20.260-06:00 level=DEBUG msg="successful metric retrieval" time=3.529039083s
```

This means you can now hit http://localhost:9001/metrics on your machine and see your metrics.

### Important Usability Information

**Very Important:** This application will show data _delayed by one minute_. This is done in an attempt to smooth out some aggregation delay. However, you may encounter issues with data appearing missing if you **use a rate interval < 2m**.

**Important:** When you scrape this endpoint, you should do so with a scrape interval **<= the rate interval of the queries in your config file, and at least 1m**.

## Deployment

Some example Kubernetes manifests are provided in the `/examples` directory. Filling in your certificates and account should get you going pretty quickly.

## Generating Config

There is a second binary you can build that can help you build a default configuration of queries to scrape and export. 

```
go build -o genconfig cmd/genconfig/main.go
./genconfig -client-cert client.crt -client-key tls.key -prom-endpoint https://<account>.tmprl.cloud/prometheus 
...
```

This will generate an example config at `config.yaml` that you may use. It looks for all the existing metrics and generates a reasonable query for you to export.
- For counters, a `rate(counter[1m])`
- For gauges, it simply queries for `gauge`
- For histograms, it does a p99 aggregated by `temporal_namespace` and `operation`. `histogram_quantile(0.99, sum(rate(metric[1m])) by (le, operation, temporal_namespace)`

Modify at your own risk. You may find you'd like to add a global latency across all namespaces for instance. You can add those queries to your config file. 

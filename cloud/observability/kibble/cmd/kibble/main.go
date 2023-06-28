package main

import (
	"flag"
	"log"
	"os"
	"time"

	"github.com/temporalio/kibble/datadog"
	"github.com/temporalio/kibble/prometheus"
	"github.com/temporalio/kibble/worker"
)

func main() {
	set := flag.NewFlagSet("app", flag.ExitOnError)
	promURL := set.String("prom-endpoint", "", "Prometheus API endpoint for the server")
	serverRootCACert := set.String("server-root-ca-cert", "", "Optional path to root server CA cert")
	clientCert := set.String("client-cert", "", "Required path to client cert")
	clientKey := set.String("client-key", "", "Required path to client key")
	serverName := set.String("server-name", "", "Server name to use for verifying the server's certificate")
	insecureSkipVerify := set.Bool("insecure-skip-verify", false, "Skip verification of the server's certificate and host name")
	metricPrefix := set.String("metric-prefix", "temporal_cloud_v0", "The metric prefix to query")
	stepDuration := set.Int("step-duration-seconds", 60, "The step between metrics")
	queryInterval := set.Int("query-interval-seconds", 600, "Interval between each Prometheus query")

	if err := set.Parse(os.Args[1:]); err != nil {
		log.Fatalf("failed parsing args: %s", err)
	} else if *clientCert == "" || *clientKey == "" {
		log.Fatalf("-client-cert and -client-key are required")
	}

	datadogClient := datadog.NewAPIClient()

	prometheusClient, err := prometheus.NewAPIClient(
		prometheus.Config{
			TargetHost:         *promURL,
			ServerRootCACert:   *serverRootCACert,
			ClientCert:         *clientCert,
			ClientKey:          *clientKey,
			ServerName:         *serverName,
			InsecureSkipVerify: *insecureSkipVerify,
		},
	)
	if err != nil {
		log.Fatalf("Failed to create Prometheus client: %s", err)
	}

	worker := worker.Worker{
		Querier:       prometheusClient,
		Submitter:     datadogClient,
		MetricPrefix:  *metricPrefix,
		StepDuration:  time.Duration(*stepDuration) * time.Second,
		QueryInterval: time.Duration(*queryInterval) * time.Second,
		Quantiles:     []float64{0.5, 0.9, 0.95, 0.99},
	}

	worker.Run()
}

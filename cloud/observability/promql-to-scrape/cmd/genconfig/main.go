package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"sort"

	"github.com/temporalio/samples-server/cloud/observability/promql-to-scrape/internal"

	"gopkg.in/yaml.v3"
)

func main() {
	set := flag.NewFlagSet("app", flag.ExitOnError)
	promURL := set.String("prom-endpoint", "", "Required Prometheus API endpoint for the server eg. https://<account>.tmprl.cloud/prometheus")
	serverRootCACert := set.String("server-root-ca-cert", "", "Optional path to root server CA cert")
	clientCert := set.String("client-cert", "", "Required path to client cert")
	clientKey := set.String("client-key", "", "Required path to client key")
	serverName := set.String("server-name", "", "Optional server name to use for verifying the server's certificate")
	insecureSkipVerify := set.Bool("insecure-skip-verify", false, "Skip verification of the server's certificate and host name")

	if err := set.Parse(os.Args[1:]); err != nil {
		log.Fatalf("failed parsing args: %s", err)
	} else if *clientCert == "" || *clientKey == "" {
		log.Fatalf("-client-cert and -client-key are required")
	}

	client, err := internal.NewAPIClient(
		internal.APIConfig{
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

	counters, gauges, histograms, err := client.ListMetrics("temporal_cloud_v0")
	if err != nil {
		log.Fatalf("Failed to pull metric names: %s", err)
	}
	fmt.Println("counters: ", counters, "\n")
	fmt.Println("gauges: ", gauges, "\n")
	fmt.Println("histograms: ", histograms, "\n")

	conf := internal.Config{}

	for _, counter := range counters {
		conf.Metrics = append(conf.Metrics, internal.Metric{
			MetricName: fmt.Sprintf("%s:rate1m", counter),
			Query:      fmt.Sprintf("rate(%s[1m])", counter),
		})
	}
	for _, gauge := range gauges {
		conf.Metrics = append(conf.Metrics, internal.Metric{
			MetricName: gauge,
			Query:      gauge,
		})
	}
	for _, histogram := range histograms {
		conf.Metrics = append(conf.Metrics, internal.Metric{
			MetricName: fmt.Sprintf("%s:histogram_quantile_p99_1m", histogram),
			Query:      fmt.Sprintf("histogram_quantile(0.99, sum(rate(%s[1m])) by (le, operation, temporal_namespace))", histogram),
		})
	}

	sort.Sort(internal.ByMetricName(conf.Metrics))

	yamlData, err := yaml.Marshal(&conf)
	if err != nil {
		log.Fatalf("error marshalling yaml: %v", err)
	}

	err = os.WriteFile("config.yaml", yamlData, 0644)
	if err != nil {
		log.Fatalf("error: %v", err)
	}
}

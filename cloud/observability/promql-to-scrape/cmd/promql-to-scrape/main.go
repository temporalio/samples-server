package main

import (
	"flag"
	"log"
	"os"

	"github.com/temporalio/samples-server/cloud/observability/promql-to-scrape/internal"

	"golang.org/x/exp/slog"
)

func main() {
	set := flag.NewFlagSet("promql-to-scrape", flag.ExitOnError)
	promURL := set.String("prom-endpoint", "", "Required Prometheus API endpoint for the server eg. https://<account>.tmprl.cloud/prometheus")
	configFile := set.String("config-file", "", "Config file for promql-to-scrape")
	serverRootCACert := set.String("server-root-ca-cert", "", "Optional path to root server CA cert")
	clientCert := set.String("client-cert", "", "Required path to client cert")
	clientKey := set.String("client-key", "", "Required path to client key")
	serverName := set.String("server-name", "", "Optional server name to use for verifying the server's certificate")
	insecureSkipVerify := set.Bool("insecure-skip-verify", false, "Skip verification of the server's certificate and host name")
	serverAddr := set.String("bind", "0.0.0.0:9001", "address:port to expose the metrics server on")
	debugLogging := set.Bool("debug", false, "Toggle debug logging")

	if err := set.Parse(os.Args[1:]); err != nil {
		log.Fatalf("failed parsing args: %v", err)
	} else if *clientCert == "" || *clientKey == "" || *configFile == "" || *promURL == "" {
		log.Fatalf("-client-cert, -client-key, -config-file, -prom-endpoint are required")
	}

	logLevel := slog.LevelInfo
	if *debugLogging {
		logLevel = slog.LevelDebug
	}
	h := slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: logLevel})
	slog.SetDefault(slog.New(h))

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
		log.Fatalf("failed to create Prometheus client: %v", err)
	}

	conf, err := internal.LoadConfig(*configFile)
	if err != nil {
		log.Fatalf("failed to load config file: %v", err)
	}

	s := internal.NewPromToScrapeServer(client, conf, *serverAddr)
	s.Start()
}

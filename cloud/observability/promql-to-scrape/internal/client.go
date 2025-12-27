package internal

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	promapi "github.com/prometheus/client_golang/api/prometheus/v1"
	"github.com/prometheus/common/model"
)

type (
	Querier interface {
		ListMetrics(metricPrefix string) ([]string, []string, []string, error)
		QueryMetricsInstant(promql string) (model.Matrix, error)
	}

	APIClient struct {
		promapi.API
	}
)

type APIConfig struct {
	TargetHost         string
	ServerRootCACert   string
	ClientCert         string
	ClientKey          string
	ServerName         string
	InsecureSkipVerify bool
}

func NewAPIClient(cfg APIConfig) (*APIClient, error) {
	tlsCfg, err := BuildTLSConfig(
		cfg.ClientCert,
		cfg.ClientKey,
		cfg.ServerRootCACert,
		cfg.ServerName,
		cfg.InsecureSkipVerify,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to build tls config %w", err)
	}

	httpClient := &http.Client{
		Transport: &http.Transport{TLSClientConfig: tlsCfg},
	}

	client, err := NewHttpClient(cfg.TargetHost, httpClient)
	if err != nil {
		return nil, fmt.Errorf("failed to build tls client %w", err)
	}

	return &APIClient{promapi.NewAPI(client)}, nil
}

func (c *APIClient) ListMetrics(metricPrefix string) ([]string, []string, []string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	values, _, err := c.LabelValues(ctx, "__name__", nil, time.Time{}, time.Time{})
	if err != nil {
		return nil, nil, nil, fmt.Errorf("failed to fetch Prometheus metric names: %w", err)
	}
	counts := []string{}
	gauges := []string{}
	histograms := []string{}
	for _, v := range values {
		if !strings.HasPrefix(string(v), metricPrefix) {
			continue
		}
		t := getMetricType(string(v))
		if t == metricTypeHistogram {
			histograms = append(histograms, string(v))
		} else if t == metricTypeCounter {
			counts = append(counts, string(v))
		} else {
			gauges = append(gauges, string(v))
		}
	}
	return counts, gauges, histograms, nil
}

func (c *APIClient) QueryMetricsInstant(promql string) (model.Vector, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	result, warnings, err := c.API.Query(ctx, promql, time.Now().Add(-60*time.Second), promapi.WithTimeout(10*time.Second))
	if err != nil {
		return nil, fmt.Errorf("failed to query Temporal Cloud: %w", err)
	}
	if len(warnings) > 0 {
		log.Printf("warning while querying Temporal Cloud: %v\n", warnings)
	}
	promVector, ok := result.(model.Vector)
	if !ok {
		log.Printf("unexpected type %T returned for bucket metric", result)
	}
	return promVector, nil
}

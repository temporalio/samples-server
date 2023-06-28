package prometheus

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
		ListMetrics(metricPrefix string) ([]string, []string, error)
		QueryMetrics(promql string, queryRange promapi.Range) (model.Matrix, error)
	}

	APIClient struct {
		promapi.API
	}
)

type Config struct {
	TargetHost         string
	ServerRootCACert   string
	ClientCert         string
	ClientKey          string
	ServerName         string
	InsecureSkipVerify bool
}

func NewAPIClient(cfg Config) (*APIClient, error) {
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

func (c *APIClient) ListMetrics(metricPrefix string) ([]string, []string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	values, _, err := c.LabelValues(ctx, "__name__", nil, time.Time{}, time.Time{})
	if err != nil {
		return nil, nil, fmt.Errorf("failed to fetch Prometheus metric names: %w", err)
	}
	buckets := []string{}
	counts := []string{}
	for _, v := range values {
		if !strings.HasPrefix(string(v), metricPrefix) {
			continue
		}
		if strings.HasSuffix(string(v), "_bucket") {
			buckets = append(buckets, string(v))
		} else {
			counts = append(counts, string(v))
		}
	}
	return buckets, counts, nil
}

func (c *APIClient) QueryMetrics(promql string, queryRange promapi.Range) (model.Matrix, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	result, warnings, err := c.API.QueryRange(ctx, promql, queryRange, promapi.WithTimeout(10*time.Second))
	if err != nil {
		return nil, fmt.Errorf("failed to query Prometheus: %w", err)
	}
	if len(warnings) > 0 {
		log.Printf("warning while querying Prometheus range: %v\n", warnings)
	}
	promMatrix, ok := result.(model.Matrix)
	if !ok {
		log.Printf("unexpected type %T returned for bucket metric", result)
	}
	return promMatrix, nil
}

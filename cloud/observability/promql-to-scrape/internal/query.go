package internal

import (
	"fmt"

	"github.com/prometheus/common/model"
)

type Data map[string][]*model.Sample

func QueryMetrics(conf *Config, client *APIClient) (Data, error) {
	// https://pkg.go.dev/github.com/prometheus/common/model#Sample
	queriedMetrics := map[string][]*model.Sample{}

	for _, metric := range conf.Metrics {
		result, err := client.QueryMetricsInstant(metric.Query)
		if err != nil {
			return nil, fmt.Errorf("failed to query for %s: %v", metric.MetricName, err)
		}

		queriedMetrics[metric.MetricName] = []*model.Sample(result)
	}

	return Data(queriedMetrics), nil
}

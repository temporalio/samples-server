package worker

import (
	"math"
	"testing"
	"time"

	"github.com/DataDog/datadog-api-client-go/v2/api/datadogV2"
	"github.com/prometheus/common/model"
	"github.com/stretchr/testify/assert"
)

func Ptr[T any](v T) *T {
	return &v
}

func TestPromMatrixToDatadogSeries(t *testing.T) {
	testCases := []struct {
		name       string
		metricName string
		quantile   float64
		matrix     model.Matrix
		wantSeries []datadogV2.MetricSeries
	}{
		{
			name:       "fully populated",
			metricName: "latency_bucket",
			quantile:   0.95,
			matrix: model.Matrix{
				&model.SampleStream{
					Metric: model.Metric{"operation": "StartWorkflowExecution", "namespace": "disneyland"},
					Values: []model.SamplePair{
						{
							Timestamp: model.Time(time.Date(2009, time.November, 10, 23, 0, 0, 0, time.UTC).Unix()),
							Value:     1.0,
						},
						{
							Timestamp: model.Time(time.Date(2009, time.November, 10, 23, 0, 0, 0, time.UTC).Unix()),
							Value:     2.0,
						},
					},
				},
			},
			wantSeries: []datadogV2.MetricSeries{
				{
					Metric: "latency_P95",
					Type:   datadogV2.METRICINTAKETYPE_GAUGE.Ptr(),
					Points: []datadogV2.MetricPoint{
						{Timestamp: Ptr(int64(1257894)), Value: Ptr(float64(1.0))},
						{Timestamp: Ptr(int64(1257894)), Value: Ptr(float64(2.0))},
					},
					Resources: []datadogV2.MetricResource{
						{Type: Ptr("operation"), Name: Ptr("StartWorkflowExecution")},
						{Type: Ptr("namespace"), Name: Ptr("disneyland")},
					},
				},
			},
		},
		{
			name:       "contains NaN vlaues",
			metricName: "latency",
			quantile:   0.5,
			matrix: model.Matrix{
				&model.SampleStream{
					Metric: model.Metric{"operation": "StartWorkflowExecution", "namespace": "disneyland"},
					Values: []model.SamplePair{
						{
							Timestamp: model.Time(time.Date(2009, time.November, 10, 23, 0, 0, 0, time.UTC).Unix()),
							Value:     model.SampleValue(math.NaN()),
						},
						{
							Timestamp: model.Time(time.Date(2009, time.November, 10, 23, 0, 0, 0, time.UTC).Unix()),
							Value:     2.0,
						},
					},
				},
			},
			wantSeries: []datadogV2.MetricSeries{
				{
					Metric: "latency_P50",
					Type:   datadogV2.METRICINTAKETYPE_GAUGE.Ptr(),
					Points: []datadogV2.MetricPoint{
						{Timestamp: Ptr(int64(1257894)), Value: Ptr(float64(0.0))},
						{Timestamp: Ptr(int64(1257894)), Value: Ptr(float64(2.0))},
					},
					Resources: []datadogV2.MetricResource{
						{Type: Ptr("operation"), Name: Ptr("StartWorkflowExecution")},
						{Type: Ptr("namespace"), Name: Ptr("disneyland")},
					},
				},
			},
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			gotSeries := PromHistogramToDatadogGauge(tc.metricName, tc.quantile, tc.matrix)
			for i := range gotSeries {
				assert.Equal(t, gotSeries[i].Metric, tc.wantSeries[i].Metric)
				assert.Equal(t, gotSeries[i].Type, tc.wantSeries[i].Type)
				assert.ElementsMatch(t, gotSeries[i].Points, tc.wantSeries[i].Points)
				assert.ElementsMatch(t, gotSeries[i].Resources, tc.wantSeries[i].Resources)
			}
		})
	}
}

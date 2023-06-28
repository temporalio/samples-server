package worker

import (
	"fmt"
	"math"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/DataDog/datadog-api-client-go/v2/api/datadogV2"
	"github.com/prometheus/common/model"
)

func interruptCh() <-chan interface{} {
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)

	ret := make(chan interface{}, 1)
	go func() {
		s := <-c
		ret <- s
		close(ret)
	}()

	return ret
}

func PromHistogramToDatadogGauge(name string, quantile float64, matrix model.Matrix) []datadogV2.MetricSeries {
	name = strings.TrimSuffix(name, "_bucket") + fmt.Sprintf("_P%2.0f", quantile*100)
	metricType := datadogV2.METRICINTAKETYPE_GAUGE
	return matrixToSeries(name, metricType, matrix)
}

func PromCountToDatadogCount(name string, matrix model.Matrix) []datadogV2.MetricSeries {
	name = strings.TrimSuffix(name, "_count") + "_rate1m"
	metricType := datadogV2.METRICINTAKETYPE_COUNT
	return matrixToSeries(name, metricType, matrix)
}

func matrixToSeries(name string, metricType datadogV2.MetricIntakeType, matrix model.Matrix) []datadogV2.MetricSeries {
	series := make([]datadogV2.MetricSeries, len(matrix))
	for i, stream := range matrix {
		labels := []datadogV2.MetricResource{}
		for k, v := range stream.Metric {
			name := string(k)
			if name == "__rollup__" {
				continue
			}
			value := string(v)
			labels = append(labels, datadogV2.MetricResource{Type: &name, Name: &value})
		}

		points := []datadogV2.MetricPoint{}
		for _, valuePair := range stream.Values {
			value := float64(valuePair.Value)
			if math.IsNaN(value) {
				value = 0.0
			}
			timestamp := valuePair.Timestamp.Unix()
			point := datadogV2.MetricPoint{
				Timestamp: &timestamp,
				Value:     &value,
			}
			points = append(points, point)
		}

		series[i] = datadogV2.MetricSeries{
			Metric:    name,
			Type:      metricType.Ptr(),
			Points:    points,
			Resources: labels,
		}
	}
	return series
}

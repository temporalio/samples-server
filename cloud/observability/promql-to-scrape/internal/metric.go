package internal

import "strings"

const (
	metricTypeHistogram = "histogram"
	metricTypeCounter   = "count"
	metricTypeGauge     = "gauge"
)

func getMetricType(v string) string {
	if strings.HasSuffix(v, "_bucket") {
		return metricTypeHistogram
	} else if strings.HasSuffix(v, "_count") || strings.HasSuffix(v, "_sum") {
		return metricTypeCounter
	}
	return metricTypeGauge
}

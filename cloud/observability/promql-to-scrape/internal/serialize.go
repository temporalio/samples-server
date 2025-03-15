package internal

import (
	"math"
	"strconv"
	"strings"
	"sync"

	"github.com/prometheus/common/model"
)

// Everything below this point is pretty shamelessly stolen and slightly modified from
// https://github.com/kubernetes/kube-state-metrics/blob/main/pkg/metric/metric.go

var (
	escapeWithDoubleQuote = strings.NewReplacer("\\", `\\`, "\n", `\n`, "\"", `\"`)
	initialNumBufSize     = 24
	numBufPool            = sync.Pool{
		New: func() interface{} {
			b := make([]byte, 0, initialNumBufSize)
			return &b
		},
	}
)

// SamplesToString turns our queried metrics map into something compatible with the Prometheus exposition format
func SamplesToString(queriedMetrics map[string][]*model.Sample) string {
	var sb strings.Builder

	for metricName, samples := range queriedMetrics {
		nameWithoutSuffix := trimSuffixes(metricName, []string{"_count", "_sum", "_bucket"})
		sb.WriteString("# HELP ")
		sb.WriteString(nameWithoutSuffix)
		sb.WriteByte(' ')
		sb.WriteString("https://docs.temporal.io/cloud/metrics#available-metrics")
		sb.WriteByte('\n')

		sb.WriteString("# TYPE ")
		sb.WriteString(nameWithoutSuffix)
		sb.WriteByte(' ')
		sb.WriteString(getMetricType(metricName))
		sb.WriteByte('\n')

		for _, s := range samples {
			sb.WriteString(metricName)

			// write labels
			var separator byte = '{'
			if len(s.Metric) == 1 {
				sb.WriteByte(separator)
			}
			for k, v := range model.LabelSet(s.Metric) {
				name := string(k)
				if name == "__name__" || name == "__rollup__" || name == "temporal_service_type" {
					continue
				}
				sb.WriteByte(separator)
				sb.WriteString(string(k))
				sb.WriteString("=\"")
				escapeString(&sb, string(v))
				sb.WriteByte('"')
				separator = ','
			}
			sb.WriteByte('}')

			// write value
			sb.WriteByte(' ')
			writeFloat(&sb, float64(s.Value))

			// write timestamp in milliseconds since epoch
			// do we want timestamp? i doubt it
			//sb.WriteByte(' ')
			//writeInt(&sb, s.Timestamp.Unix()*1000) //nolint

			// end
			sb.WriteByte('\n')
		}
	}

	return sb.String()
}

// escapeString replaces '\' by '\\', new line character by '\n', and '"' by
// '\"'.
// Taken from github.com/prometheus/common/expfmt/text_create.go.
func escapeString(m *strings.Builder, v string) {
	escapeWithDoubleQuote.WriteString(m, v) //nolint
}

// writeFloat is equivalent to fmt.Fprint with a float64 argument but hardcodes
// a few common cases for increased efficiency. For non-hardcoded cases, it uses
// strconv.AppendFloat to avoid allocations, similar to writeInt.
// Taken from github.com/prometheus/common/expfmt/text_create.go.
func writeFloat(w *strings.Builder, f float64) {
	switch {
	case f == 1:
		w.WriteByte('1')
	case f == 0:
		w.WriteByte('0')
	case f == -1:
		w.WriteString("-1")
	case math.IsNaN(f):
		w.WriteString("NaN")
	case math.IsInf(f, +1):
		w.WriteString("+Inf")
	case math.IsInf(f, -1):
		w.WriteString("-Inf")
	default:
		bp := numBufPool.Get().(*[]byte)
		*bp = strconv.AppendFloat((*bp)[:0], f, 'g', -1, 64)
		w.Write(*bp)
		numBufPool.Put(bp)
	}
}

func trimSuffixes(str string, suffixes []string) string {
	for _, suffix := range suffixes {
		str = strings.TrimSuffix(str, suffix)
	}

	return str
}

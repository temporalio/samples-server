package internal

import (
	"os"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Metrics []Metric
}

type Metric struct {
	MetricName string `yaml:"metric_name"`
	Query      string `yaml:"query"`
}

func LoadConfig(filename string) (*Config, error) {
	var config Config

	bytes, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	err = yaml.Unmarshal(bytes, &config)
	if err != nil {
		return nil, err
	}

	return &config, nil
}

// ByMetricName lets us sort metrics
type ByMetricName []Metric

func (m ByMetricName) Len() int {
	return len(m)
}

func (m ByMetricName) Less(i, j int) bool {
	return m[i].MetricName < m[j].MetricName
}

func (m ByMetricName) Swap(i, j int) {
	m[i], m[j] = m[j], m[i]
}

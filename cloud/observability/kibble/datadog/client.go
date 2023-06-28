package datadog

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/DataDog/datadog-api-client-go/v2/api/datadog"
	"github.com/DataDog/datadog-api-client-go/v2/api/datadogV2"
	"golang.org/x/sync/errgroup"
)

type (
	Submitter interface {
		SubmitMetrics(series []datadogV2.MetricSeries) error
	}

	APIClient struct {
		api *datadogV2.MetricsApi
	}
)

func NewAPIClient() *APIClient {
	configuration := datadog.NewConfiguration()
	configuration.RetryConfiguration.EnableRetry = true
	apiClient := datadog.NewAPIClient(configuration)
	return &APIClient{
		api: datadogV2.NewMetricsApi(apiClient),
	}
}

func (c *APIClient) SubmitMetrics(series []datadogV2.MetricSeries) error {
	pageNum := 0
	pageSize := 100 // calculate this dynamically based on DD's payload size limit
	g := new(errgroup.Group)

	for {
		start, end := paginate(pageNum, pageSize, len(series))
		if start == end {
			break
		}

		g.Go(func() error {
			pagedSeries := series[start:end]

			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer cancel()
			ctx = datadog.NewDefaultContext(context.Background())
			body := datadogV2.MetricPayload{Series: pagedSeries}

			resp, _, err := c.api.SubmitMetrics(ctx, body, *datadogV2.NewSubmitMetricsOptionalParameters())
			if err != nil {
				return fmt.Errorf("failed to submit metrics: %w", err)
			}

			if len(resp.Errors) > 0 {
				responseContent, err := json.MarshalIndent(resp, "", "  ")
				if err != nil {
					return fmt.Errorf("failed to marshal Datadog response: %w", err)
				}
				return fmt.Errorf("failed to submit metrics: %s", responseContent)
			}
			return nil
		})

		pageNum++
	}

	return g.Wait()
}

func paginate(pageNum int, pageSize int, sliceLength int) (int, int) {
	start := pageNum * pageSize

	if start > sliceLength {
		start = sliceLength
	}

	end := start + pageSize
	if end > sliceLength {
		end = sliceLength
	}

	return start, end
}

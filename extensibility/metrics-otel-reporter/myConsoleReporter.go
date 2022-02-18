// The MIT License
//
// Copyright (c) 2020 Temporal Technologies Inc.  All rights reserved.
//
// Copyright (c) 2020 Uber Technologies, Inc.
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
// THE SOFTWARE.

package metrics_reporter

import (
	"context"
	"net/http"
	"time"

	"go.opentelemetry.io/otel/exporters/stdout/stdoutmetric"
	"go.opentelemetry.io/otel/metric"

	"go.opentelemetry.io/otel/metric/global"
	controller "go.opentelemetry.io/otel/sdk/metric/controller/basic"
	processor "go.opentelemetry.io/otel/sdk/metric/processor/basic"
	"go.opentelemetry.io/otel/sdk/resource"
	"go.temporal.io/server/common/metrics"

	"go.temporal.io/server/common/log"
	"go.temporal.io/server/common/log/tag"
)

// Opentelemetry has a lot of distinct reporters implementations and it's most convenient to use those.
var _ metrics.OpentelemetryMustProvider = (*customReporterImpl)(nil)

type (
	// customReporterImpl is a base class for reporting metrics to opentelemetry.
	customReporterImpl struct {
		exporter   *stdoutmetric.Exporter
		meterMust  metric.MeterMust
		pusher     *controller.Controller
		gaugeCache metrics.OtelGaugeCache
	}
)

func NewOpentelemeteryReporter(logger log.Logger) (*customReporterImpl, error) {
	// Prometheus handles histogram boundaries on aggregation level, so have to be handled here
	var defaultHistogramBoundaries []float64
	perUnitBoundaries := make(map[string][]float64, 0)

	exporter, err := stdoutmetric.New(stdoutmetric.WithPrettyPrint())
	if err != nil {
		logger.Fatal(
			"Failed to initialize prometheus exporter.",
			tag.Error(err),
		)
	}

	pusher := controller.New(
		processor.NewFactory(
			metrics.NewOtelAggregatorSelector(
				defaultHistogramBoundaries,
				perUnitBoundaries,
			),
			exporter,
		),
		controller.WithResource(resource.Empty()),
		controller.WithExporter(exporter),
	)

	if err = pusher.Start(context.Background()); err != nil {
		logger.Fatal("starting push controller: %v", tag.Error(err))
	}

	meter := pusher.Meter("temporal")
	meterMust := metric.Must(meter)
	gaugeCache := metrics.NewOtelGaugeCache(meterMust)
	global.SetMeterProvider(pusher)

	reporter := &customReporterImpl{
		exporter:   exporter,
		meterMust:  meterMust,
		pusher:     pusher,
		gaugeCache: gaugeCache,
	}

	return reporter, nil
}

func (r *customReporterImpl) GetMeterMust() metric.MeterMust {
	return metric.Must(r.pusher.Meter("temporal"))
}

func (r *customReporterImpl) Stop(logger log.Logger) {
	ctx, closeCtx := context.WithTimeout(context.Background(), time.Second)
	defer closeCtx()

	if err := r.pusher.Stop(ctx); !(err == nil || err == http.ErrServerClosed) {
		logger.Error("Otel metrics pusher stop fail.", tag.Error(err))
	}
}

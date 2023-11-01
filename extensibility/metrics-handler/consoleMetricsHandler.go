// The MIT License
//
// Copyright (c) 2023 Temporal Technologies Inc.  All rights reserved.
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

package metrics_handler

import (
	"log"
	"time"

	tlog "go.temporal.io/server/common/log"
	"go.temporal.io/server/common/metrics"
)

type (
	// Replace this implementation with your code.
	consoleMetricsHandler struct {
	}

	printFormat struct {
		Operation  string
		Name       string
		Tags       []metrics.Tag
		IntValue   int64
		FloatValue float64
		Interval   time.Duration
	}
)

func NewConsoleMetricsHandler() *consoleMetricsHandler {
	return &consoleMetricsHandler{}
}

func (h consoleMetricsHandler) WithTags(tags ...metrics.Tag) metrics.Handler {
	return &consoleMetricsHandler{}
}

func (h consoleMetricsHandler) Counter(name string) metrics.CounterIface {
	return metrics.CounterFunc(func(value int64, t ...metrics.Tag) {
		doPrint(printFormat{Operation: "Counter", Name: name, Tags: t, IntValue: value})
	})
}

func (h consoleMetricsHandler) Gauge(name string) metrics.GaugeIface {
	return metrics.GaugeFunc(func(value float64, t ...metrics.Tag) {
		doPrint(printFormat{Operation: "Gauge", Name: name, Tags: t, FloatValue: value})
	})
}

func (h consoleMetricsHandler) Timer(name string) metrics.TimerIface {
	return metrics.TimerFunc(func(interval time.Duration, t ...metrics.Tag) {
		doPrint(printFormat{Operation: "Timer", Name: name, Tags: t, Interval: interval})
	})
}

func (h consoleMetricsHandler) Histogram(name string, mUnit metrics.MetricUnit) metrics.HistogramIface {
	return metrics.HistogramFunc(func(value int64, t ...metrics.Tag) {
		doPrint(printFormat{Operation: "Histogram", Name: name, Tags: t, IntValue: value})
	})
}

func (h consoleMetricsHandler) Stop(l tlog.Logger) {
}

func doPrint(src printFormat) {
	log.Print(src)
}

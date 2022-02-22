// The MIT License
//
// Copyright (c) 2020 Temporal Technologies Inc.  All rights reserved.
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

package metrics

import (
	"fmt"
	"log"
	"time"

	"github.com/uber-go/tally/v4"
)

type (
	// Replace this implementation with your code.
	myConsoleReporter struct {
	}

	printFormat struct {
		Operation  string
		Name       string
		Tags       map[string]string
		IntValue   int64
		FloatValue float64
		Interval   time.Duration
	}
)

func NewReporter() tally.StatsReporter {
	return &myConsoleReporter{}
}

func (r myConsoleReporter) Capabilities() tally.Capabilities {
	return r
}

func (r myConsoleReporter) Reporting() bool {
	return true
}

func (r myConsoleReporter) Tagging() bool {
	return true
}

func (r myConsoleReporter) Flush() {
}

func (r myConsoleReporter) ReportCounter(name string, tags map[string]string, value int64) {
	doPrint(printFormat{Operation: "Counter", Name: name, Tags: tags, IntValue: value})
}

func (r myConsoleReporter) ReportGauge(name string, tags map[string]string, value float64) {
	doPrint(printFormat{Operation: "Gauge", Name: name, Tags: tags, FloatValue: value})
}

func (r myConsoleReporter) ReportTimer(name string, tags map[string]string, interval time.Duration) {
	doPrint(printFormat{Operation: "Timer", Name: name, Tags: tags, Interval: interval})
}

func (r myConsoleReporter) ReportHistogramValueSamples(
	name string, tags map[string]string, buckets tally.Buckets, bucketLowerBound, bucketUpperBound float64,
	samples int64,
) {
	doPrint(printFormat{Operation: "HistoValueSamples", Name: name, Tags: tags})
}

func (r myConsoleReporter) ReportHistogramDurationSamples(
	name string, tags map[string]string, buckets tally.Buckets, bucketLowerBound, bucketUpperBound time.Duration,
	samples int64,
) {
	doPrint(printFormat{Operation: "HistoDurationSamples", Name: name, Tags: tags})
}

func doPrint(src printFormat) {
	log.Println(src)
	fmt.Println(src)
}

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

package main

import (
	"log"

	"go.temporal.io/server/temporal"
	"go.temporal.io/server/common/config"
	// tmetrics "go.temporal.io/server/common/metrics"
	tlog "go.temporal.io/server/common/log"


	"github.com/temporalio/service-samples/metrics-reporter"
)

func main() {
	s, err := temporal.InitializeServer(
		temporal.ForServices(temporal.Services),
		temporal.WithConfigLoader("./metrics-reporter/config", "development", ""),
		temporal.InterruptOn(temporal.InterruptCh()),
		temporal.WithMetricsReportersProvider(func(c *config.Config, log tlog.Logger) (*temporal.MetricsReporters, error){
			serverMc, sdkMc, err := c.Global.Metrics.InitMetricReporters(log, metrics.NewReporter())
			if err != nil {
				return nil, err
			}

			result := &temporal.MetricsReporters{serverMc, sdkMc}
			return result, nil
		}),
	)
	if err != nil {
		log.Fatal(err)
	}

	err = s.Start()
	if err != nil {
		log.Fatal(err)
	}

	log.Println("All services are stopped.")
}

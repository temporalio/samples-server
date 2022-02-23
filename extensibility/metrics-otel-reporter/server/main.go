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
	"context"

	"go.temporal.io/server/common/log"
	"go.temporal.io/server/common/metrics"
	"go.temporal.io/server/temporal"
	"google.golang.org/grpc"

	metrics_otel_reporter "github.com/temporalio/service-samples/metrics-otel-reporter"
)

func NewCustomInterceptor(scope metrics.UserScope) grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req interface{},
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (resp interface{}, err error) {
		scope.AddCounter(
			"CustomInterceptorInvoked",
			1,
		)
		return handler(
			ctx,
			req,
		)
	}
}

func main() {
	ctx := context.Background()
	logger := log.NewCLILogger()
	mustProvider, err := metrics_otel_reporter.NewOpentelemeteryReporter(logger)
	if err != nil {
		logger.Fatal(err.Error())
	}

	reporter, err2 := metrics.NewOpentelemeteryReporter(logger, &metrics.ClientConfig{}, mustProvider)
	if err2 != nil {
		logger.Fatal(err2.Error())
	}

	customInterceptor := NewCustomInterceptor(reporter.UserScope())

	s := temporal.NewServer(
		temporal.ForServices(temporal.Services),
		temporal.WithConfigLoader("./metrics-otel-reporter/config", "development", ""),
		temporal.InterruptOn(temporal.InterruptCh()),
		temporal.WithCustomMetricsReporter(reporter),
		temporal.WithChainedFrontendGrpcInterceptors(customInterceptor),
	)

	mustProvider.GetMeterMust().NewInt64Counter("test").Add(ctx, 11)

	err = s.Start()
	if err != nil {
		logger.Fatal(err.Error())
	}
}


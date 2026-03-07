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

	"go.temporal.io/server/common/authorization"
	"go.temporal.io/server/common/config"
	"go.temporal.io/server/temporal"

	"github.com/temporalio/service-samples/authorizer"
)

func newServer(configFile string, opts ...temporal.ServerOption) (temporal.Server, error) {
	cfg, err := config.Load(config.WithConfigFile(configFile))
	if err != nil {
		return nil, err
	}

	defaults := []temporal.ServerOption{
		temporal.ForServices(temporal.DefaultServices),
		temporal.WithConfig(cfg),
		temporal.WithClaimMapper(func(cfg *config.Config) authorization.ClaimMapper {
			return authorizer.NewMyClaimMapper(cfg)
		}),
		temporal.WithAuthorizer(authorizer.NewMyAuthorizer()),
	}

	return temporal.NewServer(append(defaults, opts...)...)
}

func main() {
	// InterruptOn is passed here rather than in newServer so tests can call s.Stop() directly.
	// Include this in production scenarios to enable graceful shutdown on SIGINT/SIGTERM.
	s, err := newServer("./config/development.yaml", temporal.InterruptOn(temporal.InterruptCh()))
	if err != nil {
		log.Fatal(err)
	}

	err = s.Start()
	if err != nil {
		log.Fatal(err)
	}

	log.Println("All services are stopped.")
}

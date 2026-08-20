// The MIT License
//
// Copyright (c) 2026 Temporal Technologies Inc.  All rights reserved.
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
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"go.temporal.io/api/workflowservice/v1"
	"go.temporal.io/server/common/config"
	_ "go.temporal.io/server/common/persistence/sql/sqlplugin/sqlite"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// TestServerBootsWithTokenProvider exercises the WithTokenProvider wiring:
// the server's boot validation rejects the combination (require=true,
// no TokenProvider), so a clean boot here confirms the provider made it into
// the server's options. Replication isn't exercised — that requires a second
// cluster.
func TestServerBootsWithTokenProvider(t *testing.T) {
	tokenFile := filepath.Join(t.TempDir(), "token")
	if err := os.WriteFile(tokenFile, []byte("test-bearer-token"), 0o600); err != nil {
		t.Fatal(err)
	}
	t.Setenv(tokenFileEnv, tokenFile)

	s, err := newServer("testdata/config.yaml")
	if err != nil {
		t.Fatal(err)
	}

	if err := s.Start(); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = s.Stop() })

	cfg, err := config.Load(config.WithConfigFile("testdata/config.yaml"))
	if err != nil {
		t.Fatal(err)
	}
	frontendAddr := fmt.Sprintf("127.0.0.1:%d", cfg.Services["frontend"].RPC.GRPCPort)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	conn, err := grpc.NewClient(frontendAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()

	wfClient := workflowservice.NewWorkflowServiceClient(conn)
	for {
		_, err := wfClient.GetSystemInfo(ctx, &workflowservice.GetSystemInfoRequest{})
		if err == nil {
			return
		}
		select {
		case <-ctx.Done():
			t.Fatalf("timed out waiting for server: %v", err)
		case <-time.After(200 * time.Millisecond):
		}
	}
}

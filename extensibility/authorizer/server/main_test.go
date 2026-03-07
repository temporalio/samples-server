package main

import (
	"context"
	"fmt"
	"testing"
	"time"

	"go.temporal.io/api/operatorservice/v1"
	"go.temporal.io/api/workflowservice/v1"
	"go.temporal.io/server/common/config"
	_ "go.temporal.io/server/common/persistence/sql/sqlplugin/sqlite"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/durationpb"
)

func TestAuthorizerDeniesUpdateNamespace(t *testing.T) {
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
	opClient := operatorservice.NewOperatorServiceClient(conn)

	// Wait for server to be ready
	for {
		_, err := wfClient.GetSystemInfo(ctx, &workflowservice.GetSystemInfoRequest{})
		if err == nil {
			break
		}
		select {
		case <-ctx.Done():
			t.Fatal("timed out waiting for server")
		case <-time.After(200 * time.Millisecond):
		}
	}

	// Create a namespace (should succeed)
	ns := "test-authorizer"
	_, err = wfClient.RegisterNamespace(ctx, &workflowservice.RegisterNamespaceRequest{
		Namespace:                        ns,
		WorkflowExecutionRetentionPeriod: durationpb.New(24 * time.Hour),
	})
	if err != nil {
		t.Fatalf("RegisterNamespace should succeed: %v", err)
	}

	// Wait for namespace to propagate
	time.Sleep(2 * time.Second)

	// UpdateNamespace should be denied by the authorizer
	_, err = wfClient.UpdateNamespace(ctx, &workflowservice.UpdateNamespaceRequest{
		Namespace: ns,
	})
	if err == nil {
		t.Fatal("UpdateNamespace should have been denied")
	}
	if s, ok := status.FromError(err); !ok || s.Code() != codes.PermissionDenied {
		t.Fatalf("expected PermissionDenied, got: %v", err)
	}

	// DeleteNamespace should still be allowed
	_, err = opClient.DeleteNamespace(ctx, &operatorservice.DeleteNamespaceRequest{
		Namespace: ns,
	})
	if err != nil {
		t.Fatalf("DeleteNamespace should succeed: %v", err)
	}
}

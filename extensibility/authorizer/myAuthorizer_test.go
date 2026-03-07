package authorizer

import (
	"context"
	"testing"

	"go.temporal.io/server/common/authorization"
)

func TestMyAuthorizer(t *testing.T) {
	a := NewMyAuthorizer()
	ctx := context.Background()

	tests := []struct {
		name     string
		claims   *authorization.Claims
		target   *authorization.CallTarget
		decision authorization.Decision
	}{
		{
			name:     "allow temporal-system namespace",
			claims:   nil,
			target:   &authorization.CallTarget{Namespace: "temporal-system", APIName: "UpdateNamespace"},
			decision: authorization.DecisionAllow,
		},
		{
			name:     "allow system admin",
			claims:   &authorization.Claims{System: authorization.RoleAdmin},
			target:   &authorization.CallTarget{Namespace: "test", APIName: "UpdateNamespace"},
			decision: authorization.DecisionAllow,
		},
		{
			name:     "deny UpdateNamespace without claims",
			claims:   nil,
			target:   &authorization.CallTarget{Namespace: "test", APIName: "UpdateNamespace"},
			decision: authorization.DecisionDeny,
		},
		{
			name:     "deny UpdateNamespace with reader role",
			claims:   &authorization.Claims{Namespaces: map[string]authorization.Role{"test": authorization.RoleReader}},
			target:   &authorization.CallTarget{Namespace: "test", APIName: "UpdateNamespace"},
			decision: authorization.DecisionDeny,
		},
		{
			name:     "allow UpdateNamespace with namespace writer role",
			claims:   &authorization.Claims{Namespaces: map[string]authorization.Role{"test": authorization.RoleWriter}},
			target:   &authorization.CallTarget{Namespace: "test", APIName: "UpdateNamespace"},
			decision: authorization.DecisionAllow,
		},
		{
			name:     "allow other calls without claims",
			claims:   nil,
			target:   &authorization.CallTarget{Namespace: "test", APIName: "ListNamespaces"},
			decision: authorization.DecisionAllow,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := a.Authorize(ctx, tt.claims, tt.target)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if result.Decision != tt.decision {
				t.Fatalf("expected %v, got %v", tt.decision, result.Decision)
			}
		})
	}
}

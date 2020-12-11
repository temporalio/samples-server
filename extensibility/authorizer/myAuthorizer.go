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

package authorizer

import (
	"context"

	"go.temporal.io/server/common/authorization"
)

type myAuthorizer struct{}

func NewMyAuthorizer() authorization.Authorizer {
	return &myAuthorizer{}
}

var decisionAllow = authorization.Result{Decision: authorization.DecisionAllow}
var decisionDeny = authorization.Result{Decision: authorization.DecisionDeny}

func (a *myAuthorizer) Authorize(_ context.Context, claims *authorization.Claims,
	target *authorization.CallTarget) (authorization.Result, error) {

	// Allow all operations within "temporal-system" namespace
	if target.Namespace == "temporal-system" {
		return decisionAllow, nil
	}

	// Allow all calls except UpdateNamespace through when claim mapper isn't invoked.
	// Claim mapper is skipped unless TLS is configured or an auth token is passed
	if claims == nil && target.APIName != "UpdateNamespace" {
		return decisionAllow, nil
	}

	// Allow all operations for system-level admins and writers
	if claims.System & (authorization.RoleAdmin | authorization.RoleWriter) != 0 {
		return decisionAllow, nil
	}

	// For other namespaces, deny "UpdateNamespace" API unless the caller has a writer role in it
	if target.APIName == "UpdateNamespace" {
		if claims.Namespaces[target.Namespace] & authorization.RoleWriter != 0 {
			return decisionAllow, nil
		} else {
			return decisionDeny, nil
		}
	}

	// Allow all other requests
	return decisionAllow, nil
}

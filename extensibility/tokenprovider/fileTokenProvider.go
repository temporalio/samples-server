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

package tokenprovider

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"

	"go.temporal.io/server/common/rpc/auth"
	"golang.org/x/oauth2"
)

// fileTokenSource implements oauth2.TokenSource by reading a JWT from disk
// and reporting its `exp` claim as the token's Expiry. A token with no parsable
// expiry returns a zero Expiry, which oauth2.ReuseTokenSource treats as
// "never expires."
//
// This is the file-backed analog of a network token source such as the one
// returned by clientcredentials.Config.TokenSource(ctx). Wrapping either in
// oauth2.ReuseTokenSource gives expiry-aware caching for free.
type fileTokenSource struct {
	path string
}

func (s *fileTokenSource) Token() (*oauth2.Token, error) {
	data, err := os.ReadFile(s.path)
	if err != nil {
		return nil, fmt.Errorf("read token file %q: %w", s.path, err)
	}
	token := strings.TrimSpace(string(data))
	if token == "" {
		return nil, fmt.Errorf("token file %q is empty", s.path)
	}
	return &oauth2.Token{
		AccessToken: token,
		Expiry:      parseJWTExpiry(token),
	}, nil
}

// fileTokenProvider implements auth.TokenProvider by delegating to a
// cached oauth2.TokenSource.
//
// The Temporal Service does not cache TokenProvider responses, so providers
// own the lifecycle (caching, refresh, deduplication of concurrent fetches).
// Here that responsibility is delegated to oauth2.ReuseTokenSource, which
// hands back the cached token until it's near its Expiry and only then calls
// through to the underlying source.
//
// For a file source this is arguably overengineered — local reads are cheap
// and a per-RPC re-read would be fine. The structure exists to mirror a
// network-backed provider: swap &fileTokenSource{path} for an oauth2
// clientcredentials.Config{...}.TokenSource(ctx) and nothing else changes.
// That symmetry is the point of the sample.
type fileTokenProvider struct {
	source oauth2.TokenSource
}

// NewFileTokenProvider returns an auth.TokenProvider that reads its token from
// the file at path and caches it until it nears its JWT `exp` claim.
func NewFileTokenProvider(path string) auth.TokenProvider {
	// Passing nil as the initial token forces ReuseTokenSource to call the
	// underlying source on the first GetToken; subsequent calls reuse the
	// cached value until it's within the library's default early-expiry
	// window (10 seconds).
	return &fileTokenProvider{
		source: oauth2.ReuseTokenSource(nil, &fileTokenSource{path: path}),
	}
}

func (p *fileTokenProvider) GetToken(_ context.Context, _ string) (string, error) {
	tok, err := p.source.Token()
	if err != nil {
		return "", err
	}
	return tok.AccessToken, nil
}

// parseJWTExpiry extracts the exp claim from a JWT without verifying the
// signature. The receiver validates the signature; this provider only needs
// the expiry to drive cache rotation. Returns the zero time if the token is
// not a JWT or has no exp claim, which oauth2.ReuseTokenSource interprets as
// "never expires."
func parseJWTExpiry(token string) time.Time {
	parts := strings.SplitN(token, ".", 3)
	if len(parts) < 2 {
		return time.Time{}
	}
	payload, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return time.Time{}
	}
	var claims struct {
		Exp *float64 `json:"exp"`
	}
	if err := json.Unmarshal(payload, &claims); err != nil || claims.Exp == nil {
		return time.Time{}
	}
	return time.Unix(int64(*claims.Exp), 0)
}

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
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestFileTokenSource(t *testing.T) {
	tests := []struct {
		name        string
		write       string
		want        string
		wantExpiry  bool
		wantErrText string
	}{
		{
			name:  "opaque token has zero expiry",
			write: "opaque-bearer-token",
			want:  "opaque-bearer-token",
		},
		{
			name:       "JWT propagates exp claim",
			write:      "",   // populated per-case below
			wantExpiry: true,
		},
		{
			name:  "trims surrounding whitespace",
			write: "  spaced-token\n",
			want:  "spaced-token",
		},
		{
			name:        "empty file is an error",
			write:       "",
			wantErrText: "is empty",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			content := tt.write
			var expectedExpiry time.Time
			if tt.wantExpiry {
				expectedExpiry = time.Now().Add(time.Hour).Truncate(time.Second)
				content = makeJWT(t, expectedExpiry)
			}

			path := filepath.Join(t.TempDir(), "token")
			if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
				t.Fatal(err)
			}

			s := &fileTokenSource{path: path}
			tok, err := s.Token()

			if tt.wantErrText != "" {
				if err == nil {
					t.Fatalf("expected error containing %q, got nil", tt.wantErrText)
				}
				if !strings.Contains(err.Error(), tt.wantErrText) {
					t.Fatalf("error = %v, want substring %q", err, tt.wantErrText)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if tt.want != "" && tok.AccessToken != tt.want {
				t.Fatalf("AccessToken = %q, want %q", tok.AccessToken, tt.want)
			}
			if tt.wantExpiry {
				if !tok.Expiry.Equal(expectedExpiry) {
					t.Fatalf("Expiry = %v, want %v", tok.Expiry, expectedExpiry)
				}
			} else if !tok.Expiry.IsZero() {
				t.Fatalf("expected zero Expiry, got %v", tok.Expiry)
			}
		})
	}
}

func TestFileTokenSource_MissingFile(t *testing.T) {
	s := &fileTokenSource{path: filepath.Join(t.TempDir(), "does-not-exist")}
	if _, err := s.Token(); err == nil {
		t.Fatal("expected error for missing file")
	}
}

// TestFileTokenProvider_CachesToken verifies that the oauth2.ReuseTokenSource
// wrapper does its job: once the provider has read a token, subsequent calls
// don't hit the underlying source. The test demonstrates this by deleting the
// file after the first call and confirming the second still succeeds.
func TestFileTokenProvider_CachesToken(t *testing.T) {
	path := filepath.Join(t.TempDir(), "token")
	if err := os.WriteFile(path, []byte("cached-token"), 0o600); err != nil {
		t.Fatal(err)
	}

	p := NewFileTokenProvider(path)
	tok1, err := p.GetToken(context.Background(), "")
	if err != nil {
		t.Fatalf("first GetToken: %v", err)
	}
	if tok1 != "cached-token" {
		t.Fatalf("first GetToken = %q, want %q", tok1, "cached-token")
	}

	if err := os.Remove(path); err != nil {
		t.Fatal(err)
	}

	tok2, err := p.GetToken(context.Background(), "")
	if err != nil {
		t.Fatalf("second GetToken should be cached, got error: %v", err)
	}
	if tok2 != tok1 {
		t.Fatalf("second GetToken = %q, want cached %q", tok2, tok1)
	}
}

func TestParseJWTExpiry(t *testing.T) {
	t.Run("valid JWT", func(t *testing.T) {
		want := time.Unix(1700000000, 0)
		token := makeJWTWithUnixExp(t, 1700000000)
		got := parseJWTExpiry(token)
		if !got.Equal(want) {
			t.Fatalf("parseJWTExpiry = %v, want %v", got, want)
		}
	})
	t.Run("JWT with no exp claim", func(t *testing.T) {
		token := makeJWTWithUnixExp(t, 0) // 0 == omit
		if got := parseJWTExpiry(token); !got.IsZero() {
			t.Fatalf("parseJWTExpiry = %v, want zero", got)
		}
	})
	t.Run("non-JWT inputs", func(t *testing.T) {
		for _, s := range []string{"", "not-a-jwt", "a.!!!invalid-base64.c"} {
			if got := parseJWTExpiry(s); !got.IsZero() {
				t.Fatalf("parseJWTExpiry(%q) = %v, want zero", s, got)
			}
		}
	})
}

// makeJWT builds an unsigned JWT-shaped string with the given exp claim. The
// receiver validates signatures; tests only need to exercise the parser.
func makeJWT(t *testing.T, exp time.Time) string {
	t.Helper()
	return makeJWTWithUnixExp(t, exp.Unix())
}

// makeJWTWithUnixExp builds an unsigned JWT with the given exp claim. An expUnix
// of 0 omits the exp claim entirely.
func makeJWTWithUnixExp(t *testing.T, expUnix int64) string {
	t.Helper()
	header := base64.RawURLEncoding.EncodeToString([]byte(`{"alg":"none","typ":"JWT"}`))
	claims := map[string]any{}
	if expUnix != 0 {
		claims["exp"] = expUnix
	}
	payload, err := json.Marshal(claims)
	if err != nil {
		t.Fatal(err)
	}
	return header + "." + base64.RawURLEncoding.EncodeToString(payload) + ".sig"
}

package cloudflareaccess

import (
	"context"
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
)

type staticFetcher struct {
	keys map[string]*rsa.PublicKey
	err  error
}

func (fetcher staticFetcher) Fetch(context.Context) (map[string]*rsa.PublicKey, error) {
	return fetcher.keys, fetcher.err
}

func TestMiddlewareRejectsInvalidAssertionsAndAllowsValidAssertion(t *testing.T) {
	gin.SetMode(gin.TestMode)
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatal(err)
	}
	wrongPrivateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatal(err)
	}
	config := Config{Issuer: "https://team.example.com", Audience: "expected-audience"}
	now := time.Now()
	validClaims := map[string]any{
		"iss": config.Issuer, "aud": []string{config.Audience},
		"exp": now.Add(time.Minute).Unix(), "nbf": now.Add(-time.Minute).Unix(),
	}

	tests := []struct {
		name       string
		token      string
		fetcher    KeyFetcher
		wantStatus int
		wantNext   bool
	}{
		{
			name:       "missing header",
			wantStatus: http.StatusUnauthorized,
		},
		{
			name:       "malformed JWT",
			token:      "not-a-jwt",
			wantStatus: http.StatusUnauthorized,
		},
		{
			name:       "unsupported algorithm",
			token:      signTokenWithAlgorithm(t, privateKey, "key-1", "HS256", validClaims),
			wantStatus: http.StatusUnauthorized,
		},
		{
			name:       "invalid signature",
			token:      signToken(t, wrongPrivateKey, "key-1", validClaims),
			fetcher:    staticFetcher{keys: map[string]*rsa.PublicKey{"key-1": &privateKey.PublicKey}},
			wantStatus: http.StatusUnauthorized,
		},
		{
			name:       "issuer mismatch",
			token:      signToken(t, privateKey, "key-1", claimsWith(validClaims, "iss", "https://other.example.com")),
			fetcher:    staticFetcher{keys: map[string]*rsa.PublicKey{"key-1": &privateKey.PublicKey}},
			wantStatus: http.StatusUnauthorized,
		},
		{
			name:       "audience mismatch",
			token:      signToken(t, privateKey, "key-1", claimsWith(validClaims, "aud", []string{"other-audience"})),
			fetcher:    staticFetcher{keys: map[string]*rsa.PublicKey{"key-1": &privateKey.PublicKey}},
			wantStatus: http.StatusUnauthorized,
		},
		{
			name:       "expired",
			token:      signToken(t, privateKey, "key-1", claimsWith(validClaims, "exp", now.Add(-time.Minute).Unix())),
			fetcher:    staticFetcher{keys: map[string]*rsa.PublicKey{"key-1": &privateKey.PublicKey}},
			wantStatus: http.StatusUnauthorized,
		},
		{
			name:       "not yet valid",
			token:      signToken(t, privateKey, "key-1", claimsWith(validClaims, "nbf", now.Add(time.Minute).Unix())),
			fetcher:    staticFetcher{keys: map[string]*rsa.PublicKey{"key-1": &privateKey.PublicKey}},
			wantStatus: http.StatusUnauthorized,
		},
		{
			name:       "key fetch failure",
			token:      signToken(t, privateKey, "key-1", validClaims),
			fetcher:    staticFetcher{err: errors.New("unavailable")},
			wantStatus: http.StatusUnauthorized,
		},
		{
			name:       "valid assertion",
			token:      signToken(t, privateKey, "key-1", validClaims),
			fetcher:    staticFetcher{keys: map[string]*rsa.PublicKey{"key-1": &privateKey.PublicKey}},
			wantStatus: http.StatusNoContent,
			wantNext:   true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			fetcher := test.fetcher
			if fetcher == nil {
				fetcher = staticFetcher{keys: map[string]*rsa.PublicKey{"key-1": &privateKey.PublicKey}}
			}
			middleware, err := NewMiddleware(config, fetcher)
			if err != nil {
				t.Fatal(err)
			}
			nextCalled := false
			router := gin.New()
			router.POST("/mcp", middleware, func(c *gin.Context) {
				nextCalled = true
				c.Status(http.StatusNoContent)
			})
			request := httptest.NewRequest(http.MethodPost, "/mcp", nil)
			if test.token != "" {
				request.Header.Set(assertionHeader, test.token)
			}
			response := httptest.NewRecorder()
			router.ServeHTTP(response, request)

			if response.Code != test.wantStatus {
				t.Fatalf("status = %d, want %d", response.Code, test.wantStatus)
			}
			if nextCalled != test.wantNext {
				t.Fatalf("next called = %t, want %t", nextCalled, test.wantNext)
			}
			if test.token != "" && !test.wantNext && strings.Contains(response.Body.String(), test.token) {
				t.Fatal("unauthorized response exposed the assertion")
			}
		})
	}
}

func TestKeyCacheRefreshesForUnknownKID(t *testing.T) {
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatal(err)
	}
	fetchCount := 0
	fetcher := keyFetcherFunc(func(context.Context) (map[string]*rsa.PublicKey, error) {
		fetchCount++
		if fetchCount == 1 {
			return map[string]*rsa.PublicKey{"old-kid": &privateKey.PublicKey}, nil
		}
		return map[string]*rsa.PublicKey{"new-kid": &privateKey.PublicKey}, nil
	})
	cache := &keyCache{fetcher: fetcher, ttl: defaultCacheTTL}
	if _, err := cache.get(context.Background(), "old-kid"); err != nil {
		t.Fatal(err)
	}
	if _, err := cache.get(context.Background(), "new-kid"); err != nil {
		t.Fatal(err)
	}
	if fetchCount != 2 {
		t.Fatalf("JWKS fetch count = %d, want 2", fetchCount)
	}
}

func TestLoadConfigFromEnvRequiresValidIssuerAndAudience(t *testing.T) {
	t.Setenv("CLOUDFLARE_ACCESS_ISSUER", "https://team.example.com")
	t.Setenv("CLOUDFLARE_ACCESS_AUD", "application-audience")
	t.Setenv("CLOUDFLARE_ACCESS_JWKS_URL", "")
	config, err := LoadConfigFromEnv()
	if err != nil {
		t.Fatal(err)
	}
	if config.JWKSURL != "https://team.example.com/cdn-cgi/access/certs" {
		t.Fatalf("JWKS URL = %q", config.JWKSURL)
	}

	t.Setenv("CLOUDFLARE_ACCESS_ISSUER", "")
	if _, err := LoadConfigFromEnv(); err == nil || !strings.Contains(err.Error(), "CLOUDFLARE_ACCESS_ISSUER") {
		t.Fatalf("missing issuer error = %v", err)
	}
	t.Setenv("CLOUDFLARE_ACCESS_ISSUER", "http://team.example.com")
	if _, err := LoadConfigFromEnv(); err == nil || !strings.Contains(err.Error(), "must use HTTPS") {
		t.Fatalf("invalid issuer error = %v", err)
	}
}

func signToken(t *testing.T, privateKey *rsa.PrivateKey, kid string, claims map[string]any) string {
	return signTokenWithAlgorithm(t, privateKey, kid, "RS256", claims)
}

func signTokenWithAlgorithm(t *testing.T, privateKey *rsa.PrivateKey, kid, algorithm string, claims map[string]any) string {
	t.Helper()
	header, err := json.Marshal(map[string]string{"alg": algorithm, "kid": kid, "typ": "JWT"})
	if err != nil {
		t.Fatal(err)
	}
	payload, err := json.Marshal(claims)
	if err != nil {
		t.Fatal(err)
	}
	encodedHeader := base64.RawURLEncoding.EncodeToString(header)
	encodedPayload := base64.RawURLEncoding.EncodeToString(payload)
	signingInput := encodedHeader + "." + encodedPayload
	digest := sha256.Sum256([]byte(signingInput))
	signature, err := rsa.SignPKCS1v15(rand.Reader, privateKey, crypto.SHA256, digest[:])
	if err != nil {
		t.Fatal(err)
	}
	return signingInput + "." + base64.RawURLEncoding.EncodeToString(signature)
}

func claimsWith(claims map[string]any, key string, value any) map[string]any {
	copy := make(map[string]any, len(claims))
	for claim, claimValue := range claims {
		copy[claim] = claimValue
	}
	copy[key] = value
	return copy
}

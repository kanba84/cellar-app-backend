package cloudflareaccess

import (
	"context"
	"crypto"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"errors"
	"fmt"
	"io"
	"log"
	"math/big"
	"net/http"
	"net/url"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

const (
	assertionHeader = "Cf-Access-Jwt-Assertion"
	maxJWTSize      = 16 * 1024
	maxJWKSSize     = 1 << 20
	defaultCacheTTL = 5 * time.Minute
)

var errUnknownKeyID = errors.New("signing key ID not found")

type Config struct {
	Issuer   string
	Audience string
	JWKSURL  string
}

type KeyFetcher interface {
	Fetch(context.Context) (map[string]*rsa.PublicKey, error)
}

type keyFetcherFunc func(context.Context) (map[string]*rsa.PublicKey, error)

func (f keyFetcherFunc) Fetch(ctx context.Context) (map[string]*rsa.PublicKey, error) {
	return f(ctx)
}

type keyCache struct {
	fetcher KeyFetcher
	ttl     time.Duration
	mu      sync.Mutex
	keys    map[string]*rsa.PublicKey
	expires time.Time
}

type jwksFetcher struct {
	url    string
	client *http.Client
}

type jwksDocument struct {
	Keys []jwk `json:"keys"`
}

type jwk struct {
	KeyID     string   `json:"kid"`
	KeyType   string   `json:"kty"`
	Algorithm string   `json:"alg"`
	Use       string   `json:"use"`
	Modulus   string   `json:"n"`
	Exponent  string   `json:"e"`
	Cert      string   `json:"cert"`
	CertChain []string `json:"x5c"`
}

type jwtHeader struct {
	Algorithm string `json:"alg"`
	KeyID     string `json:"kid"`
}

type jwtClaims struct {
	Issuer    string          `json:"iss"`
	Audience  json.RawMessage `json:"aud"`
	ExpiresAt json.Number     `json:"exp"`
	NotBefore json.Number     `json:"nbf"`
}

func LoadConfigFromEnv() (Config, error) {
	config := Config{
		Issuer:   strings.TrimSpace(os.Getenv("CLOUDFLARE_ACCESS_ISSUER")),
		Audience: strings.TrimSpace(os.Getenv("CLOUDFLARE_ACCESS_AUD")),
		JWKSURL:  strings.TrimSpace(os.Getenv("CLOUDFLARE_ACCESS_JWKS_URL")),
	}
	if err := config.validate(); err != nil {
		return Config{}, err
	}
	if config.JWKSURL == "" {
		config.JWKSURL = strings.TrimRight(config.Issuer, "/") + "/cdn-cgi/access/certs"
	}
	if err := validateURL(config.JWKSURL); err != nil {
		return Config{}, fmt.Errorf("invalid CLOUDFLARE_ACCESS_JWKS_URL: %w", err)
	}
	return config, nil
}

func (config Config) validate() error {
	if config.Issuer == "" {
		return errors.New("CLOUDFLARE_ACCESS_ISSUER is required")
	}
	if config.Audience == "" {
		return errors.New("CLOUDFLARE_ACCESS_AUD is required")
	}
	if err := validateURL(config.Issuer); err != nil {
		return fmt.Errorf("invalid CLOUDFLARE_ACCESS_ISSUER: %w", err)
	}
	return nil
}

func validateURL(rawURL string) error {
	parsed, err := url.Parse(rawURL)
	if err != nil || parsed.Host == "" || parsed.User != nil || parsed.RawQuery != "" || parsed.Fragment != "" {
		return errors.New("must be an absolute URL without credentials, query, or fragment")
	}
	if parsed.Scheme != "https" {
		return errors.New("must use HTTPS")
	}
	return nil
}

func NewMiddleware(config Config, fetcher KeyFetcher) (gin.HandlerFunc, error) {
	if err := config.validate(); err != nil {
		return nil, err
	}
	if config.JWKSURL != "" {
		if err := validateURL(config.JWKSURL); err != nil {
			return nil, fmt.Errorf("invalid JWKS URL: %w", err)
		}
	}
	config.Issuer = strings.TrimRight(config.Issuer, "/")
	if fetcher == nil {
		jwksURL := config.JWKSURL
		if jwksURL == "" {
			jwksURL = config.Issuer + "/cdn-cgi/access/certs"
		}
		fetcher = &jwksFetcher{
			url:    jwksURL,
			client: &http.Client{Timeout: 5 * time.Second},
		}
	}
	cache := &keyCache{fetcher: fetcher, ttl: defaultCacheTTL}

	return func(c *gin.Context) {
		token := c.GetHeader(assertionHeader)
		if token == "" {
			reject(c, "missing_assertion")
			return
		}
		reason := verifyAssertion(c.Request.Context(), token, config, cache, time.Now())
		if reason != "" {
			reject(c, reason)
			return
		}
		log.Printf("Cloudflare Access authentication succeeded: method=%q path=%q", c.Request.Method, c.Request.URL.Path)
		c.Next()
	}, nil
}

func reject(c *gin.Context, reason string) {
	log.Printf("Cloudflare Access authentication failed: method=%q path=%q reason=%s", c.Request.Method, c.Request.URL.Path, reason)
	c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
}

func verifyAssertion(ctx context.Context, token string, config Config, cache *keyCache, now time.Time) string {
	if len(token) > maxJWTSize {
		return "malformed_jwt"
	}
	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		return "malformed_jwt"
	}
	headerBytes, err := base64.RawURLEncoding.DecodeString(parts[0])
	if err != nil {
		return "malformed_jwt"
	}
	var header jwtHeader
	if err := json.Unmarshal(headerBytes, &header); err != nil || header.Algorithm != "RS256" || header.KeyID == "" {
		return "malformed_jwt"
	}
	payload, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return "malformed_jwt"
	}
	claimsDecoder := json.NewDecoder(strings.NewReader(string(payload)))
	claimsDecoder.UseNumber()
	var claims jwtClaims
	if err := claimsDecoder.Decode(&claims); err != nil {
		return "malformed_jwt"
	}
	var trailingData any
	if err := claimsDecoder.Decode(&trailingData); err != io.EOF {
		return "malformed_jwt"
	}

	key, err := cache.get(ctx, header.KeyID)
	if err != nil {
		if errors.Is(err, errUnknownKeyID) {
			return "unknown_kid"
		}
		return "public_key_unavailable"
	}
	signature, err := base64.RawURLEncoding.DecodeString(parts[2])
	if err != nil {
		return "malformed_jwt"
	}
	digest := sha256.Sum256([]byte(parts[0] + "." + parts[1]))
	if err := rsa.VerifyPKCS1v15(key, crypto.SHA256, digest[:], signature); err != nil {
		return "invalid_signature"
	}
	if claims.Issuer != config.Issuer {
		return "issuer_mismatch"
	}
	if !containsAudience(claims.Audience, config.Audience) {
		return "audience_mismatch"
	}
	exp, err := claims.ExpiresAt.Int64()
	if err != nil || exp <= now.Unix() {
		return "expired_or_invalid_exp"
	}
	nbf, err := claims.NotBefore.Int64()
	if err != nil {
		return "not_yet_valid_or_invalid_nbf"
	}
	if nbf > now.Unix() {
		return "not_yet_valid_or_invalid_nbf"
	}
	return ""
}

func containsAudience(raw json.RawMessage, expected string) bool {
	var single string
	if json.Unmarshal(raw, &single) == nil {
		return single == expected
	}
	var multiple []string
	if json.Unmarshal(raw, &multiple) != nil {
		return false
	}
	for _, audience := range multiple {
		if audience == expected {
			return true
		}
	}
	return false
}

func (cache *keyCache) get(ctx context.Context, kid string) (*rsa.PublicKey, error) {
	cache.mu.Lock()
	defer cache.mu.Unlock()

	if time.Now().Before(cache.expires) {
		if key, ok := cache.keys[kid]; ok && key != nil {
			return key, nil
		}
	}
	keys, err := cache.fetcher.Fetch(ctx)
	if err != nil {
		return nil, err
	}
	if len(keys) == 0 {
		return nil, errors.New("JWKS contains no usable signing keys")
	}
	cache.keys = keys
	cache.expires = time.Now().Add(cache.ttl)
	key, ok := cache.keys[kid]
	if !ok || key == nil {
		return nil, errUnknownKeyID
	}
	return key, nil
}

func (fetcher *jwksFetcher) Fetch(ctx context.Context) (map[string]*rsa.PublicKey, error) {
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, fetcher.url, nil)
	if err != nil {
		return nil, err
	}
	response, err := fetcher.client.Do(request)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("JWKS endpoint returned HTTP %d", response.StatusCode)
	}
	body, err := io.ReadAll(io.LimitReader(response.Body, maxJWKSSize+1))
	if err != nil {
		return nil, err
	}
	if len(body) > maxJWKSSize {
		return nil, errors.New("JWKS response is too large")
	}
	var document jwksDocument
	if err := json.Unmarshal(body, &document); err != nil {
		return nil, err
	}
	keys := make(map[string]*rsa.PublicKey, len(document.Keys))
	for _, item := range document.Keys {
		key, err := item.publicKey()
		if err == nil {
			keys[item.KeyID] = key
		}
	}
	return keys, nil
}

func (item jwk) publicKey() (*rsa.PublicKey, error) {
	if item.KeyID == "" || item.KeyType != "RSA" || (item.Algorithm != "" && item.Algorithm != "RS256") || (item.Use != "" && item.Use != "sig") {
		return nil, errors.New("unsupported JWK")
	}
	if item.Cert != "" {
		certificatePEM := []byte(item.Cert)
		if block, _ := pem.Decode(certificatePEM); block != nil {
			certificatePEM = block.Bytes
		} else if decoded, err := base64.StdEncoding.DecodeString(item.Cert); err == nil {
			certificatePEM = decoded
		}
		certificate, err := x509.ParseCertificate(certificatePEM)
		if err != nil {
			return nil, err
		}
		key, ok := certificate.PublicKey.(*rsa.PublicKey)
		if !ok || key.N.BitLen() < 2048 {
			return nil, errors.New("unsupported or weak RSA certificate key")
		}
		return key, nil
	}
	if len(item.CertChain) > 0 {
		certDER, err := base64.StdEncoding.DecodeString(item.CertChain[0])
		if err != nil {
			return nil, err
		}
		certificate, err := x509.ParseCertificate(certDER)
		if err != nil {
			return nil, err
		}
		key, ok := certificate.PublicKey.(*rsa.PublicKey)
		if !ok || key.N.BitLen() < 2048 {
			return nil, errors.New("unsupported or weak RSA certificate key")
		}
		return key, nil
	}
	modulus, err := base64.RawURLEncoding.DecodeString(item.Modulus)
	if err != nil {
		return nil, err
	}
	exponent, err := base64.RawURLEncoding.DecodeString(item.Exponent)
	if err != nil || len(exponent) == 0 || len(exponent) > 4 {
		return nil, errors.New("invalid RSA exponent")
	}
	exponentValue := new(big.Int).SetBytes(exponent)
	if !exponentValue.IsInt64() {
		return nil, errors.New("invalid RSA exponent")
	}
	key := &rsa.PublicKey{N: new(big.Int).SetBytes(modulus), E: int(exponentValue.Int64())}
	if key.N.BitLen() < 2048 || key.E < 3 || key.E%2 == 0 {
		return nil, errors.New("unsupported or weak RSA key")
	}
	return key, nil
}

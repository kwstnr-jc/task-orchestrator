package auth

import (
	"crypto/rsa"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"math/big"
	"net/http"
	"sync"
	"time"
)

type jwksKey struct {
	Kty string `json:"kty"`
	Kid string `json:"kid"`
	Use string `json:"use"`
	N   string `json:"n"`
	E   string `json:"e"`
	Alg string `json:"alg"`
}

type jwksResponse struct {
	Keys []jwksKey `json:"keys"`
}

var (
	jwksCache     *jwksResponse
	jwksCacheMu   sync.RWMutex
	jwksCacheTime time.Time
	jwksCacheTTL  = 1 * time.Hour
)

func fetchJWKS(jwksURL string, kid string) (*rsa.PublicKey, error) {
	jwksCacheMu.RLock()
	cached := jwksCache
	cacheValid := time.Since(jwksCacheTime) < jwksCacheTTL
	jwksCacheMu.RUnlock()

	if cached == nil || !cacheValid {
		resp, err := http.Get(jwksURL)
		if err != nil {
			return nil, fmt.Errorf("fetch JWKS: %w", err)
		}
		defer func() { _ = resp.Body.Close() }()

		var jwks jwksResponse
		if err := json.NewDecoder(resp.Body).Decode(&jwks); err != nil {
			return nil, fmt.Errorf("decode JWKS: %w", err)
		}

		jwksCacheMu.Lock()
		jwksCache = &jwks
		jwksCacheTime = time.Now()
		jwksCacheMu.Unlock()

		cached = &jwks
	}

	for _, key := range cached.Keys {
		if key.Kid == kid && key.Kty == "RSA" {
			return parseRSAPublicKey(key)
		}
	}
	return nil, fmt.Errorf("key %s not found in JWKS", kid)
}

func parseRSAPublicKey(key jwksKey) (*rsa.PublicKey, error) {
	nBytes, err := base64.RawURLEncoding.DecodeString(key.N)
	if err != nil {
		return nil, fmt.Errorf("decode N: %w", err)
	}
	eBytes, err := base64.RawURLEncoding.DecodeString(key.E)
	if err != nil {
		return nil, fmt.Errorf("decode E: %w", err)
	}

	n := new(big.Int).SetBytes(nBytes)
	e := new(big.Int).SetBytes(eBytes)

	return &rsa.PublicKey{
		N: n,
		E: int(e.Int64()),
	}, nil
}

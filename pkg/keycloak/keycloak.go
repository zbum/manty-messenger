package keycloak

import (
	"context"
	"crypto/rsa"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"math/big"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/golang-jwt/jwt/v5"

	"Mmessenger/internal/config"
)

var (
	ErrInvalidToken = errors.New("invalid token")
	ErrExpiredToken = errors.New("token has expired")
	ErrNoPublicKey  = errors.New("no public key found")
)

type Claims struct {
	Subject           string `json:"sub"`
	Email             string `json:"email"`
	EmailVerified     bool   `json:"email_verified"`
	PreferredUsername string `json:"preferred_username"`
	Name              string `json:"name"`
	GivenName         string `json:"given_name"`
	FamilyName        string `json:"family_name"`
	jwt.RegisteredClaims
}

type JWKS struct {
	Keys []JWK `json:"keys"`
}

type JWK struct {
	Kid string `json:"kid"`
	Kty string `json:"kty"`
	Alg string `json:"alg"`
	Use string `json:"use"`
	N   string `json:"n"`
	E   string `json:"e"`
}

type Service struct {
	url      string
	realm    string
	clientID string
	jwksURL  string

	mu         sync.RWMutex
	publicKeys map[string]*rsa.PublicKey
	lastFetch  time.Time
	cacheTTL   time.Duration
}

func NewService(cfg *config.KeycloakConfig) *Service {
	jwksURL := fmt.Sprintf("%s/realms/%s/protocol/openid-connect/certs", cfg.URL, cfg.Realm)
	log.Printf("[Keycloak] NewService: URL=%s, Realm=%s, ClientID=%s, JWKS_URL=%s",
		cfg.URL, cfg.Realm, cfg.ClientID, jwksURL)

	return &Service{
		url:        cfg.URL,
		realm:      cfg.Realm,
		clientID:   cfg.ClientID,
		jwksURL:    jwksURL,
		publicKeys: make(map[string]*rsa.PublicKey),
		cacheTTL:   1 * time.Hour,
	}
}

func (s *Service) ValidateToken(tokenString string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}

		kid, ok := token.Header["kid"].(string)
		if !ok {
			return nil, ErrInvalidToken
		}

		key, err := s.getPublicKey(kid)
		if err != nil {
			return nil, err
		}

		return key, nil
	})

	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			return nil, ErrExpiredToken
		}
		return nil, fmt.Errorf("%w: %v", ErrInvalidToken, err)
	}

	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, ErrInvalidToken
	}

	return claims, nil
}

func (s *Service) getPublicKey(kid string) (*rsa.PublicKey, error) {
	s.mu.RLock()
	key, exists := s.publicKeys[kid]
	needRefresh := time.Since(s.lastFetch) > s.cacheTTL
	s.mu.RUnlock()

	if exists && !needRefresh {
		return key, nil
	}

	if err := s.fetchJWKS(); err != nil {
		if exists {
			return key, nil
		}
		return nil, err
	}

	s.mu.RLock()
	key, exists = s.publicKeys[kid]
	s.mu.RUnlock()

	if !exists {
		return nil, ErrNoPublicKey
	}

	return key, nil
}

func (s *Service) fetchJWKS() error {
	log.Printf("[Keycloak] Fetching JWKS from %s", s.jwksURL)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, s.jwksURL, nil)
	if err != nil {
		log.Printf("[Keycloak] Failed to create JWKS request: %v", err)
		return fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Printf("[Keycloak] Failed to fetch JWKS: %v", err)
		return fmt.Errorf("failed to fetch JWKS: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		log.Printf("[Keycloak] JWKS fetch returned status %d", resp.StatusCode)
		return fmt.Errorf("failed to fetch JWKS: status %d", resp.StatusCode)
	}

	log.Printf("[Keycloak] JWKS fetched successfully")

	var jwks JWKS
	if err := json.NewDecoder(resp.Body).Decode(&jwks); err != nil {
		return fmt.Errorf("failed to decode JWKS: %w", err)
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	for _, jwk := range jwks.Keys {
		if jwk.Kty != "RSA" || jwk.Use != "sig" {
			continue
		}

		key, err := jwkToRSAPublicKey(&jwk)
		if err != nil {
			continue
		}

		s.publicKeys[jwk.Kid] = key
	}

	s.lastFetch = time.Now()
	return nil
}

func jwkToRSAPublicKey(jwk *JWK) (*rsa.PublicKey, error) {
	nBytes, err := base64.RawURLEncoding.DecodeString(jwk.N)
	if err != nil {
		return nil, fmt.Errorf("failed to decode N: %w", err)
	}

	eBytes, err := base64.RawURLEncoding.DecodeString(jwk.E)
	if err != nil {
		return nil, fmt.Errorf("failed to decode E: %w", err)
	}

	n := new(big.Int).SetBytes(nBytes)

	var e int
	for _, b := range eBytes {
		e = e<<8 + int(b)
	}

	return &rsa.PublicKey{
		N: n,
		E: e,
	}, nil
}

func (s *Service) GetUserInfo(tokenString string) (*Claims, error) {
	tokenString = strings.TrimPrefix(tokenString, "Bearer ")
	return s.ValidateToken(tokenString)
}

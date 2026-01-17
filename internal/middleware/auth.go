package middleware

import (
	"context"
	"log"
	"net/http"
	"strings"

	"Mmessenger/pkg/keycloak"
)

type contextKey string

const UserContextKey contextKey = "user"

type UserClaims struct {
	UserID            uint64
	KeycloakID        string
	Email             string
	Username          string
	PreferredUsername string
}

type AuthMiddleware struct {
	keycloakService *keycloak.Service
	userLookup      UserLookupFunc
}

type UserLookupFunc func(ctx context.Context, keycloakClaims *keycloak.Claims) (*UserClaims, error)

func NewAuthMiddleware(keycloakService *keycloak.Service, userLookup UserLookupFunc) *AuthMiddleware {
	return &AuthMiddleware{
		keycloakService: keycloakService,
		userLookup:      userLookup,
	}
}

func (m *AuthMiddleware) Authenticate(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			http.Error(w, "Authorization header required", http.StatusUnauthorized)
			return
		}

		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			http.Error(w, "Invalid authorization header format", http.StatusUnauthorized)
			return
		}

		keycloakClaims, err := m.keycloakService.ValidateToken(parts[1])
		if err != nil {
			log.Printf("[AuthMiddleware] Token validation failed: %v", err)
			http.Error(w, "Invalid or expired token", http.StatusUnauthorized)
			return
		}

		log.Printf("[AuthMiddleware] Token valid, KeycloakID: %s, Email: %s, Username: %s",
			keycloakClaims.Subject, keycloakClaims.Email, keycloakClaims.PreferredUsername)

		userClaims, err := m.userLookup(r.Context(), keycloakClaims)
		if err != nil {
			log.Printf("[AuthMiddleware] User lookup failed: %v, KeycloakID: %s", err, keycloakClaims.Subject)
			http.Error(w, "Failed to lookup user", http.StatusInternalServerError)
			return
		}

		log.Printf("[AuthMiddleware] User found, UserID: %d", userClaims.UserID)

		ctx := context.WithValue(r.Context(), UserContextKey, userClaims)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func GetUserFromContext(ctx context.Context) *UserClaims {
	claims, ok := ctx.Value(UserContextKey).(*UserClaims)
	if !ok {
		return nil
	}
	return claims
}

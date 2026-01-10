package middleware

import (
	"context"
	"net/http"
	"strings"

	"Mmessenger/pkg/jwt"
)

type contextKey string

const UserContextKey contextKey = "user"

type AuthMiddleware struct {
	jwtService *jwt.Service
}

func NewAuthMiddleware(jwtService *jwt.Service) *AuthMiddleware {
	return &AuthMiddleware{jwtService: jwtService}
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

		claims, err := m.jwtService.ValidateToken(parts[1])
		if err != nil {
			http.Error(w, "Invalid or expired token", http.StatusUnauthorized)
			return
		}

		ctx := context.WithValue(r.Context(), UserContextKey, claims)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func GetUserFromContext(ctx context.Context) *jwt.Claims {
	claims, ok := ctx.Value(UserContextKey).(*jwt.Claims)
	if !ok {
		return nil
	}
	return claims
}

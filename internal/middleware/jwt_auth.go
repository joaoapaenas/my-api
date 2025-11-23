package middleware

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/golang-jwt/jwt/v5"
	"github.com/joaoapaenas/my-api/internal/config"
)

type JWTAuthMiddleware struct {
	cfg *config.Config
}

func NewJWTAuthMiddleware(cfg *config.Config) *JWTAuthMiddleware {
	return &JWTAuthMiddleware{cfg: cfg}
}

func (m *JWTAuthMiddleware) Protected(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			http.Error(w, "Missing Authorization Header", http.StatusUnauthorized)
			return
		}

		// Header format: "Bearer <token>"
		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			http.Error(w, "Invalid Authorization Header Format", http.StatusUnauthorized)
			return
		}

		tokenString := parts[1]

		// Parse and Validate
		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
			}
			return []byte(m.cfg.JWTSecret), nil
		})

		if err != nil || !token.Valid {
			http.Error(w, "Invalid or Expired Token", http.StatusUnauthorized)
			return
		}

		if claims, ok := token.Claims.(jwt.MapClaims); ok {
			ctx := context.WithValue(r.Context(), "userID", claims["sub"])
			ctx = context.WithValue(ctx, "userEmail", claims["email"]) // ADD THIS
			next.ServeHTTP(w, r.WithContext(ctx))
		}
	})
}

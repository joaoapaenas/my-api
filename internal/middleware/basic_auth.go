package middleware

import (
	"log/slog"
	"net/http"

	"github.com/joaoapaenas/my-api/internal/service"
	"golang.org/x/crypto/bcrypt"
)

type BasicAuthMiddleware struct {
	userService service.UserService
}

func NewBasicAuthMiddleware(userService service.UserService) *BasicAuthMiddleware {
	return &BasicAuthMiddleware{userService: userService}
}

func (m *BasicAuthMiddleware) BasicAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		email, password, ok := r.BasicAuth()
		if !ok {
			w.Header().Set("WWW-Authenticate", `Basic realm="Restricted"`)
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		user, err := m.userService.GetUserByEmail(r.Context(), email)
		if err != nil {
			// User not found or DB error
			slog.Warn("Basic Auth failed: user not found", "email", email, "error", err)
			w.Header().Set("WWW-Authenticate", `Basic realm="Restricted"`)
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		// Verify password
		if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
			slog.Warn("Basic Auth failed: invalid password", "email", email)
			w.Header().Set("WWW-Authenticate", `Basic realm="Restricted"`)
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		next.ServeHTTP(w, r)
	})
}

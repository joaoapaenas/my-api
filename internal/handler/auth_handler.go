package handler

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/joaoapaenas/my-api/internal/config"
	"github.com/joaoapaenas/my-api/internal/service"
	"golang.org/x/crypto/bcrypt"
)

type AuthHandler struct {
	userService service.UserService
	cfg         *config.Config
}

func NewAuthHandler(userService service.UserService, cfg *config.Config) *AuthHandler {
	return &AuthHandler{userService: userService, cfg: cfg}
}

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type LoginResponse struct {
	Token string `json:"token"`
}

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	// 1. Find User
	user, err := h.userService.GetUserByEmail(r.Context(), req.Email)
	if err != nil {
		// Use generic message for security
		http.Error(w, "Invalid credentials", http.StatusUnauthorized)
		return
	}

	// 2. Check Password
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
		http.Error(w, "Invalid credentials", http.StatusUnauthorized)
		return
	}

	// 3. Generate JWT
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub":   user.ID,
		"email": user.Email, // Ensure this is here
		"exp":   time.Now().Add(time.Hour * 24).Unix(),
	})

	tokenString, err := token.SignedString([]byte(h.cfg.JWTSecret))
	if err != nil {
		http.Error(w, "Failed to generate token", http.StatusInternalServerError)
		return
	}

	// 4. Return Token
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(LoginResponse{Token: tokenString})
}

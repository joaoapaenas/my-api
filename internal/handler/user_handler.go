package handler

import (
	"database/sql"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-playground/validator/v10"
	"github.com/joaoapaenas/my-api/internal/service"
)

type UserHandler struct {
	svc      service.UserService
	validate *validator.Validate
}

func NewUserHandler(svc service.UserService) *UserHandler {
	return &UserHandler{svc: svc, validate: validator.New()}
}

// --- Structs ---

type CreateUserRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Name     string `json:"name" validate:"required,min=2"`
	Password string `json:"password" validate:"required,min=6"`
}

type ChangePasswordRequest struct {
	OldPassword string `json:"old_password" validate:"required"`
	NewPassword string `json:"new_password" validate:"required,min=6"`
}

// --- Handlers ---

// CreateUser godoc
// @Summary Create a new user
// @Tags users
// @Accept json
// @Produce json
// @Param input body CreateUserRequest true "User info"
// @Success 201 {object} database.User
// @Router /users [post]
func (h *UserHandler) CreateUser(w http.ResponseWriter, r *http.Request) {
	var req CreateUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.respondWithError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}

	if err := h.validate.Struct(req); err != nil {
		h.respondWithJSON(w, http.StatusBadRequest, map[string]interface{}{
			"error":   "Validation failed",
			"details": formatValidationErrors(err),
		})
		return
	}

	user, err := h.svc.CreateUser(r.Context(), req.Email, req.Name, req.Password)
	if err != nil {
		if errors.Is(err, service.ErrEmailTaken) {
			h.respondWithError(w, http.StatusConflict, "Email already exists")
			return
		}
		slog.Error("Failed to create user", "error", err)
		h.respondWithError(w, http.StatusInternalServerError, "Internal server error")
		return
	}

	h.respondWithJSON(w, http.StatusCreated, user)
}

// GetUser godoc
// @Summary Get user by Email
// @Tags users
// @Param email path string true "User Email"
// @Success 200 {object} database.User
// @Router /users/{email} [get]
func (h *UserHandler) GetUser(w http.ResponseWriter, r *http.Request) {
	email := chi.URLParam(r, "email")

	user, err := h.svc.GetUserByEmail(r.Context(), email)
	if err != nil {
		if err == sql.ErrNoRows || errors.Is(err, service.ErrUserNotFound) {
			http.Error(w, "User not found", http.StatusNotFound)
			return
		}
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(user)
}

// ChangePassword godoc
// @Summary Change user password
// @Tags users
// @Accept json
// @Produce json
// @Param input body ChangePasswordRequest true "Password info"
// @Success 200 {object} handler.MessageResponse
// @Router /users/password [put]
func (h *UserHandler) ChangePassword(w http.ResponseWriter, r *http.Request) {
	// Extract email from context (set by JWT middleware)
	// We cast to string safely
	emailVal := r.Context().Value("userEmail")
	email, ok := emailVal.(string)

	if !ok || email == "" {
		h.respondWithError(w, http.StatusUnauthorized, "User context invalid")
		return
	}

	var req ChangePasswordRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.respondWithError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}

	if err := h.validate.Struct(req); err != nil {
		h.respondWithJSON(w, http.StatusBadRequest, map[string]interface{}{
			"error":   "Validation failed",
			"details": formatValidationErrors(err),
		})
		return
	}

	err := h.svc.UpdatePassword(r.Context(), email, req.OldPassword, req.NewPassword)
	if err != nil {
		h.respondWithError(w, http.StatusUnauthorized, "Failed to update password. Check old password.")
		return
	}

	h.respondWithJSON(w, http.StatusOK, map[string]string{"message": "Password updated successfully"})
}

// --- Helpers ---

func (h *UserHandler) respondWithJSON(w http.ResponseWriter, code int, payload interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(payload)
}

func (h *UserHandler) respondWithError(w http.ResponseWriter, code int, message string) {
	h.respondWithJSON(w, code, map[string]string{"error": message})
}

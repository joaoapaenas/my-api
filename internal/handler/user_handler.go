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

type CreateUserRequest struct {
	// required: cannot be empty
	// email: must be a valid email format
	Email string `json:"email" validate:"required,email"`

	// min=2: must be at least 2 chars
	Name string `json:"name" validate:"required,min=2"`
}

// CreateUser godoc
// @Summary Create a new user
// @Description Create a user with email and name
// @Tags users
// @Accept json
// @Produce json
// @Param input body CreateUserRequest true "User info"
// @Success 201 {object} database.User
// @Failure 400 {string} string "Invalid request"
// @Failure 500 {string} string "Internal server error"
// @Router /users [post]
func (h *UserHandler) CreateUser(w http.ResponseWriter, r *http.Request) {
	// 1. Decode & Basic Validation
	var req CreateUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.respondWithError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}

	if err := h.validate.Struct(req); err != nil {
		// Return friendly validation errors
		validationErrors := formatValidationErrors(err)
		h.respondWithJSON(w, http.StatusBadRequest, map[string]interface{}{
			"error":   "Validation failed",
			"details": validationErrors,
		})
		return
	}

	// 2. Call Service (Business Logic)
	// Notice: We don't generate UUIDs here anymore.
	user, err := h.svc.CreateUser(r.Context(), req.Email, req.Name)
	if err != nil {
		// Check for specific domain errors if you defined them
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
// @Failure 404 {string} string "User not found"
// @Router /users/{email} [get]
func (h *UserHandler) GetUser(w http.ResponseWriter, r *http.Request) {
	email := chi.URLParam(r, "email")

	user, err := h.svc.GetUserByEmail(r.Context(), email)
	if err != nil {
		// Handle "Not Found" specifically
		if err == sql.ErrNoRows || errors.Is(err, service.ErrUserNotFound) {
			http.Error(w, "User not found", http.StatusNotFound)
			return
		}
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(user)
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

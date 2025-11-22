package handler

import "github.com/go-playground/validator/v10"

// formatValidationErrors formats validator errors into a readable map
func formatValidationErrors(err error) map[string]string {
	errors := make(map[string]string)
	if validationErrors, ok := err.(validator.ValidationErrors); ok {
		for _, e := range validationErrors {
			errors[e.Field()] = e.Tag()
		}
	}
	return errors
}

// Response DTOs for Swagger documentation
type SubjectResponse struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	ColorHex  string `json:"color_hex,omitempty"`
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
	DeletedAt string `json:"deleted_at,omitempty"`
}

type TopicResponse struct {
	ID        string `json:"id"`
	SubjectID string `json:"subject_id"`
	Name      string `json:"name"`
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
	DeletedAt string `json:"deleted_at,omitempty"`
}

type StudyCycleResponse struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
	IsActive    int    `json:"is_active"`
	CreatedAt   string `json:"created_at"`
	UpdatedAt   string `json:"updated_at"`
	DeletedAt   string `json:"deleted_at,omitempty"`
}

type CycleItemResponse struct {
	ID                     string `json:"id"`
	CycleID                string `json:"cycle_id"`
	SubjectID              string `json:"subject_id"`
	OrderIndex             int    `json:"order_index"`
	PlannedDurationMinutes int    `json:"planned_duration_minutes,omitempty"`
	CreatedAt              string `json:"created_at"`
	UpdatedAt              string `json:"updated_at"`
}

type StudySessionResponse struct {
	ID                   string `json:"id"`
	SubjectID            string `json:"subject_id"`
	CycleItemID          string `json:"cycle_item_id,omitempty"`
	StartedAt            string `json:"started_at"`
	FinishedAt           string `json:"finished_at,omitempty"`
	GrossDurationSeconds int    `json:"gross_duration_seconds,omitempty"`
	NetDurationSeconds   int    `json:"net_duration_seconds,omitempty"`
	Notes                string `json:"notes,omitempty"`
}

type SessionPauseResponse struct {
	ID        string `json:"id"`
	SessionID string `json:"session_id"`
	StartedAt string `json:"started_at"`
	EndedAt   string `json:"ended_at,omitempty"`
}

type ExerciseLogResponse struct {
	ID             string `json:"id"`
	SessionID      string `json:"session_id,omitempty"`
	SubjectID      string `json:"subject_id"`
	TopicID        string `json:"topic_id,omitempty"`
	QuestionsCount int    `json:"questions_count"`
	CorrectCount   int    `json:"correct_count"`
	CreatedAt      string `json:"created_at"`
}

type MessageResponse struct {
	Message string `json:"message"`
}

type ErrorResponse struct {
	Error string `json:"error"`
}

type ValidationErrorResponse struct {
	Error   string            `json:"error"`
	Details map[string]string `json:"details"`
}

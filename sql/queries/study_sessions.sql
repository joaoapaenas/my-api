-- name: CreateStudySession :one
INSERT INTO study_sessions (id, subject_id, cycle_item_id, started_at)
VALUES (?, ?, ?, ?)
RETURNING *;

-- name: UpdateSessionDuration :exec
UPDATE study_sessions
SET finished_at = ?, gross_duration_seconds = ?, net_duration_seconds = ?, notes = ?
WHERE id = ?;

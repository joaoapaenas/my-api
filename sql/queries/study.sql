-- name: CreateSubject :one
INSERT INTO subjects (id, name, color_hex)
VALUES (?, ?, ?)
RETURNING *;

-- name: ListSubjects :many
SELECT * FROM subjects
WHERE deleted_at IS NULL
ORDER BY name;

-- name: CreateTopic :one
INSERT INTO topics (id, subject_id, name)
VALUES (?, ?, ?)
RETURNING *;

-- name: ListTopicsBySubject :many
SELECT * FROM topics
WHERE subject_id = ? AND deleted_at IS NULL
ORDER BY name;

-- name: CreateStudyCycle :one
INSERT INTO study_cycles (id, name, description, is_active)
VALUES (?, ?, ?, ?)
RETURNING *;

-- name: GetActiveStudyCycle :one
SELECT * FROM study_cycles
WHERE is_active = 1 AND deleted_at IS NULL
LIMIT 1;

-- name: CreateCycleItem :one
INSERT INTO cycle_items (id, cycle_id, subject_id, order_index, planned_duration_minutes)
VALUES (?, ?, ?, ?, ?)
RETURNING *;

-- name: ListCycleItems :many
SELECT * FROM cycle_items
WHERE cycle_id = ?
ORDER BY order_index;

-- name: CreateStudySession :one
INSERT INTO study_sessions (id, subject_id, cycle_item_id, started_at)
VALUES (?, ?, ?, ?)
RETURNING *;

-- name: UpdateSessionDuration :exec
UPDATE study_sessions
SET finished_at = ?, gross_duration_seconds = ?, net_duration_seconds = ?, notes = ?
WHERE id = ?;

-- name: CreateSessionPause :one
INSERT INTO session_pauses (id, session_id, started_at)
VALUES (?, ?, ?)
RETURNING *;

-- name: EndSessionPause :exec
UPDATE session_pauses
SET ended_at = ?
WHERE id = ?;

-- name: CreateExerciseLog :one
INSERT INTO exercise_logs (id, session_id, subject_id, topic_id, questions_count, correct_count)
VALUES (?, ?, ?, ?, ?, ?)
RETURNING *;

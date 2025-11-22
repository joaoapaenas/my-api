-- name: CreateStudySession :one
INSERT INTO study_sessions (id, subject_id, cycle_item_id, started_at)
VALUES (?, ?, ?, ?)
RETURNING *;

-- name: UpdateSessionDuration :exec
UPDATE study_sessions
SET finished_at = ?, gross_duration_seconds = ?, net_duration_seconds = ?, notes = ?
WHERE id = ?;

-- name: GetStudySession :one
SELECT * FROM study_sessions
WHERE id = ?;

-- name: DeleteStudySession :exec
DELETE FROM study_sessions
WHERE id = ?;

-- name: GetOpenSession :one
SELECT 
    ss.id,
    ss.subject_id,
    ss.cycle_item_id,
    ss.started_at,
    s.name AS subject_name,
    s.color_hex
FROM study_sessions ss
JOIN subjects s ON ss.subject_id = s.id
WHERE ss.finished_at IS NULL
ORDER BY ss.started_at DESC
LIMIT 1;

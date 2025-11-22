-- name: CreateSessionPause :one
INSERT INTO session_pauses (id, session_id, started_at)
VALUES (?, ?, ?)
RETURNING *;

-- name: EndSessionPause :exec
UPDATE session_pauses
SET ended_at = ?
WHERE id = ?;

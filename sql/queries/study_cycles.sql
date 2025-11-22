-- name: CreateStudyCycle :one
INSERT INTO study_cycles (id, name, description, is_active)
VALUES (?, ?, ?, ?)
RETURNING *;

-- name: GetActiveStudyCycle :one
SELECT * FROM study_cycles
WHERE is_active = 1 AND deleted_at IS NULL
LIMIT 1;

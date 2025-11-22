-- name: CreateStudyCycle :one
INSERT INTO study_cycles (id, name, description, is_active)
VALUES (?, ?, ?, ?)
RETURNING *;

-- name: GetActiveStudyCycle :one
SELECT * FROM study_cycles
WHERE is_active = 1 AND deleted_at IS NULL
LIMIT 1;

-- name: GetStudyCycle :one
SELECT * FROM study_cycles
WHERE id = ? AND deleted_at IS NULL;

-- name: UpdateStudyCycle :exec
UPDATE study_cycles
SET name = ?, description = ?, is_active = ?, updated_at = datetime('now')
WHERE id = ? AND deleted_at IS NULL;

-- name: DeleteStudyCycle :exec
UPDATE study_cycles
SET deleted_at = datetime('now')
WHERE id = ?;

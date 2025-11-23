-- name: CreateSubject :one
INSERT INTO subjects (id, user_id, name, color_hex)
VALUES (?, ?, ?, ?)
RETURNING *;

-- name: ListSubjects :many
SELECT * FROM subjects
WHERE user_id = ? AND deleted_at IS NULL
ORDER BY name;

-- name: GetSubject :one
SELECT * FROM subjects
WHERE id = ? AND user_id = ? AND deleted_at IS NULL;

-- name: UpdateSubject :exec
UPDATE subjects
SET name = ?, color_hex = ?, updated_at = datetime('now')
WHERE id = ? AND user_id = ? AND deleted_at IS NULL;

-- name: DeleteSubject :exec
UPDATE subjects
SET deleted_at = datetime('now')
WHERE id = ? AND user_id = ?;

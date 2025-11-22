-- name: CreateSubject :one
INSERT INTO subjects (id, name, color_hex)
VALUES (?, ?, ?)
RETURNING *;

-- name: ListSubjects :many
SELECT * FROM subjects
WHERE deleted_at IS NULL
ORDER BY name;

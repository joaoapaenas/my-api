-- name: CreateTopic :one
INSERT INTO topics (id, subject_id, name)
VALUES (?, ?, ?)
RETURNING *;

-- name: ListTopicsBySubject :many
SELECT * FROM topics
WHERE subject_id = ? AND deleted_at IS NULL
ORDER BY name;

-- name: GetTopic :one
SELECT * FROM topics
WHERE id = ? AND deleted_at IS NULL;

-- name: UpdateTopic :exec
UPDATE topics
SET name = ?, updated_at = datetime('now')
WHERE id = ? AND deleted_at IS NULL;

-- name: DeleteTopic :exec
UPDATE topics
SET deleted_at = datetime('now')
WHERE id = ?;

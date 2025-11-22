-- name: CreateTopic :one
INSERT INTO topics (id, subject_id, name)
VALUES (?, ?, ?)
RETURNING *;

-- name: ListTopicsBySubject :many
SELECT * FROM topics
WHERE subject_id = ? AND deleted_at IS NULL
ORDER BY name;

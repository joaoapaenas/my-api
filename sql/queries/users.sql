-- name: CreateUser :one
INSERT INTO users (id, email, name, password_hash)
VALUES (?, ?, ?, ?)
RETURNING *;

-- name: GetUserByEmail :one
SELECT * FROM users
WHERE email = ? LIMIT 1;

-- name: UpdateUserPassword :exec
UPDATE users
SET password_hash = ?
WHERE id = ?;

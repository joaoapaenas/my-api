-- name: CreateUser :one
INSERT INTO users (id, email, name)
VALUES (?, ?, ?)
RETURNING *;

-- name: GetUserByEmail :one
SELECT * FROM users
WHERE email = ? LIMIT 1;

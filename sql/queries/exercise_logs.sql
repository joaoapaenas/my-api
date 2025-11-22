-- name: CreateExerciseLog :one
INSERT INTO exercise_logs (id, session_id, subject_id, topic_id, questions_count, correct_count)
VALUES (?, ?, ?, ?, ?, ?)
RETURNING *;

-- name: GetExerciseLog :one
SELECT * FROM exercise_logs
WHERE id = ?;

-- name: DeleteExerciseLog :exec
DELETE FROM exercise_logs
WHERE id = ?;

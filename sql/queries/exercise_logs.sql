-- name: CreateExerciseLog :one
INSERT INTO exercise_logs (id, session_id, subject_id, topic_id, questions_count, correct_count)
VALUES (?, ?, ?, ?, ?, ?)
RETURNING *;

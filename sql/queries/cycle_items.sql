-- name: CreateCycleItem :one
INSERT INTO cycle_items (id, cycle_id, subject_id, order_index, planned_duration_minutes)
VALUES (?, ?, ?, ?, ?)
RETURNING *;

-- name: ListCycleItems :many
SELECT * FROM cycle_items
WHERE cycle_id = ?
ORDER BY order_index;

-- name: GetCycleItem :one
SELECT * FROM cycle_items
WHERE id = ?;

-- name: UpdateCycleItem :exec
UPDATE cycle_items
SET subject_id = ?, order_index = ?, planned_duration_minutes = ?, updated_at = datetime('now')
WHERE id = ?;

-- name: DeleteCycleItem :exec
DELETE FROM cycle_items
WHERE id = ?;

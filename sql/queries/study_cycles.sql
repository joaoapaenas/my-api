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

-- name: GetActiveCycleWithItems :many
SELECT 
    ci.id AS cycle_item_id,
    ci.order_index,
    ci.planned_duration_minutes,
    s.id AS subject_id,
    s.name AS subject_name,
    s.color_hex
FROM cycle_items ci
JOIN study_cycles sc ON ci.cycle_id = sc.id
JOIN subjects s ON ci.subject_id = s.id
WHERE sc.is_active = 1 
  AND sc.deleted_at IS NULL
ORDER BY ci.order_index ASC;

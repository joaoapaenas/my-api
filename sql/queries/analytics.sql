-- Analytics Queries for Study App

-- name: GetTimeReportBySubject :many
SELECT 
    s.id AS subject_id,
    s.name AS subject_name,
    s.color_hex,
    COUNT(ss.id) AS sessions_count,
    ROUND(COALESCE(SUM(ss.net_duration_seconds), 0) / 3600.0, 2) AS total_hours_net
FROM subjects s
LEFT JOIN study_sessions ss ON s.id = ss.subject_id 
    AND ss.finished_at IS NOT NULL
    AND (? = '' OR ss.started_at >= ?)
    AND (? = '' OR ss.started_at <= ?)
WHERE s.deleted_at IS NULL
GROUP BY s.id, s.name, s.color_hex
HAVING sessions_count > 0
ORDER BY total_hours_net DESC;

-- name: GetAccuracyBySubject :many
SELECT 
    s.id AS subject_id,
    s.name AS subject_name,
    s.color_hex,
    SUM(el.questions_count) AS total_questions,
    SUM(el.correct_count) AS total_correct,
    ROUND(
        (SUM(el.correct_count) * 100.0) / NULLIF(SUM(el.questions_count), 0), 
        2
    ) AS accuracy_percentage
FROM subjects s
LEFT JOIN exercise_logs el ON s.id = el.subject_id
WHERE s.deleted_at IS NULL
GROUP BY s.id, s.name, s.color_hex
HAVING total_questions > 0
ORDER BY accuracy_percentage ASC;

-- name: GetAccuracyByTopic :many
SELECT 
    t.id AS topic_id,
    t.name AS topic_name,
    SUM(el.questions_count) AS total_questions,
    SUM(el.correct_count) AS total_correct,
    ROUND(
        (SUM(el.correct_count) * 100.0) / NULLIF(SUM(el.questions_count), 0), 
        2
    ) AS accuracy_percentage
FROM topics t
LEFT JOIN exercise_logs el ON t.id = el.topic_id
WHERE t.subject_id = ?
  AND t.deleted_at IS NULL
GROUP BY t.id, t.name
HAVING total_questions > 0
ORDER BY accuracy_percentage ASC;

-- name: GetActivityHeatmap :many
SELECT 
    strftime('%Y-%m-%d', started_at) AS study_date,
    COUNT(DISTINCT id) AS sessions_count,
    COALESCE(SUM(net_duration_seconds), 0) AS total_seconds
FROM study_sessions
WHERE finished_at IS NOT NULL
  AND datetime(started_at) >= datetime('now', '-' || CAST(? AS TEXT) || ' days')
GROUP BY study_date
ORDER BY study_date DESC;

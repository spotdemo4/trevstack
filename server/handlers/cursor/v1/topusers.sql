SELECT
    u.username,
    COUNT(*) AS count,
    COUNT(DISTINCT c.stream_id) AS streams,
    COALESCE(SUM(c.distance), 0.0) AS distance
FROM cursors AS c
INNER JOIN cursor_streams AS s ON c.stream_id = s.id
INNER JOIN users AS u ON s.user_id = u.id
WHERE
    (? IS NULL OR c.timestamp >= ?)
    AND (? IS NULL OR c.timestamp <= ?)
GROUP BY u.id, u.username
ORDER BY count DESC, distance DESC
LIMIT ?

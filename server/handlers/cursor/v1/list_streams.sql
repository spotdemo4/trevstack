SELECT
    s.id,
    u.username,
    s.source,
    MIN(c.timestamp) AS first_timestamp,
    MAX(c.timestamp) AS last_timestamp,
    COUNT(*) AS count,
    SUM(c.distance) AS distance
FROM cursor_streams AS s
INNER JOIN users AS u ON s.user_id = u.id
INNER JOIN cursors AS c ON s.id = c.stream_id
WHERE
    (? IS NULL OR u.username LIKE '%' || ? || '%')
    AND (? IS NULL OR s.source = ?)
GROUP BY s.id, u.username, s.source
HAVING
    (? IS NULL OR MIN(c.timestamp) >= ?)
    AND (? IS NULL OR MIN(c.timestamp) <= ?)
ORDER BY s.id DESC

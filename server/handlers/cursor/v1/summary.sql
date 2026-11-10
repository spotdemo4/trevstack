SELECT
    COUNT(*) AS count,
    COUNT(DISTINCT c.stream_id) AS streams,
    COUNT(DISTINCT s.user_id) AS users,
    COALESCE(SUM(c.distance), 0.0) AS distance
FROM cursors AS c
INNER JOIN cursor_streams AS s ON c.stream_id = s.id
WHERE
    (? IS NULL OR c.timestamp >= ?)
    AND (? IS NULL OR c.timestamp <= ?)

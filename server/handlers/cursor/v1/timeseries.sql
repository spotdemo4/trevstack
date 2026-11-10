SELECT
    %s AS bucket,
    COUNT(*) AS count,
    COUNT(DISTINCT c.stream_id) AS streams,
    COUNT(DISTINCT s.user_id) AS users
FROM cursors AS c
INNER JOIN cursor_streams AS s ON c.stream_id = s.id
WHERE
    (? IS NULL OR c.timestamp >= ?)
    AND (? IS NULL OR c.timestamp <= ?)
GROUP BY bucket
ORDER BY bucket ASC

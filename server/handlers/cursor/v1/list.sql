SELECT
    c.stream_id,
    u.username,
    s.source,
    c.x,
    c.y,
    c.timestamp
FROM cursors AS c
INNER JOIN cursor_streams AS s ON c.stream_id = s.id
INNER JOIN users AS u ON s.user_id = u.id
WHERE
    (? IS NULL OR u.username LIKE '%' || ? || '%')
    AND (? IS NULL OR s.source = ?)
    AND (? IS NULL OR c.stream_id = ?)
    AND (? IS NULL OR c.x >= ?)
    AND (? IS NULL OR c.x <= ?)
    AND (? IS NULL OR c.y >= ?)
    AND (? IS NULL OR c.y <= ?)
    AND (? IS NULL OR c.timestamp >= ?)
    AND (? IS NULL OR c.timestamp <= ?)
ORDER BY c.id DESC

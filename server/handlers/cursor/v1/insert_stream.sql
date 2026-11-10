INSERT INTO cursor_streams (user_id, source)
SELECT
    id AS user_id,
    ? AS source
FROM users
WHERE username = ?
RETURNING id

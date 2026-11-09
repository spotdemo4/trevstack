SELECT users.username
FROM sessions
INNER JOIN users ON sessions.user_id = users.id
WHERE sessions.token_hash = ? AND sessions.expires_at > ?

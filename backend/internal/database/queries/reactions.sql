-- name: AddReaction :one
INSERT INTO reactions (message_id, user_id, emoji)
VALUES ($1, $2, $3)
ON CONFLICT (message_id, user_id, emoji) DO NOTHING
RETURNING *;

-- name: RemoveReaction :exec
DELETE FROM reactions WHERE message_id = $1 AND user_id = $2 AND emoji = $3;

-- name: ListMessageReactions :many
SELECT r.emoji, r.user_id, u.display_name
FROM reactions r
JOIN users u ON r.user_id = u.id
WHERE r.message_id = $1
ORDER BY r.created_at;

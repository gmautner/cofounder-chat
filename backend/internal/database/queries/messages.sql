-- name: CreateMessage :one
INSERT INTO messages (user_id, channel_id, conversation_id, parent_id, content)
VALUES ($1, $2, $3, $4, $5)
RETURNING *;

-- name: GetMessageByID :one
SELECT * FROM messages WHERE id = $1;

-- name: ListChannelMessages :many
SELECT m.*, u.display_name as author_name, u.avatar_url as author_avatar,
       (SELECT COUNT(*) FROM messages r WHERE r.parent_id = m.id) as reply_count
FROM messages m
JOIN users u ON m.user_id = u.id
WHERE m.channel_id = $1 AND m.parent_id IS NULL
ORDER BY m.created_at ASC;

-- name: ListConversationMessages :many
SELECT m.*, u.display_name as author_name, u.avatar_url as author_avatar,
       (SELECT COUNT(*) FROM messages r WHERE r.parent_id = m.id) as reply_count
FROM messages m
JOIN users u ON m.user_id = u.id
WHERE m.conversation_id = $1 AND m.parent_id IS NULL
ORDER BY m.created_at ASC;

-- name: ListThreadReplies :many
SELECT m.*, u.display_name as author_name, u.avatar_url as author_avatar
FROM messages m
JOIN users u ON m.user_id = u.id
WHERE m.parent_id = $1
ORDER BY m.created_at ASC;

-- name: UpdateMessage :one
UPDATE messages SET content = $2, is_edited = true, updated_at = now()
WHERE id = $1
RETURNING *;

-- name: DeleteMessage :exec
DELETE FROM messages WHERE id = $1;

-- name: CountUnreadChannelMessages :one
SELECT COUNT(*) FROM messages m
WHERE m.channel_id = $1
AND m.created_at > $2
AND m.parent_id IS NULL;

-- name: CountUnreadConversationMessages :one
SELECT COUNT(*) FROM messages m
WHERE m.conversation_id = $1
AND m.created_at > $2
AND m.parent_id IS NULL;

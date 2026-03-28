-- name: CreateConversation :one
INSERT INTO conversations DEFAULT VALUES
RETURNING *;

-- name: AddConversationMember :exec
INSERT INTO conversation_members (conversation_id, user_id)
VALUES ($1, $2)
ON CONFLICT DO NOTHING;

-- name: ListUserConversations :many
SELECT c.* FROM conversations c
JOIN conversation_members cm ON c.id = cm.conversation_id
WHERE cm.user_id = $1
ORDER BY c.created_at DESC;

-- name: ListConversationMembers :many
SELECT u.* FROM users u
JOIN conversation_members cm ON u.id = cm.user_id
WHERE cm.conversation_id = $1
ORDER BY u.display_name;

-- name: IsConversationMember :one
SELECT EXISTS(SELECT 1 FROM conversation_members WHERE conversation_id = $1 AND user_id = $2);

-- name: UpdateConversationLastRead :exec
UPDATE conversation_members SET last_read_at = now()
WHERE conversation_id = $1 AND user_id = $2;

-- name: GetConversationLastRead :one
SELECT last_read_at FROM conversation_members
WHERE conversation_id = $1 AND user_id = $2;

-- name: FindExistingConversation :one
SELECT cm1.conversation_id FROM conversation_members cm1
WHERE cm1.user_id = $1
AND cm1.conversation_id IN (
    SELECT cm2.conversation_id FROM conversation_members cm2
    WHERE cm2.user_id = $2
)
AND (SELECT COUNT(*) FROM conversation_members cm3 WHERE cm3.conversation_id = cm1.conversation_id) = 2
LIMIT 1;

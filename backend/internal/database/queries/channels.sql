-- name: CreateChannel :one
INSERT INTO channels (name, description, is_private, created_by)
VALUES ($1, $2, $3, $4)
RETURNING *;

-- name: GetChannelByID :one
SELECT * FROM channels WHERE id = $1;

-- name: GetChannelByName :one
SELECT * FROM channels WHERE name = $1;

-- name: ListPublicChannels :many
SELECT * FROM channels WHERE is_private = false ORDER BY name;

-- name: ListUserChannels :many
SELECT c.* FROM channels c
JOIN channel_members cm ON c.id = cm.channel_id
WHERE cm.user_id = $1
ORDER BY c.name;

-- name: AddChannelMember :exec
INSERT INTO channel_members (channel_id, user_id)
VALUES ($1, $2)
ON CONFLICT DO NOTHING;

-- name: RemoveChannelMember :exec
DELETE FROM channel_members WHERE channel_id = $1 AND user_id = $2;

-- name: IsChannelMember :one
SELECT EXISTS(SELECT 1 FROM channel_members WHERE channel_id = $1 AND user_id = $2);

-- name: ListChannelMembers :many
SELECT u.* FROM users u
JOIN channel_members cm ON u.id = cm.user_id
WHERE cm.channel_id = $1
ORDER BY u.display_name;

-- name: UpdateChannelLastRead :exec
UPDATE channel_members SET last_read_at = now()
WHERE channel_id = $1 AND user_id = $2;

-- name: GetChannelLastRead :one
SELECT last_read_at FROM channel_members
WHERE channel_id = $1 AND user_id = $2;

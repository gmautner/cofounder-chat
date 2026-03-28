-- name: CreateAttachment :one
INSERT INTO attachments (message_id, file_name, file_size, content_type, storage_path)
VALUES ($1, $2, $3, $4, $5)
RETURNING *;

-- name: ListMessageAttachments :many
SELECT * FROM attachments WHERE message_id = $1;

-- name: GetMessagesByStatusWithLimit :many
SELECT id, content, phone_number
FROM messages
WHERE status = $1
ORDER BY created_at ASC
LIMIT $2;

-- name: GetAllMessagesByStatus :many
SELECT id, content, phone_number
FROM messages
WHERE status = $1
ORDER BY created_at DESC;

-- name: UpdateMessageStatus :exec
UPDATE messages
SET status = $1, updated_at = NOW()
WHERE id = $2;

-- name: UpdateMessageId :exec
UPDATE messages
SET status = $1, message_id = $2, updated_at = NOW()
WHERE id = $3;
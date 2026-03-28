-- name: UpsertUser :one
INSERT INTO users (google_id, email, display_name, avatar_url)
VALUES ($1, $2, $3, $4)
ON CONFLICT (google_id) DO UPDATE SET
    email = EXCLUDED.email,
    display_name = EXCLUDED.display_name,
    avatar_url = EXCLUDED.avatar_url,
    updated_at = now()
RETURNING *;

-- name: GetUserByID :one
SELECT * FROM users WHERE id = $1;

-- name: GetUserByEmail :one
SELECT * FROM users WHERE email = $1;

-- name: ListUsers :many
SELECT * FROM users ORDER BY display_name;

-- name: SearchUsers :many
SELECT * FROM users
WHERE display_name ILIKE '%' || $1 || '%' OR email ILIKE '%' || $1 || '%'
ORDER BY display_name
LIMIT 20;

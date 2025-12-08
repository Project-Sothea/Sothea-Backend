-- name: GetUserByUsername :one
SELECT id, username
FROM users
WHERE username = $1;

-- name: GetUserByID :one
SELECT id, username
FROM users
WHERE id = $1;

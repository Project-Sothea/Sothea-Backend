-- name: GetUserByUsername :one
SELECT id, username, password_hash
FROM users
WHERE username = $1;

-- name: GetUserByID :one
SELECT id, username, password_hash
FROM users
WHERE id = $1;

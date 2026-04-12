-- name: GetUserByUsername :one
SELECT id, username, name
FROM users
WHERE username = $1;

-- name: GetUserByID :one
SELECT id, username, name
FROM users
WHERE id = $1;

-- name: ListUsers :many
SELECT id, username, name
FROM users
ORDER BY username;

-- name: CreateUser :one
INSERT INTO users (id, created_at, updated_at, name)
VALUES (
    $1,
    $2,
    $3,
    $4
)
RETURNING *;

-- name: GetUser :one
SELECT * FROM USERS WHERE Name=$1;

-- name: DeleteUsers :exec
DELETE FROM USERS;

-- name: GetUsers :many
SELECT * FROM USERS;

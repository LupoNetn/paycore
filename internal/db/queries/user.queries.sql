-- name: CreateUser :one
INSERT INTO users (
    full_name,
    phone_number,
    email,
    passwordHash,
    username,
    account_no,
    nationality
) VALUES (
    $1, $2, $3, $4, $5, $6, $7
)
RETURNING *;

-- name: GetUserByID :one
SELECT * FROM users
WHERE id = $1 LIMIT 1;

-- name: GetUserByEmail :one
SELECT * FROM users
WHERE email = $1 LIMIT 1;

-- name: GetUserByUsername :one
SELECT * FROM users
WHERE username = $1 LIMIT 1;

-- name: UpdateUser :one
UPDATE users
SET
    full_name = COALESCE(sqlc.narg('full_name'), full_name),
    phone_number = COALESCE(sqlc.narg('phone_number'), phone_number),
    email = COALESCE(sqlc.narg('email'), email),
    passwordHash = COALESCE(sqlc.narg('password_hash'), passwordHash),
    username = COALESCE(sqlc.narg('username'), username),
    account_no = COALESCE(sqlc.narg('account_no'), account_no),
    nationality = COALESCE(sqlc.narg('nationality'), nationality),
    updated_at = NOW()
WHERE id = sqlc.arg('id')
RETURNING *;

-- name: DeleteUser :exec
DELETE FROM users
WHERE id = $1;

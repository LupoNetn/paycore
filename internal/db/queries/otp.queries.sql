-- name: CreateOTP :one
INSERT INTO otps (
    user_id,
    code,
    purpose,
    expires_at,
    used
) VALUES (
    $1,
    $2,
    $3,
    $4,
    $5
) RETURNING *;
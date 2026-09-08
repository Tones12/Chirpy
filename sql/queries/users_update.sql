-- name: UpdateUser :one
UPDATE users
SET email = $1, hashed_passwords = $2
WHERE id = $3
RETURNING id, created_at, updated_at, email;
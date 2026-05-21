-- name: CreateGuestUser :one
INSERT INTO users (role)
VALUES ('guest')
RETURNING *;

-- name: CreateRegisteredUser :one
INSERT INTO users (
    role,
    email,
    password_hash,
    display_name
)
VALUES ('registered', $1, $2, $3)
RETURNING *;

-- name: GetUserByID :one
SELECT *
FROM users
WHERE id = $1;

-- name: GetUserByEmail :one
SELECT *
FROM users
WHERE email = $1;

-- name: ConvertGuestToRegistered :one
UPDATE users
SET
    role = 'registered',
    email = $2,
    password_hash = $3,
    display_name = COALESCE($4, display_name),
    converted_at = now()
WHERE id = $1
  AND role = 'guest'
RETURNING *;

-- name: UpdateUserDisplayName :one
UPDATE users
SET display_name = $2
WHERE id = $1
RETURNING *;

-- name: EmailExists :one
SELECT EXISTS(SELECT 1 FROM users WHERE email = $1) AS exists;

-- name: ListAccessibleEQProfiles :many
SELECT *
FROM eq_profiles
WHERE is_system = true OR owner_user_id = $1
ORDER BY
    is_system DESC,
    CASE WHEN is_system THEN is_default END DESC,
    CASE WHEN is_system THEN lower(name) END ASC,
    created_at DESC;

-- name: GetDefaultEQProfile :one
SELECT *
FROM eq_profiles
WHERE is_system = true
  AND is_default = true;

-- name: ListSystemEQProfiles :many
SELECT *
FROM eq_profiles
WHERE is_system = true
ORDER BY is_default DESC, lower(name) ASC;

-- name: ListCustomEQProfiles :many
SELECT *
FROM eq_profiles
WHERE is_system = false
  AND owner_user_id = $1
ORDER BY created_at DESC;

-- name: GetAccessibleEQProfile :one
SELECT *
FROM eq_profiles
WHERE id = $1
  AND (is_system = true OR owner_user_id = $2);

-- name: CreateCustomEQProfile :one
INSERT INTO eq_profiles (
    owner_user_id,
    name,
    is_system,
    bands
)
VALUES ($1, $2, false, $3)
RETURNING *;

-- name: UpdateCustomEQProfile :one
UPDATE eq_profiles
SET
    name = $3,
    bands = $4,
    updated_at = now()
WHERE id = $1
  AND owner_user_id = $2
  AND is_system = false
RETURNING *;

-- name: DeleteCustomEQProfile :execrows
DELETE FROM eq_profiles
WHERE id = $1
  AND owner_user_id = $2
  AND is_system = false;

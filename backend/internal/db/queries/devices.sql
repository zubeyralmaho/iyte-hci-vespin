-- name: ListDevicesByUser :many
SELECT *
FROM devices
WHERE user_id = $1
ORDER BY created_at DESC;

-- name: GetDeviceByIDAndUser :one
SELECT *
FROM devices
WHERE id = $1
  AND user_id = $2;

-- name: ListDevicesByIDsAndUser :many
SELECT *
FROM devices
WHERE user_id = $1
  AND id = ANY(sqlc.arg(device_ids)::uuid[])
ORDER BY created_at DESC;

-- name: CountDevicesByIDsAndUser :one
SELECT count(*)
FROM devices
WHERE user_id = $1
  AND id = ANY(sqlc.arg(device_ids)::uuid[]);

-- name: CreateDevice :one
INSERT INTO devices (
    user_id,
    name,
    device_type,
    firmware_version,
    battery_level,
    is_connected,
    active_eq_profile_id
)
VALUES ($1, $2, $3, $4, $5, $6, $7)
RETURNING *;

-- name: UpdateDevice :one
UPDATE devices
SET
    name = $3,
    active_eq_profile_id = $4,
    updated_at = now()
WHERE id = $1
  AND user_id = $2
RETURNING *;

-- name: DeleteDevice :execrows
DELETE FROM devices
WHERE id = $1
  AND user_id = $2;

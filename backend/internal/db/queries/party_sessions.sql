-- name: ListPartySessionsByOwner :many
SELECT
    ps.*,
    COALESCE(
        array_agg(psd.device_id ORDER BY psd.joined_at) FILTER (WHERE psd.device_id IS NOT NULL),
        ARRAY[]::uuid[]
    )::uuid[] AS device_ids
FROM party_sessions ps
LEFT JOIN party_session_devices psd ON psd.party_session_id = ps.id
WHERE ps.owner_user_id = $1
GROUP BY ps.id
ORDER BY ps.started_at DESC;

-- name: ListPartySessionsByOwnerAndStatus :many
SELECT
    ps.*,
    COALESCE(
        array_agg(psd.device_id ORDER BY psd.joined_at) FILTER (WHERE psd.device_id IS NOT NULL),
        ARRAY[]::uuid[]
    )::uuid[] AS device_ids
FROM party_sessions ps
LEFT JOIN party_session_devices psd ON psd.party_session_id = ps.id
WHERE ps.owner_user_id = $1
  AND ps.status = $2
GROUP BY ps.id
ORDER BY ps.started_at DESC;

-- name: GetPartySessionByIDAndOwner :one
SELECT
    ps.*,
    COALESCE(
        array_agg(psd.device_id ORDER BY psd.joined_at) FILTER (WHERE psd.device_id IS NOT NULL),
        ARRAY[]::uuid[]
    )::uuid[] AS device_ids
FROM party_sessions ps
LEFT JOIN party_session_devices psd ON psd.party_session_id = ps.id
WHERE ps.id = $1
  AND ps.owner_user_id = $2
GROUP BY ps.id;

-- name: CreatePartySession :one
INSERT INTO party_sessions (
    owner_user_id,
    name
)
VALUES ($1, $2)
RETURNING *;

-- name: UpdatePartySession :one
UPDATE party_sessions
SET
    name = $3,
    status = $4,
    ended_at = CASE
        WHEN $4 = 'ended' THEN COALESCE(ended_at, now())
        ELSE NULL
    END,
    updated_at = now()
WHERE id = $1
  AND owner_user_id = $2
RETURNING *;

-- name: DeletePartySession :execrows
DELETE FROM party_sessions
WHERE id = $1
  AND owner_user_id = $2;

-- name: AddPartySessionDevice :execrows
INSERT INTO party_session_devices (
    party_session_id,
    device_id
)
SELECT $1, $2
WHERE EXISTS (
    SELECT 1
    FROM party_sessions
    WHERE id = $1
      AND owner_user_id = $3
);

-- name: AddPartySessionDevices :execrows
INSERT INTO party_session_devices (
    party_session_id,
    device_id
)
SELECT $1, input.device_id
FROM unnest(sqlc.arg(device_ids)::uuid[]) AS input(device_id)
WHERE EXISTS (
    SELECT 1
    FROM party_sessions
    WHERE id = $1
      AND owner_user_id = $2
);

-- name: RemovePartySessionDevice :execrows
DELETE FROM party_session_devices psd
USING party_sessions ps
WHERE ps.id = psd.party_session_id
  AND ps.owner_user_id = $3
  AND psd.party_session_id = $1
  AND psd.device_id = $2;

-- name: ListPartySessionDeviceIDs :many
SELECT psd.device_id
FROM party_session_devices psd
JOIN party_sessions ps ON ps.id = psd.party_session_id
WHERE psd.party_session_id = $1
  AND ps.owner_user_id = $2
ORDER BY psd.joined_at;

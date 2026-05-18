-- name: GetUserPreferences :one
SELECT *
FROM user_preferences
WHERE user_id = $1;

-- name: UpsertUserPreferences :one
INSERT INTO user_preferences (
    user_id,
    theme,
    language,
    notifications_enabled
)
VALUES ($1, $2, $3, $4)
ON CONFLICT (user_id) DO UPDATE SET
    theme = EXCLUDED.theme,
    language = EXCLUDED.language,
    notifications_enabled = EXCLUDED.notifications_enabled,
    updated_at = now()
RETURNING *;

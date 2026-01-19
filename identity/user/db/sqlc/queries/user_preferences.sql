-- name: CreateUserPreferences :exec
INSERT INTO user_preferences (
    user_id, language, timezone, date_format, time_format,
    email_notifications, sms_notifications, push_notifications,
    custom_preferences
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8, $9
);

-- name: GetUserPreferences :one
SELECT * FROM user_preferences
WHERE user_id = $1;

-- name: UpdateUserPreferences :exec
UPDATE user_preferences
SET language = $2,
    timezone = $3,
    date_format = $4,
    time_format = $5,
    email_notifications = $6,
    sms_notifications = $7,
    push_notifications = $8,
    custom_preferences = $9,
    updated_at = now()
WHERE user_id = $1;

-- name: UpsertUserPreferences :exec
INSERT INTO user_preferences (
    user_id, language, timezone, date_format, time_format,
    email_notifications, sms_notifications, push_notifications,
    custom_preferences
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8, $9
)
ON CONFLICT (user_id) DO UPDATE SET
    language = EXCLUDED.language,
    timezone = EXCLUDED.timezone,
    date_format = EXCLUDED.date_format,
    time_format = EXCLUDED.time_format,
    email_notifications = EXCLUDED.email_notifications,
    sms_notifications = EXCLUDED.sms_notifications,
    push_notifications = EXCLUDED.push_notifications,
    custom_preferences = EXCLUDED.custom_preferences,
    updated_at = now();

-- name: DeleteUserPreferences :exec
DELETE FROM user_preferences
WHERE user_id = $1;
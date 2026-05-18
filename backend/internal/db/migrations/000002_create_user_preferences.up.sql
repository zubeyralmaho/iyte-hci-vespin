CREATE TABLE user_preferences (
    user_id uuid PRIMARY KEY REFERENCES users(id) ON DELETE CASCADE,
    theme text NOT NULL DEFAULT 'system' CHECK (theme IN ('light', 'dark', 'system')),
    language text NOT NULL DEFAULT 'en' CHECK (language ~ '^[a-z]{2}$'),
    notifications_enabled boolean NOT NULL DEFAULT true,
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now()
);

CREATE TABLE eq_profiles (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    owner_user_id uuid REFERENCES users(id) ON DELETE CASCADE,
    name text NOT NULL CHECK (char_length(trim(name)) BETWEEN 1 AND 100),
    is_system boolean NOT NULL DEFAULT false,
    bands jsonb NOT NULL,
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now(),
    CHECK (
        (is_system = true AND owner_user_id IS NULL)
        OR
        (is_system = false AND owner_user_id IS NOT NULL)
    )
);

CREATE UNIQUE INDEX eq_profiles_system_name_unique_idx
    ON eq_profiles (lower(name))
    WHERE is_system = true;

CREATE INDEX eq_profiles_owner_user_created_at_idx
    ON eq_profiles (owner_user_id, created_at DESC)
    WHERE owner_user_id IS NOT NULL;

CREATE TABLE devices (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id uuid NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    name text NOT NULL CHECK (char_length(trim(name)) BETWEEN 1 AND 100),
    device_type text NOT NULL CHECK (device_type IN ('vespin_classic', 'vespin_mini', 'vespin_pro')),
    firmware_version text NOT NULL,
    battery_level integer NOT NULL CHECK (battery_level BETWEEN 0 AND 100),
    is_connected boolean NOT NULL,
    active_eq_profile_id uuid REFERENCES eq_profiles(id) ON DELETE SET NULL,
    paired_at timestamptz NOT NULL DEFAULT now(),
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now()
);

CREATE INDEX devices_user_created_at_idx
    ON devices (user_id, created_at DESC);


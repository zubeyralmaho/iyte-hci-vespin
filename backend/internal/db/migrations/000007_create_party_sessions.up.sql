CREATE TABLE party_sessions (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    owner_user_id uuid NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    name text CHECK (name IS NULL OR char_length(trim(name)) BETWEEN 1 AND 100),
    status text NOT NULL DEFAULT 'active' CHECK (status IN ('active', 'paused', 'ended')),
    started_at timestamptz NOT NULL DEFAULT now(),
    ended_at timestamptz,
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now(),
    CHECK (
        (status = 'ended' AND ended_at IS NOT NULL)
        OR
        (status IN ('active', 'paused') AND ended_at IS NULL)
    )
);

CREATE INDEX party_sessions_owner_started_at_idx
    ON party_sessions (owner_user_id, started_at DESC);

CREATE INDEX party_sessions_owner_status_started_at_idx
    ON party_sessions (owner_user_id, status, started_at DESC);

CREATE TABLE party_session_devices (
    party_session_id uuid NOT NULL REFERENCES party_sessions(id) ON DELETE CASCADE,
    device_id uuid NOT NULL REFERENCES devices(id) ON DELETE CASCADE,
    joined_at timestamptz NOT NULL DEFAULT now(),
    PRIMARY KEY (party_session_id, device_id)
);

CREATE INDEX party_session_devices_device_id_idx
    ON party_session_devices (device_id);


ALTER TABLE eq_profiles
ADD COLUMN is_default boolean NOT NULL DEFAULT false;

ALTER TABLE eq_profiles
ADD CONSTRAINT eq_profiles_default_must_be_system_check
CHECK (is_default = false OR is_system = true);

CREATE UNIQUE INDEX eq_profiles_one_default_idx
    ON eq_profiles (is_default)
    WHERE is_default = true;

UPDATE eq_profiles
SET is_default = true
WHERE is_system = true
  AND name = 'Flat';

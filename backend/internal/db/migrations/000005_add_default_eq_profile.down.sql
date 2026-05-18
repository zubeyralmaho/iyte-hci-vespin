DROP INDEX IF EXISTS eq_profiles_one_default_idx;

ALTER TABLE eq_profiles
DROP CONSTRAINT IF EXISTS eq_profiles_default_must_be_system_check;

ALTER TABLE eq_profiles
DROP COLUMN IF EXISTS is_default;

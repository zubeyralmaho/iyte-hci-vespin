INSERT INTO eq_profiles (name, is_system, bands)
VALUES
    ('Flat', true, '{"subBass":0,"bass":0,"mid":0,"treble":0,"presence":0}'::jsonb),
    ('Bass Boost', true, '{"subBass":5,"bass":4,"mid":0,"treble":1,"presence":0}'::jsonb),
    ('Rock', true, '{"subBass":3,"bass":4,"mid":-1,"treble":3,"presence":2}'::jsonb),
    ('Jazz', true, '{"subBass":1,"bass":2,"mid":3,"treble":2,"presence":1}'::jsonb),
    ('Classical', true, '{"subBass":0,"bass":1,"mid":2,"treble":3,"presence":2}'::jsonb),
    ('R&B', true, '{"subBass":4,"bass":5,"mid":1,"treble":2,"presence":1}'::jsonb);

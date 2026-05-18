DELETE FROM eq_profiles
WHERE is_system = true
  AND name IN (
    'Flat',
    'Bass Boost',
    'Rock',
    'Jazz',
    'Classical',
    'R&B'
  );

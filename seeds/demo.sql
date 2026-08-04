INSERT INTO platform_metadata (key, value, updated_at)
VALUES (
    'demo_seed',
    '{"version": 1, "accounts": "identity-phase"}'::jsonb,
    now()
)
ON CONFLICT (key) DO UPDATE
SET value = EXCLUDED.value,
    updated_at = EXCLUDED.updated_at;

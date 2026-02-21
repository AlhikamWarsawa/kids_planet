-- Optional player authentication schema upgrade (email + 6-digit PIN)

ALTER TABLE users
    ADD COLUMN IF NOT EXISTS public_id UUID,
    ADD COLUMN IF NOT EXISTS pin_hash VARCHAR(255);

DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1
        FROM pg_constraint
        WHERE conname = 'ck_users_player_auth_fields'
    ) THEN
        ALTER TABLE users
            ADD CONSTRAINT ck_users_player_auth_fields
                CHECK (role <> 'player' OR (public_id IS NOT NULL AND pin_hash IS NOT NULL));
    END IF;
END $$;

CREATE UNIQUE INDEX IF NOT EXISTS idx_users_public_id_unique
    ON users (public_id)
    WHERE public_id IS NOT NULL;

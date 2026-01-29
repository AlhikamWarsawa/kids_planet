DO $$
    BEGIN
        IF NOT EXISTS (
            SELECT 1
            FROM pg_constraint
            WHERE conname = 'ck_age_categories_valid_range'
        ) THEN
            ALTER TABLE age_categories
                ADD CONSTRAINT ck_age_categories_valid_range
                    CHECK (min_age >= 0 AND max_age >= min_age);
        END IF;
    END $$;

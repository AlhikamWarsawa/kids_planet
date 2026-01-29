DO $$
    BEGIN
        IF EXISTS (
            SELECT 1
            FROM pg_constraint
            WHERE conname = 'ck_age_categories_valid_range'
        ) THEN
            ALTER TABLE age_categories
                DROP CONSTRAINT ck_age_categories_valid_range;
        END IF;
    END $$;

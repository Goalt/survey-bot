DO $$ BEGIN
    ALTER TABLE users DROP COLUMN last_activity;
EXCEPTION
    WHEN undefined_column THEN null;
END $$;

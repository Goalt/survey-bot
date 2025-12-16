DO $$ BEGIN
    ALTER TABLE users DROP COLUMN nickname;
EXCEPTION
    WHEN undefined_column THEN null;
END $$;

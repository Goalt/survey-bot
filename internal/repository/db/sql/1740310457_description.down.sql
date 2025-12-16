DO $$ BEGIN
    ALTER TABLE surveys DROP COLUMN description;
EXCEPTION
    WHEN undefined_column THEN null;
END $$;

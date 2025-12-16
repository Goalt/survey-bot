DO $$ BEGIN
    ALTER TABLE surveys ADD description varchar NOT NULL DEFAULT '';
EXCEPTION
    WHEN duplicate_column THEN null;
END $$;

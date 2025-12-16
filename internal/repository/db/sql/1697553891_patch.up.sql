DO $$ BEGIN
    ALTER TABLE surveys ADD calculations_type varchar NOT NULL;
EXCEPTION
    WHEN duplicate_column THEN null;
END $$;
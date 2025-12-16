DO $$ BEGIN
    ALTER TABLE users ADD nickname varchar NOT NULL DEFAULT 'unknown';
EXCEPTION
    WHEN duplicate_column THEN null;
END $$;
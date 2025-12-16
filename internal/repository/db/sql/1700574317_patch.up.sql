DO $$ BEGIN
    ALTER TABLE public.survey_states ADD results JSONB;
EXCEPTION
    WHEN duplicate_column THEN null;
END $$;

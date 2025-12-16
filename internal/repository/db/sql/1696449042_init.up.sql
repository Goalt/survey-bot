DO $$ BEGIN
    CREATE TYPE survey_states_types AS enum ('finished', 'active');
EXCEPTION
    WHEN duplicate_object THEN null;
END $$;

CREATE TABLE IF NOT EXISTS surveys (
    guid UUID NOT NULL,
    id NUMERIC NOT NULL,
    name varchar(400) NOT NULL,
    questions JSONB NOT NULL,
    created_at TIMESTAMPTZ NOT NULL,
    updated_at TIMESTAMPTZ NOT NULL,
    deleted_at TIMESTAMPTZ,
    CONSTRAINT surveys_pk PRIMARY KEY (guid)
);

CREATE TABLE IF NOT EXISTS users (
    guid UUID NOT NULL,
    current_survey UUID,
    user_id NUMERIC NOT NULL,
    chat_id NUMERIC NOT NULL,
    created_at TIMESTAMPTZ NOT NULL,
    updated_at TIMESTAMPTZ NOT NULL,
    CONSTRAINT users_pk PRIMARY KEY (guid)
);

CREATE TABLE IF NOT EXISTS survey_states (
    user_guid UUID NOT NULL,
    survey_guid UUID NOT NULL,
    state survey_states_types,
    answers JSONB NOT NULL,
    created_at TIMESTAMPTZ NOT NULL,
    updated_at TIMESTAMPTZ NOT NULL
);

ALTER TABLE survey_states DROP CONSTRAINT IF EXISTS survey_states_user_guid_fk;
ALTER TABLE
    survey_states
ADD
    CONSTRAINT survey_states_user_guid_fk FOREIGN KEY (user_guid) REFERENCES users(guid);

ALTER TABLE survey_states DROP CONSTRAINT IF EXISTS survey_states_survey_guid_fk;
ALTER TABLE
    survey_states
ADD
    CONSTRAINT survey_states_survey_guid_fk FOREIGN KEY (survey_guid) REFERENCES surveys(guid);

ALTER TABLE users DROP CONSTRAINT IF EXISTS users_current_survey_fk;
ALTER TABLE
    users
ADD
    CONSTRAINT users_current_survey_fk FOREIGN KEY (current_survey) REFERENCES surveys(guid);
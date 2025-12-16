SELECT *
FROM pg_catalog.pg_tables
WHERE schemaname NOT IN ('pg_catalog','information_schema');

select * from users;


select * from survey_states where user_guid = '413c85aa-829a-4ee3-8e7c-c02732a4b2e7' and survey_guid = 'c89a44a1-5c09-43b1-a55b-33f8ccf6cf17';


select * from surveys where guid = 'c89a44a1-5c09-43b1-a55b-33f8ccf6cf17';

update users set current_survey = NULL;


†

DELETE from survey_states;
DO $$ 
DECLARE
    schema_name text;
BEGIN
    FOR schema_name IN 
        SELECT schema_name 
        FROM information_schema.schemata 
        WHERE schema_name NOT LIKE 'pg_%' AND schema_name != 'information_schema' -- Exclude system schemas if desired
    LOOP
        EXECUTE format('CREATE EXTENSION IF NOT EXISTS "uuid-ossp" SCHEMA %I;', schema_name);
    END LOOP;
END $$;

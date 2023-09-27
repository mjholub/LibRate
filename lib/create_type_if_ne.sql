DO $$
BEGIN
    IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = '{{typeName}}') THEN
        CREATE TYPE {{schema}}.{{typeName}} AS ENUM ({{enum_values}});
    END IF;
END
$$;

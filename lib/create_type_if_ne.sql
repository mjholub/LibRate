DO $$
BEGIN
  BEGIN 
        CREATE TYPE {{schema}}.{{typeName}} AS ENUM ({{enum_values}});
EXCEPTION
        WHEN duplicate_object THEN
          RAISE EXCEPTION 'type_exists';
    END;
END $$;

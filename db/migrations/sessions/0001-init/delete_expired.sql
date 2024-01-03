CREATE OR REPLACE FUNCTION delete_expired_records()
RETURNS VOID AS $$
BEGIN
    DELETE FROM your_table
    WHERE expires < CURRENT_TIMESTAMP AND (NOT authenticated OR NOT authenticating);
END;
$$ LANGUAGE plpgsql;

-- modified is a UNIX seconds int64 timestamp for internal use
-- Create a common trigger function
CREATE OR REPLACE FUNCTION modified() RETURNS TRIGGER AS $$
BEGIN
  NEW.modified = EXTRACT(EPOCH FROM now());
  RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Add new columns and triggers for each table
DO $$
DECLARE
  _tbl text;
  _schema text;
BEGIN
  FOR _schema, _tbl IN (SELECT table_schema, table_name FROM information_schema.tables WHERE table_schema IN ('public', 'media', 'people') AND table_name IN ('members', 'media', 'group', 'studio'))
  LOOP
    EXECUTE format('ALTER TABLE %I.%I ADD COLUMN IF NOT EXISTS modified bigint', _schema, _tbl);
    EXECUTE format('DROP TRIGGER IF EXISTS %I_modified ON %I.%I', _tbl, _schema, _tbl);
    EXECUTE format('CREATE TRIGGER %I_modified BEFORE INSERT OR UPDATE ON %I.%I FOR EACH ROW EXECUTE FUNCTION modified()', _tbl, _schema, _tbl);
  END LOOP;
END
$$;

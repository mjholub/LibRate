CREATE OR REPLACE FUNCTION update_modified_column()
RETURNS TRIGGER AS $$
BEGIN
  NEW.modified = NOW(); -- Update the 'modified' column with the current timestamp
  RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER person_update_modified
BEFORE UPDATE ON people.person
FOR EACH ROW
EXECUTE FUNCTION update_modified_column();

CREATE TRIGGER media_update_modified
BEFORE UPDATE ON media.media
FOR EACH ROW
EXECUTE FUNCTION update_modified_column();

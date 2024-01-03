CREATE OR REPLACE FUNCTION on_image_delete()
  RETURNS TRIGGER AS $$
BEGIN
  -- Set the value of referencing foreign keys to NULL
  UPDATE public.members
  SET profilepic_id = NULL
  WHERE profilepic_id = OLD.id;

   RETURN OLD;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER image_delete_trigger
BEFORE DELETE ON cdn.images
FOR EACH ROW
EXECUTE FUNCTION on_image_delete();

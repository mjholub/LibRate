CREATE OR REPLACE FUNCTION couchdb_delete() RETURNS trigger AS $BODY$
DECLARE
    RES RECORD;
BEGIN
  IF (OLD.from_pg) IS NULL THEN
    RETURN OLD;
  ELSE 
    -- note that this works only in docker
    SELECT status FROM http_delete('librate-search:5984/' || TG_TABLE_NAME || '/' || OLD.id::text) INTO RES;    
  
    --Need to check RES for response code
    RAISE EXCEPTION 'Result: %', RES;
    RETURN null;
  END IF;
  END;
$BODY$
LANGUAGE plpgsql VOLATILE;

CREATE OR REPLACE TRIGGER couchdb_member_deletion_sync
AFTER DELETE ON public.members
FOR EACH ROW
EXECUTE PROCEDURE couchdb_delete();

CREATE OR REPLACE TRIGGER couchdb_media_deletion_sync
AFTER DELETE ON media.media
FOR EACH ROW
EXECUTE PROCEDURE couchdb_delete();

CREATE OR REPLACE TRIGGER couchdb_person_deletion_sync
AFTER DELETE ON people.person
FOR EACH ROW
EXECUTE PROCEDURE couchdb_delete();

CREATE OR REPLACE TRIGGER couchdb_group_deletion_sync
AFTER DELETE ON people.group
FOR EACH ROW
EXECUTE PROCEDURE couchdb_delete();

CREATE OR REPLACE TRIGGER couchdb_studio_deletion_sync
AFTER DELETE ON people.studio
FOR EACH ROW
EXECUTE PROCEDURE couchdb_delete();

CREATE OR REPLACE TRIGGER couchdb_place_deletion_sync
AFTER DELETE ON places.place
FOR EACH ROW
EXECUTE PROCEDURE couchdb_delete();

CREATE OR REPLACE TRIGGER couchdb_genre_deletion_sync
AFTER DELETE ON media.genres
FOR EACH ROW
EXECUTE PROCEDURE couchdb_delete();

CREATE OR REPLACE TRIGGER couchdb_genre_description_deletion_sync
AFTER DELETE ON media.genre_descriptions
FOR EACH ROW
EXECUTE PROCEDURE couchdb_delete();

CREATE OR REPLACE TRIGGER couchdb_genre_characteristics_deletion_sync
AFTER DELETE ON media.genre_characteristics
FOR EACH ROW
EXECUTE PROCEDURE couchdb_delete();

CREATE OR REPLACE TRIGGER couchdb_ratings_deletion_sync
AFTER DELETE ON reviews.ratings
FOR EACH ROW
EXECUTE PROCEDURE couchdb_delete();

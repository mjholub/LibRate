ALTER TABLE reviews.ratings RENAME COLUMN id TO id_numeric;

ALTER TABLE reviews.ratings ADD "id" uuid NOT NULL DEFAULT uuid_time_nextval(30,65536);
COMMENT ON COLUMN reviews.ratings."id" IS 'needed for sync with couchdb';

ALTER TABLE people.studio RENAME COLUMN id TO id_numeric;

ALTER TABLE people.studio ADD "id" uuid NOT NULL DEFAULT uuid_time_nextval(30,65536);
COMMENT ON COLUMN people.studio."id" IS 'needed for sync with couchdb';

ALTER TABLE public.members RENAME COLUMN id TO id_numeric;
ALTER TABLE public.members RENAME COLUMN uuid TO id;



-- not sure if we need this
CREATE TABLE since_checkpoints (
  pgtable text NOT NULL,
  since numeric DEFAULT 0,
  enabled boolean DEFAULT false,
  CONSTRAINT since_checkpoints_pkey PRIMARY KEY (pgtable)
);

-- create a function to sync data to couchdb
CREATE OR REPLACE FUNCTION couchdb_put() RETURNS trigger AS $BODY$
DECLARE
    REFERER_HEADER public.http_header;
    DOC_ID TEXT;
    RES RECORD;
BEGIN
 REFERER_HEADER := http_header('Referer', 'librate-search:5984');
 IF (NEW.from_pg) IS NULL THEN
   RETURN NEW;
 ELSE 
  PERFORM doc_id FROM NEW WHERE doc_id IS NOT NULL;

  IF FOUND THEN
    -- note that this works only in docker
    SELECT status FROM http_put('librate-search:5984/' || TG_TABLE_NAME || '/' || NEW.doc_id::text, NEW.doc::text, 'application/json') INTO RES;
  ELSE
    -- note that this works only in docker
    SELECT status, content::json->'id' FROM http_post('librate-search:5984/' || TG_TABLE_NAME || '/', NEW.doc::text, 'application/json') INTO RES;
  END IF;

  DOC_ID := RES.id;

   --Need to check RES for response code
   IF RES.status != 201 THEN
   	RAISE EXCEPTION 'Error: %', RES;
   END IF;
  
  UPDATE NEW.doc_id SET doc_id = 'id';
  RETURN null;
 END IF;
END;
$BODY$
LANGUAGE plpgsql VOLATILE

-- create a loop to add doc (jsonb encoded data for what will be copied to couchdb),
-- from_pg (to trigger updating couchdb) and doc_id (to store the couchdb id)
-- If doc_id is not null, it'll be then used to update the document in couchdb
-- to each table that will be synced to couchdb
DO $$
DECLARE
_schema text;
_tbl text;
BEGIN 
  FOR _schema, _tbl IN (SELECT table_schema, table_name
    FROM information_schema.tables WHERE table_schema IN
    ('public', 'media', 'people', 'reviews') AND table_name IN
    ('members', 'media', 'person', 'group', 'studio', 'ratings', 'genres', 'genre_descriptions', 'genre_characteristics'))
  LOOP
    EXECUTE format('ALTER TABLE %I.%I ADD COLUMN IF NOT EXISTS doc jsonb', _schema, _tbl);
    EXECUTE format('ALTER TABLE %I.%I ADD COLUMN IF NOT EXISTS from_pg boolean DEFAULT true', _schema, _tbl);
    EXECUTE format('ALTER TABLE %I.%I ADD COLUMN IF NOT EXISTS doc_id text', _schema, _tbl);
    EXECUTE format('UPDATE %I.%I SET from_pg = true WHERE from_pg IS NULL', _schema, _tbl);
    EXECUTE format('ALTER TABLE %I.%I ALTER COLUMN from_pg SET NOT NULL', _schema, _tbl);
  END LOOP;
END;
$$;



CREATE OR REPLACE TRIGGER couchdb_sync_members
AFTER INSERT OR UPDATE OF
bio, display_name, webfinger
ON public.members
FOR EACH ROW
EXECUTE PROCEDURE couchdb_put();

-- media 

CREATE OR REPLACE TRIGGER couchdb_sync_media
AFTER INSERT OR UPDATE OF
title, kind, created, added, modified
ON media.media
FOR EACH ROW
EXECUTE PROCEDURE couchdb_put();

CREATE OR REPLACE TRIGGER couchdb_sync_media_images
AFTER INSERT OR UPDATE ON media.media_images
FOR EACH ROW
EXECUTE PROCEDURE couchdb_put();

-- artists and studios

CREATE OR REPLACE TRIGGER couchdb_sync_artist_person
AFTER INSERT OR UPDATE OF
first_name, last_name, nick_names, roles, bio, added, modified
ON people.person
FOR EACH ROW
EXECUTE PROCEDURE couchdb_put();

CREATE OR REPLACE TRIGGER couchdb_sync_artist_group
AFTER INSERT OR UPDATE OF
name, active, bio, added, modified
ON people.group
FOR EACH ROW
EXECUTE PROCEDURE couchdb_put();

CREATE OR REPLACE TRIGGER couchdb_sync_studios
AFTER INSERT OR UPDATE OF
name, kind, city, added, modified
ON people.studio
FOR EACH ROW
  EXECUTE PROCEDURE couchdb_put();

-- genres

CREATE OR REPLACE TRIGGER couchdb_sync_genres
AFTER INSERT OR UPDATE OF
name, kinds
ON media.genres
FOR EACH ROW
  EXECUTE PROCEDURE couchdb_put();

CREATE OR REPLACE TRIGGER couchdb_sync_genre_descriptions
AFTER INSERT OR UPDATE OF
language, description, genre_id
ON media.genre_descriptions
FOR EACH ROW
  EXECUTE PROCEDURE couchdb_put();

CREATE OR REPLACE TRIGGER couchdb_sync_genre_characteristics
AFTER INSERT OR UPDATE OF
id, name, description
ON media.genre_characteristics
FOR EACH ROW
  EXECUTE PROCEDURE couchdb_put();

-- reviews

CREATE OR REPLACE TRIGGER couchdb_sync_reviews_basic
AFTER INSERT OR UPDATE OF
topic, body, user_id, media_id, added, modified
ON reviews.ratings
FOR EACH ROW
  EXECUTE PROCEDURE couchdb_put();

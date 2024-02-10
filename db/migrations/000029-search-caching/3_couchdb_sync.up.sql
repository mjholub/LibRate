ALTER TABLE reviews.ratings RENAME COLUMN id TO id_numeric;

ALTER TABLE reviews.ratings ADD "id" uuid NOT NULL DEFAULT uuid_time_nextval(30,65536);
COMMENT ON COLUMN reviews.ratings."id" IS 'needed for sync with couchdb';

ALTER TABLE people.studio RENAME COLUMN id TO id_numeric;

ALTER TABLE people.studio ADD "id" uuid NOT NULL DEFAULT uuid_time_nextval(30,65536);
COMMENT ON COLUMN people.studio."id" IS 'needed for sync with couchdb';

ALTER TABLE public.members RENAME COLUMN id TO id_numeric;
ALTER TABLE public.members RENAME COLUMN uuid TO id;

ALTER TABLE media.genre_descriptions
ADD COLUMN IF NOT EXISTS id bigserial NOT NULL;

-- not sure if we need this
CREATE TABLE since_checkpoints (
  pgtable text NOT NULL,
  since numeric DEFAULT 0,
  enabled boolean DEFAULT false,
  CONSTRAINT since_checkpoints_pkey PRIMARY KEY (pgtable)
);

-- create a function to sync data to couchdb
-- rip my sanity ;_;
CREATE OR REPLACE FUNCTION couchdb_put() RETURNS trigger AS $BODY$
DECLARE
    DOC_ID TEXT;
    REV_ID TEXT;
	  AUTH TEXT;
	 	DOC jsonb;
    RES public.http_response;
	  FINAL_REQUEST public.http_request;
	 	FULL_ARTIST_NAME TEXT;
 		REVIEWER_WF TEXT;
 		REVIEW_MEDIA_TITLE TEXT;
BEGIN
	-- it might seem stupid, but even a proper 
	-- alphabetic trigger naming did not work as expected
	-- and doc column contained NULL data
	CASE TG_TABLE_NAME
		WHEN 'genres' THEN
			DOC := jsonb_build_object('name', NEW.name, 'kinds', NEW.kinds);
		WHEN 'members' THEN
			DOC := jsonb_build_object('bio', NEW.bio, 'display_name', NEW.display_name, 'webfinger', NEW.webfinger);
		WHEN 'media' THEN
			-- TODO: add more complex logic and
			-- relevant CouchDB datastores for albums, films etc.
			-- based on NEW.kind
			DOC := jsonb_build_object('title', NEW.title, 'kind', NEW.kind, 'created', NEW.created, 'added', NEW.added, 'modified', NEW.modified);
		WHEN 'person' THEN
			FULL_ARTIST_NAME := CONCAT(NEW.first_name, ' ', NEW.last_name);
    	DOC := jsonb_build_object('name', FULL_ARTIST_NAME, 'nick_names', NEW.nick_names, 'bio', NEW.bio, 'added', NEW.added, 'modified', NEW.modified);
		WHEN 'group' THEN
			DOC := jsonb_build_object('name', NEW.name, 'active', NEW.active, 'bio', NEW.bio, 'added', NEW.added, 'modified', NEW.modified);
		WHEN 'studio' THEN
			DOC := jsonb_build_object('name', NEW.name, 'kind', NEW.kind, 'city', NEW.city, 'added', NEW.added, 'modified', NEW.modified);
		WHEN 'genre_descriptions' THEN
			DOC := jsonb_build_object('language', NEW.language, 'description', NEW.description, 'genre_id', NEW.genre_id);
		WHEN 'genre_characteristics' THEN
			DOC := jsonb_build_object('id', NEW.id, 'name', NEW.name, 'description', NEW.description);
		WHEN 'ratings' THEN
			REVIEWER_WF := (SELECT webfinger FROM public.members WHERE uuid = NEW.user_id);
  		REVIEW_MEDIA_TITLE := (SELECT title FROM media.media WHERE id = NEW.media_id);
  		DOC := jsonb_build_object('topic', NEW.topic, 'body', NEW.body, 'user_id', webfinger, 'media_id', media_title, 'added', NEW.added, 'modified', NEW.modified);
		END CASE;
	
	-- the actual request starts here
 AUTH := encode('librate:librate', 'base64');
    EXECUTE format('SELECT doc_id FROM %I.%I WHERE id = $1', TG_TABLE_SCHEMA, TG_TABLE_NAME) INTO DOC_ID USING NEW.id;
  IF DOC_ID IS NOT NULL THEN
    -- first, get the rev_id from couchdb to avoid conflicts
    SELECT content::json->>'_rev' FROM http_get('librate-search:5984/' || TG_TABLE_NAME || '/' || DOC_ID) INTO REV_ID;
   RAISE NOTICE 'REV_ID: %', REV_ID;
    FINAL_REQUEST := (
    'PUT',
    'http://172.20.0.6:5984/' || TG_TABLE_NAME || '/' || NEW.doc_id::text, 
    ARRAY[http_header('Accept','application/json'),
 http_header('Authorization', 'Basic' || AUTH),
 http_header('Referer', 'librate-search:5984')],
 'application/json',
      jsonb_set(DOC, '{_rev}', to_jsonb(REV_ID), '{_id}', to_jsonb(NEW.doc_id::text)))::public.http_request;
  ELSE
    -- note that this works only in docker
  	-- For authorization you need to derive your
    FINAL_REQUEST := (
    'POST',
  'http://172.20.0.6:5984/' || TG_TABLE_NAME || '/',
 ARRAY[http_header('Accept','application/json'),
 http_header('Authorization', 'Basic' || AUTH),
 http_header('Referer', 'librate-search:5984')],
  'application/json', DOC::text)::public.http_request;
  END IF;
     
 	RES := public.http(FINAL_REQUEST)::public.http_response;
  RAISE NOTICE 'RES.status: %', RES.status;
  RAISE NOTICE 'RES.content: %', RES.content;

  DOC_ID := RES."content"::json->>'_id';

   --Need to check RES for response code
   IF RES.status != 201 THEN
   	RAISE EXCEPTION 'Error: %', RES;
   END IF;
  
  EXECUTE format('UPDATE %I.%I SET doc_id = DOC_ID WHERE id = $1', TG_TABLE_SCHEMA, TG_TABLE_NAME) USING NEW.id;
  RETURN null;
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



CREATE OR REPLACE TRIGGER zcouchdb_sync_members
AFTER INSERT OR UPDATE OF
bio, display_name, webfinger
ON public.members
FOR EACH ROW
EXECUTE PROCEDURE couchdb_put();

-- media 

CREATE OR REPLACE TRIGGER zcouchdb_sync_media
AFTER INSERT OR UPDATE OF
title, kind, created, added, modified
ON media.media
FOR EACH ROW
EXECUTE PROCEDURE couchdb_put();

CREATE OR REPLACE TRIGGER zcouchdb_sync_media_images
AFTER INSERT OR UPDATE ON media.media_images
FOR EACH ROW
EXECUTE PROCEDURE couchdb_put();

-- artists and studios

CREATE OR REPLACE TRIGGER zcouchdb_sync_artist_person
AFTER INSERT OR UPDATE OF
first_name, last_name, nick_names, roles, bio, added, modified
ON people.person
FOR EACH ROW
EXECUTE PROCEDURE couchdb_put();

CREATE OR REPLACE TRIGGER zcouchdb_sync_artist_group
AFTER INSERT OR UPDATE OF
name, active, bio, added, modified
ON people.group
FOR EACH ROW
EXECUTE PROCEDURE couchdb_put();

CREATE OR REPLACE TRIGGER zcouchdb_sync_studios
AFTER INSERT OR UPDATE OF
name, kind, city, added, modified
ON people.studio
FOR EACH ROW
  EXECUTE PROCEDURE couchdb_put();

-- genres

CREATE OR REPLACE TRIGGER zcouchdb_sync_genres
AFTER INSERT OR UPDATE OF
name, kinds
ON media.genres
FOR EACH ROW
  EXECUTE PROCEDURE couchdb_put();

CREATE OR REPLACE TRIGGER zcouchdb_sync_genre_descriptions
AFTER INSERT OR UPDATE OF
language, description, genre_id
ON media.genre_descriptions
FOR EACH ROW
  EXECUTE PROCEDURE couchdb_put();

CREATE OR REPLACE TRIGGER zcouchdb_sync_genre_characteristics
AFTER INSERT OR UPDATE OF
id, name, description
ON media.genre_characteristics
FOR EACH ROW
  EXECUTE PROCEDURE couchdb_put();

-- reviews

CREATE OR REPLACE TRIGGER zcouchdb_sync_reviews_basic
AFTER INSERT OR UPDATE OF
topic, body, user_id, media_id, added, modified
ON reviews.ratings
FOR EACH ROW
  EXECUTE PROCEDURE couchdb_put();

-- write the hitherto existing data to couchdb
UPDATE media.genres SET name = name;
UPDATE media.genre_descriptions SET language = language;
UPDATE public.members SET webfinger = webfinger;

CREATE TABLE places.city_name (
	id uuid NOT NULL REFERENCES places.city(uuid),
	name varchar NOT NULL,
	lang varchar NOT NULL
);

DROP TRIGGER zcouchdb_sync_genre_characteristics ON media.genre_characteristics;
DROP TRIGGER couchdb_genre_characteristics_deletion_sync ON media.genre_characteristics;

CREATE TYPE genre_description AS (
	description TEXT,
	language varchar(10)
);

CREATE OR REPLACE TRIGGER zcouchdb_sync_genres AFTER
INSERT
    OR
UPDATE
    OF name,
    id,
    kinds ON
    media.genres FOR EACH ROW EXECUTE FUNCTION couchdb_put();

CREATE OR REPLACE FUNCTION public.json_serialize(target_table TEXT, table_data RECORD)
 RETURNS jsonb
 LANGUAGE plpgsql IMMUTABLE
AS $function$
DECLARE
    DOC jsonb;
    FULL_ARTIST_NAME TEXT;
    REVIEWER_WF TEXT;
    REVIEW_MEDIA_TITLE TEXT;
    GENRE_DESCRIPTIONS public.genre_description[];
    GENRE_NAME text;
    GENRE_KINDS text[];
    CITY TEXT;
    ADDED timestamptz;
    MODIFIED timestamptz;
BEGIN
    MODIFIED := CURRENT_TIMESTAMP AT TIME ZONE 'UTC';
   IF NOT target_table = 'genres' OR target_table = 'members' THEN
  ADDED := to_timestamp(table_data.added) AT TIME ZONE 'UTC';
   END IF;
   CASE target_table
        WHEN 'genres' THEN
            GENRE_DESCRIPTIONS := ARRAY(SELECT ROW(description, language) FROM media."genre_descriptions" WHERE genre_id = table_data.id);
            DOC := jsonb_build_object('name', table_data.name, 'kinds', table_data.kinds, 'descriptions', jsonb_build_array(GENRE_DESCRIPTIONS));
        WHEN 'members' THEN
            DOC := jsonb_build_object('bio', table_data.bio, 'display_name', table_data.display_name, 'webfinger', table_data.webfinger);
        WHEN 'media' THEN
            DOC := jsonb_build_object('title', table_data.title, 'kind', table_data.kind, 
            'created', table_data.created, 'added', ADDED, 'modified', MODIFIED);
        WHEN 'person' THEN
            FULL_ARTIST_NAME := CONCAT(table_data.first_name, ' ', table_data.last_name);
            DOC := jsonb_build_object('name', FULL_ARTIST_NAME, 'nick_names', table_data.nick_names,
             'bio', table_data.bio, 'added', ADDED, 'modified', MODIFIED);
        WHEN 'group' THEN
            DOC := jsonb_build_object('name', table_data.name, 'active', table_data.active, 
            'bio', table_data.bio, 'added', ADDED, 'modified', MODIFIED);
        WHEN 'studio' THEN
            DOC := jsonb_build_object('name', table_data.name, 'kind', table_data.kind, 'city', table_data.city, 'added', ADDED, 'modified', MODIFIED);
        WHEN 'genre_descriptions' THEN
            GENRE_NAME := (SELECT name FROM media."genres" WHERE id = table_data.genre_id);
            GENRE_KINDS := (SELECT kinds FROM media."genres" WHERE id = table_data.genre_id);
            DOC := jsonb_build_object('name', GENRE_NAME, 'kinds', GENRE_KINDS,
             'descriptions', jsonb_build_array('language', table_data.language, 'description', table_data.description));
        WHEN 'ratings' THEN
            REVIEWER_WF := (SELECT webfinger FROM public.members WHERE uuid = table_data.user_id);
            REVIEW_MEDIA_TITLE := (SELECT title FROM media.media WHERE id = table_data.media_id);
            DOC := jsonb_build_object('topic', NEW.topic, 'body', NEW.body, 'user', webfinger, 'media_title', media_title, 'added', NEW.added, 'modified', NEW.modified);
    END CASE;
    RETURN DOC;
END;
$function$
;

CREATE OR REPLACE FUNCTION public.get_couchdb_target(localname TEXT)
  RETURNS text
  LANGUAGE plpgsql IMMUTABLE
AS $function$
DECLARE DBNAME text;
BEGIN
    CASE localname
        WHEN 'person', 'group' THEN 
            DBNAME := 'artists';
        WHEN 'genre_descriptions' THEN
            DBNAME := 'genres';
        ELSE
            DBNAME := localname;
    END CASE;
    RETURN DBNAME;
END;
$function$
;
    

CREATE OR REPLACE FUNCTION public.couchdb_put()
 RETURNS trigger
 LANGUAGE plpgsql
AS $function$
DECLARE
    DOC_ID text;
    REV_ID text;
	  AUTH text;
	 	DOC jsonb;
    RES public.http_response;
    COUCHDB_TARGET text;
	  FINAL_REQUEST public.http_request;
BEGIN
	-- it might seem stupid, but even a proper 
	-- alphabetic trigger naming did not work as expected
	-- and doc column contained NULL data
    DOC := public.json_serialize(TG_TABLE_NAME, NEW);
    COUCHDB_TARGET := public.get_couchdb_target(TG_TABLE_NAME);
	-- the actual request starts here
 AUTH := encode('librate:librate', 'base64');
    EXECUTE format('SELECT doc_id FROM %I.%I WHERE id = $1', TG_TABLE_SCHEMA, TG_TABLE_NAME) INTO DOC_ID USING NEW.id;
  IF DOC_ID IS NOT NULL THEN
    -- first, get the rev_id from couchdb to avoid conflicts
    SELECT content::json->>'_rev' FROM http_get('librate-search:5984/' || COUCHDB_TARGET || '/' || DOC_ID) INTO REV_ID;
   RAISE NOTICE 'REV_ID: %', REV_ID;
    FINAL_REQUEST := (
    'PUT',
    'http://172.20.0.6:5984/' || COUCHDB_TARGET || '/' || NEW.doc_id::text, 
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
$function$
;

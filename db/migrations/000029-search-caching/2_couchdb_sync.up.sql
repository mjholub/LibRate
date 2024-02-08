CREATE SCHEMA contributors;
-- not used, we store that in public.members
DROP SCHEMA members;

CREATE TABLE contributors.media (
contributor varchar NOT NULL REFERENCES public.members("webfinger") ON DELETE CASCADE,
media_id uuid NOT NULL REFERENCES media.media("id") ON DELETE CASCADE);

CREATE INDEX idx_contributor_media_id ON contributors.media (contributor, media_id);

CREATE TABLE contributors.person (
  contributor varchar NOT NULL REFERENCES public.members("webfinger") ON DELETE CASCADE,
  person_id uuid NOT NULL REFERENCES people.person("id") ON DELETE CASCADE);

CREATE INDEX idx_contributor_person_id ON contributors.person (contributor, person_id);

CREATE TABLE contributors.group (
  contributor varchar NOT NULL REFERENCES public.members("webfinger") ON DELETE CASCADE,
  group_id uuid NOT NULL REFERENCES people.group("id") ON DELETE CASCADE);

CREATE INDEX idx_contributor_group_id ON contributors.group (contributor, group_id);

CREATE TABLE contributors.studio (
  contributor varchar NOT NULL REFERENCES public.members("webfinger") ON DELETE CASCADE,
  studio_id int4 NOT NULL REFERENCES people.studio("id") ON DELETE CASCADE);

CREATE INDEX idx_contributor_studio_id ON contributors.studio (contributor, studio_id);

ALTER TABLE reviews.ratings
RENAME COLUMN "comment" TO "body";

ALTER TABLE reviews.ratings
DROP COLUMN created_at;

CREATE OR REPLACE FUNCTION couchdb_put() RETURNS trigger AS $BODY$
DECLARE
    RES RECORD;
BEGIN
 IF (NEW.from_pg) IS NULL THEN
   RETURN NEW;
 ELSE 
  -- note that this works only in docker
   SELECT status FROM http_post('http://librate-search:5984/' || TG_TABLE_NAME || '/' || NEW.id::text, '', NEW.doc::text, 'application/json'::text) INTO RES;    

   --Need to check RES for response code
   RAISE EXCEPTION 'Result: %', RES;
   RETURN null;
 END IF;
END;
$BODY$
LANGUAGE plpgsql VOLATILE  

CREATE TRIGGER couchdb_sync
AFTER INSERT OR UPDATE OR DELETE OF
bio, display_name, webfinger
ON public.members
OR INSERT OR UPDATE OR DELETE OF
title, kind, created, added, modified 
ON media.media
OR INSERT OR UPDATE OR DELETE OF
first_name, last_name, 
nick_names, roles, bio, added, modified
ON people.person
OR INSERT OR UPDATE OR DELETE OF
name, active, bio, added, modified
ON people.group
OR INSERT OR UPDATE OR DELETE OF
name, active, kind, added, modified 
ON people.studio
OR INSERT OR UPDATE OR DELETE OF
name, kind, country ON 
places.place
OR INSERT OR UPDATE OR DELETE OF
name, kinds ON
media.genre
OR INSERT OR UPDATE OR DELETE OF
language, description, genre_id ON
media.genre_description
OR INSERT OR UPDATE OR DELETE OF
id, name, description
ON media.genre_characteristics
OR INSERT OR UPDATE OR DELETE OF
topic, body, user_id, media_id, created_at
ON reviews.ratings
FOR EACH ROW
EXECUTE PROCEDURE couchdb_put();


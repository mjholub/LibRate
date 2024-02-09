-- create a function to write to the doc column for public.members
CREATE OR REPLACE FUNCTION member_to_jsonb() RETURNS trigger AS $BODY$ 
BEGIN
  NEW.doc = jsonb_build_object('bio', NEW.bio, 'display_name', NEW.display_name, 'webfinger', NEW.webfinger);
  RETURN NEW;
END;
$BODY$
LANGUAGE plpgsql VOLATILE;

CREATE OR REPLACE TRIGGER member_write_searchable_data
AFTER INSERT OR UPDATE OF
bio, display_name, webfinger
ON public.members
FOR EACH ROW
EXECUTE PROCEDURE member_to_jsonb();

CREATE OR REPLACE FUNCTION media_to_jsonb() RETURNS trigger AS $BODY$
BEGIN
  NEW.doc = jsonb_build_object('title', NEW.title, 'kind', NEW.kind, 'created', NEW.created, 'added', NEW.added, 'modified', NEW.modified);
  RETURN NEW;
END;
$BODY$
LANGUAGE plpgsql VOLATILE;

CREATE OR REPLACE TRIGGER media_write_searchable_data
AFTER INSERT OR UPDATE OF
title, kind, created, added, modified
ON media.media
FOR EACH ROW
EXECUTE PROCEDURE media_to_jsonb();

CREATE OR REPLACE FUNCTION person_to_jsonb() RETURNS trigger AS $BODY$
DECLARE
  name text;
BEGIN
  name := CONCAT(NEW.first_name, ' ', NEW.last_name);
    NEW.doc = jsonb_build_object('name', name, 'nick_names', NEW.nick_names, 'bio', NEW.bio, 'added', NEW.added, 'modified', NEW.modified);
  RETURN NEW;
END;
$BODY$
LANGUAGE plpgsql VOLATILE;

CREATE OR REPLACE TRIGGER person_write_searchable_data
AFTER INSERT OR UPDATE OF
first_name, last_name, nick_names, roles, bio, added, modified
ON people.person
FOR EACH ROW
EXECUTE PROCEDURE person_to_jsonb();

CREATE OR REPLACE FUNCTION group_to_jsonb() RETURNS trigger AS $BODY$
BEGIN
  NEW.doc = jsonb_build_object('name', NEW.name, 'active', NEW.active, 'bio', NEW.bio, 'added', NEW.added, 'modified', NEW.modified);
  RETURN NEW;
END;
$BODY$
LANGUAGE plpgsql VOLATILE;

CREATE OR REPLACE TRIGGER group_write_searchable_data
AFTER INSERT OR UPDATE OF
name, active, bio, added, modified
ON people.group
FOR EACH ROW
EXECUTE PROCEDURE group_to_jsonb();

CREATE OR REPLACE FUNCTION studio_to_jsonb() RETURNS trigger AS $BODY$
BEGIN
  NEW.doc = jsonb_build_object('name', NEW.name, 'kind', NEW.kind, 'city', NEW.city, 'added', NEW.added, 'modified', NEW.modified);
  RETURN NEW;
END;
$BODY$
LANGUAGE plpgsql VOLATILE;

CREATE OR REPLACE TRIGGER studio_write_searchable_data
AFTER INSERT OR UPDATE OF
name, kind, city, added, modified
ON people.studio
FOR EACH ROW
EXECUTE PROCEDURE studio_to_jsonb();

CREATE OR REPLACE FUNCTION genre_to_jsonb() RETURNS trigger AS $BODY$
BEGIN
  NEW.doc = jsonb_build_object('name', NEW.name, 'kinds', NEW.kinds);
  RETURN NEW;
END;
$BODY$
LANGUAGE plpgsql VOLATILE;

CREATE OR REPLACE TRIGGER genre_write_searchable_data
AFTER INSERT OR UPDATE OF
name, kinds
ON media.genres
FOR EACH ROW
EXECUTE PROCEDURE genre_to_jsonb();

CREATE OR REPLACE FUNCTION genre_description_to_jsonb() RETURNS trigger AS $BODY$
BEGIN
  NEW.doc = jsonb_build_object('language', NEW.language, 'description', NEW.description, 'genre_id', NEW.genre_id);
  RETURN NEW;
END;
$BODY$
LANGUAGE plpgsql VOLATILE;

CREATE OR REPLACE TRIGGER genre_description_write_searchable_data
AFTER INSERT OR UPDATE OF
language, description, genre_id
ON media.genre_descriptions
FOR EACH ROW
EXECUTE PROCEDURE genre_description_to_jsonb();

CREATE OR REPLACE FUNCTION genre_characteristics_to_jsonb() RETURNS trigger AS $BODY$
BEGIN
  NEW.doc = jsonb_build_object('id', NEW.id, 'name', NEW.name, 'description', NEW.description);
  RETURN NEW;
END;
$BODY$
LANGUAGE plpgsql VOLATILE;

CREATE OR REPLACE TRIGGER genre_characteristics_write_searchable_data
AFTER INSERT OR UPDATE OF
id, name, description
ON media.genre_characteristics
FOR EACH ROW
EXECUTE PROCEDURE genre_characteristics_to_jsonb();

CREATE OR REPLACE FUNCTION rating_to_jsonb() RETURNS trigger AS $BODY$
DECLARE
  webfinger text;
  media_title text;
BEGIN
  webfinger := (SELECT webfinger FROM public.members WHERE uuid = NEW.user_id);
  media_title := (SELECT title FROM media.media WHERE id = NEW.media_id);
  NEW.doc = jsonb_build_object('topic', NEW.topic, 'body', NEW.body, 'user_id', webfinger, 'media_id', media_title, 'added', NEW.added, 'modified', NEW.modified);
  RETURN NEW;
END;
$BODY$
LANGUAGE plpgsql VOLATILE;

CREATE OR REPLACE TRIGGER rating_write_searchable_data
AFTER INSERT OR UPDATE OF
topic, body, user_id, media_id, added, modified
ON reviews.ratings
FOR EACH ROW
EXECUTE PROCEDURE rating_to_jsonb();

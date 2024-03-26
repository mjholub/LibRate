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

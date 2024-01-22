-- INFO: Recreating foreign keys referencing people.person(id) to use UUIDs
ALTER TABLE media.album_artists ADD CONSTRAINT album_artists_artist_fkey FOREIGN KEY (artist) REFERENCES people.person(id) ON DELETE CASCADE;

ALTER TABLE media.media_creators DROP COLUMN creator_id;
ALTER TABLE media.media_creators ADD COLUMN creator_id uuid NULL REFERENCES people.person(id) ON DELETE CASCADE;

ALTER TABLE people.person_photos DROP COLUMN person_id;
ALTER TABLE people.person_photos ADD COLUMN person_id uuid NULL REFERENCES people.person(id) ON DELETE CASCADE;

ALTER TABLE people.person_works DROP COLUMN person_id;
ALTER TABLE people.person_works ADD COLUMN person_id uuid NULL REFERENCES people.person(id) ON DELETE CASCADE;

ALTER TABLE media.track_artists DROP COLUMN artist;
ALTER TABLE media.track_artists ADD COLUMN artist uuid NULL REFERENCES people.person(id) ON DELETE CASCADE;

ALTER TABLE media.tv_show_cast DROP COLUMN person;
ALTER TABLE media.tv_show_cast ADD COLUMN person uuid NULL REFERENCES people.person(id) ON DELETE CASCADE;

ALTER TABLE people.group_members DROP COLUMN person_id;
ALTER TABLE people.group_members ADD COLUMN person_id uuid NULL REFERENCES people.person(id) ON DELETE CASCADE;

ALTER TABLE people.studio_artists DROP COLUMN person_id;
ALTER TABLE people.studio_artists ADD COLUMN person_id uuid NULL REFERENCES people.person(id) ON DELETE CASCADE;

ALTER TABLE people.actor_cast DROP COLUMN person_id;
ALTER TABLE people.actor_cast ADD COLUMN person_id uuid NULL REFERENCES people.person(id) ON DELETE CASCADE;

ALTER TABLE people.director_cast DROP COLUMN person_id;
ALTER TABLE people.director_cast ADD COLUMN person_id uuid NULL REFERENCES people.person(id) ON DELETE CASCADE;

ALTER TABLE media.book_authors DROP COLUMN person;
ALTER TABLE media.book_authors ADD COLUMN person uuid NULL REFERENCES people.person(id) ON DELETE CASCADE;

ALTER TABLE people.group DROP COLUMN id CASCADE;
--constraint group_locations_group_id_fkey on table people.group_locations depends on column id of table people."group"
--constraint group_photos_group_id_fkey on table people.group_photos depends on column id of table people."group"
--constraint group_members_group_id_fkey on table people.group_members depends on column id of table people."group"
--constraint group_genres_group_id_fkey on table people.group_genres depends on column id of table people."group"
--constraint group_works_group_id_fkey on table people.group_works depends on column id of table people."group"
--constraint album_artists_artist_group_fkey on table media.album_artists depends on column id of table people."group"
ALTER TABLE people.group ADD COLUMN id uuid DEFAULT uuid_time_nextval(30,65536) NOT NULL;
ALTER TABLE people."group" ADD CONSTRAINT group_pk PRIMARY KEY (id);

ALTER TABLE people.group_locations DROP COLUMN group_id;
ALTER TABLE people.group_locations ADD COLUMN group_id uuid NOT NULL;
ALTER TABLE people.group_locations ADD CONSTRAINT group_locations_group_id_fkey FOREIGN KEY (group_id) REFERENCES people."group"(id) ON DELETE CASCADE;

ALTER TABLE people.group_photos DROP COLUMN group_id;
ALTER TABLE people.group_photos ADD COLUMN group_id uuid NOT NULL REFERENCES people."group"(id) ON DELETE CASCADE;

ALTER TABLE people.group_members DROP COLUMN group_id;
ALTER TABLE people.group_members ADD COLUMN group_id uuid NOT NULL REFERENCES people."group"(id) ON DELETE CASCADE;

ALTER TABLE people.group_genres DROP COLUMN group_id;
ALTER TABLE people.group_genres ADD COLUMN group_id uuid NOT NULL REFERENCES people."group"(id) ON DELETE CASCADE;

ALTER TABLE people.group_works DROP COLUMN group_id;
ALTER TABLE people.group_works ADD COLUMN group_id uuid NOT NULL REFERENCES people."group"(id) ON DELETE CASCADE;

ALTER TABLE media.album_artists ADD CONSTRAINT album_artists_artist_group_fkey FOREIGN KEY (artist) REFERENCES people."group"(id) ON DELETE CASCADE;

-- INFO: people.person recreation with uuid as id
ALTER TABLE people.person ALTER COLUMN id TYPE uuid USING id::uuid::uuid;
ALTER TABLE people.person ALTER COLUMN id SET DEFAULT uuid_time_nextval(30,65536);

DROP TABLE IF EXISTS people.person CASCADE;

-- constraint media_creators_creator_id_fkey on table media.media_creators depends on table people.person
-- constraint media_creator_fkey on table media.media depends on table people.person
-- NOTE: for the two fkeys above, we recreate only one, since we're already using a junction table

-- constraint person_photos_person_id_fkey on table people.person_photos depends on table people.person
-- constraint person_works_person_id_fkey on table people.person_works depends on table people.person
-- constraint track_artists_artist_fkey on table media.track_artists depends on table people.person
-- constraint tv_show_cast_person_fkey on table media.tv_show_cast depends on table people.person
-- constraint group_members_person_id_fkey on table people.group_members depends on table people.person
-- constraint studio_artists_person_id_fkey on table people.studio_artists depends on table people.person
-- constraint actor_cast_person_id_fkey on table people.actor_cast depends on table people.person
-- constraint director_cast_person_id_fkey on table people.director_cast depends on table people.person
-- constraint book_authors_person_fkey on table media.book_authors depends on table people.person
-- constraint album_artists_artist_fkey on table media.album_artists depends on table people.person

CREATE TABLE people.person (
	-- use sequential uuids for better caching
    id uuid DEFAULT uuid_time_nextval(30,65536) NOT NULL,
    first_name varchar(255) NOT NULL,
    other_names _varchar NULL,
    last_name varchar(255) NOT NULL,
    nick_names _varchar NULL,
    roles people."_role" NULL,
    birth date NULL,
    death date NULL,
    website varchar(255) NULL,
    bio text NULL,
    added timestamp NOT NULL DEFAULT now(),
    modified timestamp NULL DEFAULT now(),
    CONSTRAINT person_pkey PRIMARY KEY (id)
);

CREATE OR REPLACE FUNCTION update_modified_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.modified = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;


CREATE TRIGGER person_update_modified
BEFORE UPDATE ON people.person
FOR EACH ROW
EXECUTE FUNCTION update_modified_column();

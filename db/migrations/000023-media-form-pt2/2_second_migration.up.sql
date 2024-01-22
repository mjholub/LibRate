-- NOTE: media.album_artists update to use single artist column (1/2)
ALTER TABLE media.album_artists DROP CONSTRAINT album_artists_pkey;
ALTER TABLE media.album_artists DROP CONSTRAINT album_artist_check;
ALTER TABLE media.album_artists DROP CONSTRAINT album_artists_group_artist_fkey;
ALTER TABLE media.album_artists DROP COLUMN group_artist;
ALTER TABLE media.album_artists DROP CONSTRAINT album_artists_person_artist_fkey;
ALTER TABLE media.album_artists DROP COLUMN person_artist;

ALTER TABLE media.album_artists ADD COLUMN artist uuid NULL;

-- artist can either reference a person or a group
ALTER TABLE media.album_artists ADD CONSTRAINT album_artists_pkey PRIMARY KEY (album, artist);

CREATE TYPE media."artist_type" AS ENUM ('individual', 'group');

ALTER TABLE media.album_artists ADD COLUMN artist_type media.artist_type NOT NULL DEFAULT 'individual';

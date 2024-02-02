-- partial index for primary genres, so that we can build our genre pages structure
CREATE INDEX primary_genre_idx ON media.genres (name) WHERE parent IS NULL;

-- UNIX time to be used when populating cache mostly
ALTER TABLE public.members ADD COLUMN updated_at BIGINT NOT NULL DEFAULT 0;


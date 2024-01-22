CREATE TYPE genre_kind AS ENUM ('music', 'film', 'tv', 'book', 'game');
ALTER TABLE media.genres ADD COLUMN kind genre_kind NOT NULL DEFAULT 'music';

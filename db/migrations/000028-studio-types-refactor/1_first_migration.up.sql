CREATE TYPE people.studio_type AS ENUM ('film', 'music', 'game', 'tv', 'publishing', 'visual_art_other', 'unknown');

ALTER TABLE people.studio
DROP COLUMN is_film,
DROP COLUMN is_music,
DROP COLUMN is_game,
DROP COLUMN is_tv,
DROP COLUMN is_publishing,

ADD COLUMN kind people.studio_type NOT NULL DEFAULT 'unknown'; 

ALTER TABLE media.tracks DROP CONSTRAINT fk_tracks_album;
ALTER TABLE media.tracks ALTER COLUMN duration TYPE interval 
ALTER TABLE media.tracks ALTER COLUMN media_id SET DEFAULT uuid_generate_v4();


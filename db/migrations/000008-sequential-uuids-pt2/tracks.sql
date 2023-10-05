ALTER TABLE media.tracks ALTER COLUMN media_id SET DEFAULT uuid_time_nextval();
ALTER TABLE media.tracks
ADD CONSTRAINT fk_tracks_album
FOREIGN KEY ("album")
REFERENCES media.albums ("media_id");
ALTER TABLE media.tracks ALTER COLUMN duration TYPE time USING duration::time;

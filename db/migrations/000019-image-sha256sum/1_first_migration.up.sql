-- nullable and not unique because we only use this so that we don't waste computing resources
-- if an user tries to overwrite their profile picture with the same image
ALTER TABLE cdn.images ADD COLUMN sha256sum CHAR(64) NULL;

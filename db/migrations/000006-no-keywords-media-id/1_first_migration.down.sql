ALTER TABLE media.keywords ADD media_id UUID NOT NULL references media.media(media_id);

ALTER TABLE cdn.images ADD uploader varchar NOT NULL;
COMMENT ON COLUMN cdn.images.uploader IS 'not sure whether to use fkey here since unsure how that''ll work with federation';
ALTER TABLE cdn.images ALTER COLUMN thumbnail DROP NOT NULL;
-- TODO: this seems to already be set elsewhere, check where
-- ALTER TABLE public.members ADD CONSTRAINT members_un UNIQUE (nick);

-- UNIX time to be used when populating cache mostly
ALTER TABLE public.members ADD COLUMN updated_at BIGINT NOT NULL DEFAULT 0;


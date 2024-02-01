ALTER TABLE public.bans ADD COLUMN banned_by varchar NOT NULL REFERENCES public.members("webfinger");

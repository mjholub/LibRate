ALTER TABLE public.members DROP COLUMN homepage;
ALTER TABLE public.members DROP COLUMN irc;
ALTER TABLE public.members DROP COLUMN xmpp;
ALTER TABLE public.members DROP COLUMN matrix;

ALTER TABLE public.members ADD custom_fields jsonb NULL;

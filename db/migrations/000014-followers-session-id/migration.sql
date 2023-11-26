ALTER TABLE public.members ADD COLUMN followers text NOT NULL DEFAULT '';
ALTER TABLE public.members ADD COLUMN session_timeout bigint NULL;
ALTER TABLE public.members ADD COLUMN public_key_pem TEXT NOT NULL DEFAULT '';
ALTER TABLE public.members ALTER COLUMN visibility SET DEFAULT 'public'::text;
UPDATE public.members SET visibility = 'public' WHERE visibility IS NULL; 
ALTER TABLE public.members RENAME COLUMN "following" TO following_uri;
ALTER TABLE public.members RENAME COLUMN followers TO followers_uri;
ALTER TABLE public.members ALTER COLUMN following_uri TYPE text USING following_uri::text;
ALTER TABLE public.members ALTER COLUMN following_uri SET DEFAULT '';
ALTER TYPE public."role" ADD VALUE 'member';

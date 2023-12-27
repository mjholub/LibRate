ALTER TABLE public.members DROP COLUMN followers;
ALTER TABLE public.members DROP COLUMN session_timeout;
ALTER TABLE public.members DROP COLUMN public_key_pem;
ALTER TABLE public.members ALTER COLUMN visibility SET DEFAULT 'private'::text;
ALTER TABLE public.members RENAME COLUMN following_uri TO "following";
ALTER TABLE public.members RENAME COLUMN followers_uri TO "followers";
ALTER TABLE public.members ALTER COLUMN "following" TYPE jsonb USING "following"::jsonb;
ALTER TABLE public.members ALTER COLUMN "following" SET DEFAULT '[]'::jsonb;
ALTER TYPE public."role" DROP VALUE 'member';

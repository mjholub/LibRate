ALTER TABLE public.members ADD COLUMN following jsonb NOT NULL DEFAULT '[]'::jsonb;
ALTER TABLE public.members ADD COLUMN visibility text NOT NULL DEFAULT 'private';
-- retroactively set visibility to private for all existing members
UPDATE public.members SET visibility = 'private' WHERE visibility IS NULL; 

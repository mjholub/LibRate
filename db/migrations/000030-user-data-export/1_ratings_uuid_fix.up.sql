ALTER TABLE reviews.ratings DROP COLUMN IF EXISTS user_id CASCADE;

ALTER TABLE reviews.ratings ADD COLUMN user_id uuid 
NOT NULL REFERENCES public.members(uuid) ON DELETE CASCADE;

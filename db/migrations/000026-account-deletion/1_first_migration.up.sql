ALTER TABLE reviews.cast_ratings DROP CONSTRAINT cast_ratings_user_id_fkey;
ALTER TABLE reviews.cast_ratings ADD CONSTRAINT cast_ratings_user_id_fkey FOREIGN KEY (user_id) REFERENCES public.members(id) ON DELETE CASCADE;
ALTER TABLE reviews.track_ratings DROP CONSTRAINT track_ratings_user_id_fkey;
ALTER TABLE reviews.track_ratings ADD CONSTRAINT track_ratings_user_id_fkey FOREIGN KEY (user_id) REFERENCES public.members(id) ON DELETE CASCADE;
ALTER TABLE reviews.ratings DROP CONSTRAINT ratings_user_id_fkey;
ALTER TABLE reviews.ratings ADD CONSTRAINT ratings_user_id_fkey FOREIGN KEY (user_id) REFERENCES public.members(id) ON DELETE CASCADE;
ALTER TABLE public.bans DROP CONSTRAINT bans_member_uuid_fkey;
ALTER TABLE public.bans ADD CONSTRAINT bans_member_uuid_fkey FOREIGN KEY (member_uuid) REFERENCES public.members("uuid") ON DELETE CASCADE;

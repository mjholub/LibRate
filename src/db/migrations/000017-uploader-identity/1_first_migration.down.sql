ALTER TABLE public.members DROP CONSTRAINT members_un;
ALTER TABLE cdn.images ALTER COLUMN thumbnail SET NOT NULL; 
ALTER TABLE cdn.images DROP COLUMN uploader;

ALTER TABLE public.members ALTER COLUMN id 
SET DEFAULT uuid_time_nextval(30,65536);

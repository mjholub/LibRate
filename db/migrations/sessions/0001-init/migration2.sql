ALTER TABLE public."default" ADD authenticated boolean NOT NULL DEFAULT false;
ALTER TABLE public."default" ADD expires timestamptz NULL DEFAULT CURRENT_TIMESTAMP + INTERVAL '30 minutes';
ALTER TABLE public."default" ALTER COLUMN member_name DROP DEFAULT;
ALTER TABLE public."default" ALTER COLUMN created SET DEFAULT CURRENT_TIMESTAMP;

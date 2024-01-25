ALTER TABLE public.members DROP COLUMN followers_uri;

CREATE TABLE public.followers (
	followee varchar NOT NULL,
	follower varchar NOT NULL,
	CONSTRAINT followers_pk PRIMARY KEY (follower,followee),
  FOREIGN KEY (followee) REFERENCES public.members (webfinger) ON DELETE CASCADE,
  FOREIGN KEY (follower) REFERENCES public.members (webfinger) ON DELETE CASCADE
);

-- WARN: not suer if this should be not null
ALTER TABLE public.members ADD webfinger varchar NOT NULL;
ALTER TABLE public.members ADD CONSTRAINT members_unique_1 UNIQUE (webfinger) DEFERRABLE INITIALLY DEFERRED;

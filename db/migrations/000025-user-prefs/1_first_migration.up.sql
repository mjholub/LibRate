CREATE TABLE public.member_prefs (
	member_id int4 NOT NULL,
	auto_accept_follow bool NOT NULL DEFAULT true,
	locally_searchable bool NOT NULL DEFAULT true,
	robots_searchable bool NOT NULL DEFAULT false,
  blur_nsfw bool NOT NULL DEFAULT true,
	rating_scale_lower int2 NOT NULL DEFAULT 1,
	rating_scale_upper int2 NOT NULL DEFAULT 10,
	searchable_to_federated bool NOT NULL DEFAULT true,
	message_autohide_words _text NULL,
	muted_instances _text NULL,
	CONSTRAINT member_prefs_pk PRIMARY KEY (member_id)
);
  
ALTER table public.member_prefs ADD FOREIGN KEY (member_id) REFERENCES public.members(id) ON DELETE CASCADE;

-- validation will be done in application layer
CREATE TABLE public.follow_requests (
  id serial8 NOT NULL PRIMARY KEY,
  reblogs bool NOT NULL DEFAULT true,
  notifications bool NOT NULL DEFAULT false,
	requester_webfinger varchar NOT NULL,
	target_webfinger varchar NOT NULL,
	created timestamptz NOT NULL DEFAULT now()
);

CREATE TABLE public.member_blocks (
	id serial8 NOT NULL,
	requester_webfinger varchar NOT NULL,
	target_webfinger varchar NOT NULL,
	created timestamptz NOT NULL,
	CONSTRAINT member_blocks_pk PRIMARY KEY (id)
);

DROP TABLE public.followers;

CREATE TABLE public.followers (
	follower varchar NOT NULL,
	followee varchar NOT NULL,
  reblogs bool NOT NULL DEFAULT true,
  notifications bool NOT NULL DEFAULT false,
	created timestamptz NOT NULL DEFAULT now(),
	CONSTRAINT followers_pk PRIMARY KEY (follower)
);

CREATE INDEX members_uuid_idx ON members (uuid);


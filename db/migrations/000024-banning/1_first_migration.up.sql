ALTER TABLE public.members ADD CONSTRAINT members_unique UNIQUE (uuid);

CREATE TABLE public.bans (
	member_uuid uuid NOT NULL,
	reason text NOT NULL,
	ends time with time zone NOT NULL,
	can_appeal boolean NOT NULL DEFAULT true,
	mask inet NULL,
	occurrence smallint NOT NULL DEFAULT 1,
	started time with time zone NOT NULL DEFAULT now(),
	CONSTRAINT bans_pk PRIMARY KEY (member_uuid),
	FOREIGN KEY (member_uuid) REFERENCES public.members(uuid)
);


CREATE DATABASE librate_sessions ENCODING 'UTF8' LC_COLLATE 'en_US.UTF-8' LC_CTYPE 'en_US.UTF-8' TEMPLATE template0 OWNER = librate;

CREATE TABLE public."default" (
	session_id uuid NOT NULL,
	member_name varchar NOT NULL,
	created timestamptz NOT NULL,
	device_id uuid NOT NULL,
	ip inet NOT NULL,
	user_agent varchar NOT NULL,
	CONSTRAINT default_pk PRIMARY KEY (session_id)
);


CREATE TABLE cdn.videos (
	id bigserial NOT NULL,
	"source" varchar(255) NOT NULL,
	thumbnail varchar(255) NULL, -- TODO: implement auto-thumbnail functionality
	alt varchar(255) null,
	CONSTRAINT videos_pkey PRIMARY KEY (id)
);

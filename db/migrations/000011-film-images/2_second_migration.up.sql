CREATE TABLE media.film_posters (
	film_id uuid NOT NULL,
	image_id int8 NOT NULL,
	country_id int2 null,
	CONSTRAINT film_posters_pkey PRIMARY KEY (film_id, image_id)
);

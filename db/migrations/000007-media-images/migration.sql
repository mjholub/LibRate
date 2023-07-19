CREATE TABLE media.media_images (
	media_id uuid NOT NULL,
	image_id bigint NOT NULL,
	is_main boolean not null default false,
	CONSTRAINT media_media_images_pkey PRIMARY KEY (media_id, image_id),
	CONSTRAINT media_media_images_id_fkey FOREIGN KEY (image_id) REFERENCES cdn.images(id),
	CONSTRAINT media_media_images_media_id_fkey FOREIGN KEY (media_id) REFERENCES media.media(id)
);

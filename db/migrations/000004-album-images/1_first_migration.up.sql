CREATE TABLE media.album_images (
album_id UUID not null references media.albums(media_id),
image_id BIGINT not null references cdn.images(id),
  PRIMARY KEY (album_id, image_id)
);

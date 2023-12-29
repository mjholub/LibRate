ALTER TABLE reviews.ratings DROP CONSTRAINT ratings_user_id_fkey;

ALTER TABLE reviews.ratings RENAME TO ratings_temp;

CREATE TABLE reviews.ratings (
    id bigserial NOT NULL PRIMARY KEY,
    stars int2 NOT NULL,
    "comment" text NULL,
    topic text NULL,
    attribution text NULL,
    user_id serial4 NOT NULL,
    media_id uuid NULL,
    created_at timestamptz NOT NULL DEFAULT now()
);

INSERT INTO reviews.ratings (
  stars, "comment",
  topic, attribution, user_id, media_id, created_at)
SELECT stars, "comment",
  topic, attribution, user_id, media_id, created_at
FROM reviews.ratings_temp;

DROP TABLE reviews.ratings_temp;

CREATE TABLE reviews.ratings_temp (
    new_id bigserial NOT NULL PRIMARY KEY,
    stars int2 NOT NULL,
    "comment" text NULL,
    topic text NULL,
    attribution text NULL,
    user_id serial4 NOT NULL,
    media_id uuid NULL,
    created_at timestamptz NOT NULL DEFAULT now()
);

-- 2. Copy the data from the original table to the temporary table
INSERT INTO reviews.ratings_temp (stars, "comment", topic, attribution, user_id, media_id, created_at)
SELECT stars, "comment", topic, attribution, user_id, media_id, created_at
FROM reviews.ratings;

-- 3. Drop the original table
DROP TABLE reviews.ratings;

-- 4. Rename the temporary table to the original table name
ALTER TABLE reviews.ratings_temp RENAME TO ratings;

ALTER TABLE reviews.ratings
ADD CONSTRAINT ratings_user_id_fkey FOREIGN KEY (user_id) REFERENCES public.members(id);

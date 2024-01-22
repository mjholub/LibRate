ALTER TABLE reviews.ratings
ADD COLUMN created_at timestamp
WITH time zone NOT NULL DEFAULT now();

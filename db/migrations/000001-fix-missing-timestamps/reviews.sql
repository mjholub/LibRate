ALTER TABLE reviews.ratings ADD COLUMN created_at timestamp with time zone NOT NULL DEFAULT now();

-- models.GetGenres() in application layer
CREATE index if not exists idx_genres_kinds ON media.genres USING GIN (kinds);

-- models.GetGenre() in application layer
CREATE index if not exists idx_genres_kinds_name ON media.genres USING GIN (kinds);
CREATE index if not exists idx_genre_descriptions_genre_id ON media.genre_descriptions (genre_id);
CREATE index if not exists idx_genre_descriptions_language ON media.genre_descriptions (language);

-- for faster finding of genres with common characteristics
CREATE index if not exists idx_genres_children ON media.genres USING GIN (children);
CREATE INDEX if not exists idx_genres_parent ON media.genres (parent);
CREATE INDEX if not exists idx_genre_char_mapping_genre_id ON media.genre_characteristics_mapping (genre_id);
CREATE INDEX if not exists idx_genre_char_mapping_char_id ON media.genre_characteristics_mapping (characteristic_id);

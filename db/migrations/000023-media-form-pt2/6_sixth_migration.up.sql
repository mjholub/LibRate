INSERT INTO media.genre_descriptions (genre_id, description, language)
SELECT id, 'TODO', 'en'
FROM media.genres
WHERE parent IS NULL AND id NOT IN (SELECT genre_id FROM media.genre_descriptions WHERE language = 'en');

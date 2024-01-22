-- We already have a lot of junction tables. While this seems like a good idea
-- for mapping many-to-many relationships, it makes the whole database more bloated, harder to maintain
-- and harder to query.
-- Instead, we'll use people.cast column, which contains an ID (bigserial) and the media ID (uuid).
-- Then, a media.films column can reference the ID of said cast.
-- For that cast, there are junction tables in the people schema: `actors_cast` and `directors_cast`.
-- So for example, when we want to get all the actors in a movie, we can do:
-- SELECT * FROM people.actors_cast
-- WHERE cast_id IN
-- (SELECT cast_id FROM media.films WHERE media_id = 'uuid');
DROP TABLE media.film_cast;
-- we don't really need a non-null media."media"(creator) column
-- This was quite a bad idea. Many works have more than one author 
-- and generally speaking, it panders to things like giving a director
-- most credit for creating a movie, which many artists consider a major example of stolen valor.
-- tl;dr/TODO: gradually move away from assigning a 'creator' to a media object and rely
-- on the media_authors junction table instead. This is a bit more tricky to load, but
-- way more accurate and flexible.
ALTER TABLE media.media ALTER COLUMN creator DROP NOT NULL;


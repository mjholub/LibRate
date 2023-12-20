ALTER TABLE media.book_genres ALTER COLUMN genre TYPE SMALLSERIAL NOT NULL REFERENCES media.genres(id);
ALTER TABLE media.book_authors ALTER COLUMN person TYPE SERIAL NOT NULL REFERENCES people.person(id)

ALTER TABLE media.books DROP CONSTRAINT books_publisher_fkey;
ALTER TABLE media.books ALTER COLUMN publisher TYPE SERIAL REFERENCES people.group(id); 

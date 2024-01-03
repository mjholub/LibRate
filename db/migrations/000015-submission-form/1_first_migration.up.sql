ALTER TABLE media.books ALTER COLUMN publisher TYPE int4 USING publisher::int4;

COMMENT ON TABLE people.studio IS 'Publishing houses also count as studios';

ALTER TABLE media.books DROP CONSTRAINT books_publisher_fkey;

ALTER TABLE media.books
ADD CONSTRAINT books_publisher_fkey
FOREIGN KEY (publisher)
REFERENCES people."studio"(id);


ALTER TABLE media.book_authors ALTER COLUMN person TYPE int4 USING person::int4;

ALTER TABLE media.book_genres ALTER COLUMN genre TYPE smallint USING genre::smallint;

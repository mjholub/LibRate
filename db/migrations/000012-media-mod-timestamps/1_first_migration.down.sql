CREATE TABLE IF NOT EXISTS media.film_cast (
			film UUID NOT NULL REFERENCES media.films(media_id),
			person SERIAL NOT NULL REFERENCES people.person(id),
			PRIMARY KEY (film, person)
		);

ALTER TABLE media.media ALTER COLUMN creator SET NOT NULL;

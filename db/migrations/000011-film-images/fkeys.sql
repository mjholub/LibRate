ALTER TABLE media.film_posters ADD CONSTRAINT film_posters_film_id_fkey FOREIGN KEY (film_id) REFERENCES media.films(media_id);
ALTER TABLE media.film_posters ADD CONSTRAINT film_posters_image_id_fkey FOREIGN KEY (image_id) REFERENCES cdn.images(id);
ALTER TABLE media.film_posters ADD CONSTRAINT film_posters_country_id_fkey foreign key (country_id) references places.country(id);

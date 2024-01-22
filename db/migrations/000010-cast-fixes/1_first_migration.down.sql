DROP TABLE IF EXISTS poeple.director_cast CASCADE;
DROP TABLE IF EXISTS poeple.actor_cast CASCADE;

ALTER TABLE people."cast" ADD COLUMN actors int4[] NULL;
ALTER TABLE people."cast" ADD COLUMN directors int4[] NULL;

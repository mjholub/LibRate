-- Drop columns actors and directors from people."cast"
ALTER TABLE people."cast" DROP COLUMN actors;
ALTER TABLE people."cast" DROP COLUMN directors;

-- Create the actor_cast junction table
CREATE TABLE people.actor_cast (
    cast_id bigserial NOT NULL,
    person_id int4 NOT NULL,
    CONSTRAINT actor_cast_pkey PRIMARY KEY (cast_id, person_id),
    FOREIGN KEY (cast_id) REFERENCES people."cast"(id),
    FOREIGN KEY (person_id) REFERENCES people.person(id)
);

-- Create the director_cast junction table
CREATE TABLE people.director_cast (
    cast_id bigserial NOT NULL,
    person_id int4 NOT NULL,
    CONSTRAINT director_cast_pkey PRIMARY KEY (cast_id, person_id),
    FOREIGN KEY (cast_id) REFERENCES people."cast"(id),
    FOREIGN KEY (person_id) REFERENCES people.person(id)
);

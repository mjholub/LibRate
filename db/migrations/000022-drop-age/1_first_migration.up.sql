CREATE TABLE media.genre_characteristics (
    id bigserial NOT NULL,
    "name" text NOT NULL,
    description text NULL,
    CONSTRAINT genre_characteristics_pk PRIMARY KEY (id), -- Add a primary key constraint
    CONSTRAINT genre_characteristics_name_key UNIQUE ("name"),
    CONSTRAINT genre_characteristics_description_key UNIQUE (description)
);

CREATE TABLE media.genre_characteristics_mapping (
    genre_id bigint NOT NULL,
    characteristic_id bigint NOT NULL,
    CONSTRAINT genre_characteristics_mapping_pk PRIMARY KEY (genre_id, characteristic_id),
    CONSTRAINT genre_characteristics_mapping_genre_fk FOREIGN KEY (genre_id) REFERENCES media.genres (id) ON DELETE CASCADE,
    CONSTRAINT genre_characteristics_mapping_characteristic_fk FOREIGN KEY (characteristic_id) REFERENCES media.genre_characteristics (id) ON DELETE CASCADE
);

ALTER TABLE media.genres ADD CONSTRAINT genres_unique UNIQUE ("name");
ALTER TABLE media.genres DROP COLUMN "characteristics";
ALTER TABLE media.genres DROP COLUMN description;

CREATE TABLE media.genre_descriptions (
    genre_id bigint NOT NULL,
    language varchar(10) NOT NULL,
    description text NULL,
    CONSTRAINT genre_descriptions_pk PRIMARY KEY (genre_id, language),
    CONSTRAINT genre_descriptions_fk FOREIGN KEY (genre_id) REFERENCES media.genres(id) ON DELETE CASCADE
);

ALTER TABLE media.genres ADD kinds _varchar NOT NULL;

ALTER TABLE people.group ADD COLUMN aliases varchar(255)[] DEFAULT '{}'::varchar(255)[] NOT NULL;
ALTER TABLE people.studio ADD COLUMN aliases varchar(255)[] DEFAULT '{}'::varchar(255)[] NOT NULL;

ALTER TABLE people.person ADD COLUMN name varchar(255) NOT NULL;
ALTER TABLE people.person ADD COLUMN aliases varchar(255)[] NOT NULL;

-- move data from first_nammes and last_names to name
UPDATE people.person SET name = CONCAT(first_name ' ', last_name);

-- move data from other_names and nick_names to aliases
UPDATE people.person SET aliases = ARRAY[other_names, nick_names];


-- drop old columns
ALTER TABLE people.personm DROP COLUMN first_name;
ALTER TABLE people.person DROP COLUMN last_name;
ALTER TABLE people.person DROP COLUMN other_names;
ALTER TABLE people.person DROP COLUMN nick_names;

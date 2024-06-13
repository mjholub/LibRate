-- 1. recreate old columns
ALTER TABLE people.person CREATE COLUMN first_name varchar(255) NOT NULL;
ALTER TABLE people.person CREATE COLUMN last_name varchar(255) NOT NULL;
ALTER TABLE people.person CREATE COLUMN other_names varchar(255)[];
ALTER TABLE people.person CREATE COLUMN nick_names varchar(255)[];

-- move data back to the old columns
UPDATE people.person SET first_name = SPLIT_PART(name, ' ', 1);
UPDATE people.person SET last_name = SPLIT_PART(name, ' ', 2);

-- copy array data
-- since relationship is not preserved, treat every alias as other_name
UPDATE people.person SET nick_names = aliases;

-- drop new columns
DROP COLUMN name;
DROP COLUMN aliases;

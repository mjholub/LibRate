ALTER TABLE media.genres ALTER COLUMN id TYPE bigint; 
ALTER TABLE media.genres DROP COLUMN desc_long CASCADE;
ALTER TABLE media.genres RENAME COLUMN desc_short TO description;
ALTER TABLE media.genres RENAME COLUMN keywords TO characteristics;

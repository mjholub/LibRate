DELETE FUNCTION update_modified_column();
DELETE TRIGGER person_update_modified ON people.person;
DELETE TRIGGER media_update_modified ON media.media;

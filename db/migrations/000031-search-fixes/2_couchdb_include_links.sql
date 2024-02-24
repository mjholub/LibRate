-- modify the json serializer to include relevant links
CREATE OR REPLACE FUNCTION public.couchdb_put()
 RETURNS trigger
 LANGUAGE plpgsql
AS $function$
DECLARE
    DOC_ID text;
    REV_ID text;
	  AUTH text;
	 	DOC jsonb;
    RES public.http_response;
    COUCHDB_TARGET text;
	  FINAL_REQUEST public.http_request;
BEGIN
	-- it might seem stupid, but even a proper 
	-- alphabetic trigger naming did not work as expected
	-- and doc column contained NULL data
    DOC := public.json_serialize(TG_TABLE_NAME, NEW);
    COUCHDB_TARGET := public.get_couchdb_target(TG_TABLE_NAME);
	-- the actual request starts here
 AUTH := encode('librate:librate', 'base64');
    EXECUTE format('SELECT doc_id FROM %I.%I WHERE id = $1', TG_TABLE_SCHEMA, TG_TABLE_NAME) INTO DOC_ID USING NEW.id;
  IF DOC_ID IS NOT NULL THEN
    -- first, get the rev_id from couchdb to avoid conflicts
    SELECT content::json->>'_rev' FROM http_get('172.20.0.6:5984/' || COUCHDB_TARGET || '/' || DOC_ID) INTO REV_ID;
   RAISE NOTICE 'REV_ID: %', REV_ID;
    FINAL_REQUEST := (
    'PUT',
    'http://172.20.0.6:5984/' || COUCHDB_TARGET || '/' || NEW.doc_id::text, 
    ARRAY[http_header('Accept','application/json'),
 http_header('Authorization', 'Basic' || AUTH),
 http_header('Referer', 'librate-search:5984')],
 'application/json',
      jsonb_set(DOC, '{_rev}', to_jsonb(REV_ID), '{_id}', to_jsonb(NEW.doc_id::text)))::public.http_request;
  ELSE
    -- note that this works only in docker
  	-- For authorization you need to derive your
    FINAL_REQUEST := (
    'POST',
  'http://172.20.0.6:5984/' || TG_TABLE_NAME || '/',
 ARRAY[http_header('Accept','application/json'),
 http_header('Authorization', 'Basic' || AUTH),
 http_header('Referer', 'librate-search:5984')],
  'application/json', DOC::text)::public.http_request;
  END IF;
     
 	RES := public.http(FINAL_REQUEST)::public.http_response;
  RAISE NOTICE 'RES.status: %', RES.status;
  RAISE NOTICE 'RES.content: %', RES.content;

  DOC_ID := RES."content"::json->>'_id';

   --Need to check RES for response code
   IF RES.status != 201 THEN
   	RAISE EXCEPTION 'Error: %', RES;
   END IF;
  
  EXECUTE format('UPDATE %I.%I SET doc_id = DOC_ID WHERE id = $1', TG_TABLE_SCHEMA, TG_TABLE_NAME) USING NEW.id;
  RETURN null;
END;
$function$
;


-- trigger the hooks to push data to CouchDB
UPDATE media.media SET title = title;
UPDATE reviews.ratings SET body = body;
UPDATE public.members SET webfinger = webfinger;
UPDATE media.genres SET name = name;
UPDATE people.person SET first_name = first_name;
UPDATE people.group SET name = name;
UPDATE people.studio SET name = name;

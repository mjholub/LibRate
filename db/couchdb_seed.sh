#!/bin/sh
# NOTE: change the -u flag here if you've already changed the
# default credentials in ../couchdb/config.ini
# test if we're in a container or host machine to use the right address
hostname=""
if [ "$(netstat -tunlap | grep -c 5984)" -gt 0 ]; then
	hostname="http://0.0.0.0:5984"
else
	hostname="172.20.0.6:5984"
fi
for db in "members" "ratings" "media" "studio" "person" \
	"group" "media_images" "genre_descriptions" \
	"genres" "genre_characteristics"; do
	curl -u librate:librate --connect-timeout 3 -f -m 2 -X PUT "$hostname"/"$db" || true
done

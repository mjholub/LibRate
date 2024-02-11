#!/usr/bin/sh

if [ ! -d "/var/lib/postgresql/data" ]; then
	initdb -D /var/lib/postgresql/data &&
		pg_ctl -D /var/lib/postgresql/data -o "-c listen_addresses='*'" -l /var/lib/postgresql/logfile start &&
		createdb -U postgres librate &&
		createdb -U postgres librate_sessions &&
		createdb -U postgres librate_test
fi

psql -U postgres -d librate -f "/tmp/sequential-uuids/sequential_uuids--1.0.1.sql" || true
psql -U postgres -d librate -c "CREATE EXTENSION IF NOT EXISTS sequential_uuids;" || true
psql -U postgres -d librate -c "CREATE EXTENSION IF NOT EXISTS http;" || true

if [ -z $(pgrep postgres) ]; then
	pg_ctl -D /var/lib/postgresql/data -o "-c listen_addresses='*'" -l /var/lib/postgresql/logfile start
fi
tail -f /dev/null

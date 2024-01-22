#!/usr/bin/sh

if [ ! -d "/var/lib/postgresql/data" ]; then
	initdb -D /var/lib/postgresql/data &&
		pg_ctl -D /var/lib/postgresql/data -o "-c listen_addresses='*'" -l /var/lib/postgresql/logfile start &&
		createdb -U postgres librate &&
		createdb -U postgres librate_sessions &&
		createdb -U postgres librate_test &&
		psql -U postgres -d librate -f "/tmp/sequential-uuids/sequential_uuids--1.0.1.sql"
	psql -U postgres -d librate -c "CREATE EXTENSION IF NOT EXISTS sequential_uuids"
fi

# Write the pg_hba.conf and postgresql.conf files
if [ "$(wc -l /var/lib/postgresql/data/pg_hba.conf)" -lt 5 ]; then
	echo "host    all             all             127.0.0.1/32            trust" >/var/lib/postgresql/data/pg_hba.conf
	echo "host    all             all             172.20.0.0/16           trust" >>/var/lib/postgresql/data/pg_hba.conf
	echo "host    all             all             librate-app             trust" >>/var/lib/postgresql/data/pg_hba.conf
	echo "host    all             all             [local]                 trust" >>/var/lib/postgresql/data/pg_hba.conf
	echo "local   all             all                                     trust" >>/var/lib/postgresql/data/pg_hba.conf
	echo "host    all             all             ::1/128                 trust" >>/var/lib/postgresql/data/pg_hba.conf
	echo "host    all             all             0.0.0.0/0               trust" >>/var/lib/postgresql/data/pg_hba.conf
	echo "listen_addresses = '*'" >>/var/lib/postgresql/data/postgresql.conf
	echo "port = 5432" >>/var/lib/postgresql/data/postgresql.conf
fi

if [ -z $(pgrep postgres) ]; then
	pg_ctl -D /var/lib/postgresql/data -o "-c listen_addresses='*'" -l /var/lib/postgresql/logfile start
fi
tail -f /dev/null

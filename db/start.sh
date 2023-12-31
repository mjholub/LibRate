#!/usr/bin/sh

if test ! -d /var/lib/postgresql/data; then
	mkdir -p /var/lib/postgresql/data
	chown -R postgres:postgres /var/lib/postgresql/data
	initdb -D /var/lib/postgresql/data
fi
createdb -U postgres librate &&
	createdb -U portgres test_librate &&
	pg_ctl -D /var/lib/postgresql/data -o "-c listen_addresses='localhost'" -l /var/lib/postgresql/logfile start &&
	psql -U postgres -d librate -c "CREATE EXTENSION IF NOT EXISTS \"sequential-uuids\""

# Write the pg_hba.conf and postgresql.conf files
echo "host    all             all             127.0.0.1/32            trust" >/var/lib/postgresql/data/pg_hba.conf
echo "host    all             all             10.5.0.0/16             trust" >>/var/lib/postgresql/data/pg_hba.conf
echo "host    all             all             [local]                 trust" >>/var/lib/postgresql/data/pg_hba.conf
echo "local   all             all                                     trust" >>/var/lib/postgresql/data/pg_hba.conf
echo "host    all             all             ::1/128                 trust" >>/var/lib/postgresql/data/pg_hba.conf
echo "host    all             all             0.0.0.0/0               trust" >>/var/lib/postgresql/data/pg_hba.conf
echo "listen_addresses = '*'" >/var/lib/postgresql/data/postgresql.conf
echo "port = 5432" >>/var/lib/postgresql/data/postgresql.conf

postgres -D /var/lib/postgresql/data

# keep the script running

tail -f /dev/null

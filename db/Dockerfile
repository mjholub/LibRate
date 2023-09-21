FROM postgres:15-alpine

RUN apk add --no-cache \
  postgresql-dev \
  postgresql-sequential-uuids \
  postgresql15-contrib

ENV POSTGRES_USER postgres
ENV POSTGRES_PASSWORD postgres

USER postgres
RUN initdb -D /var/lib/postgresql/data || true \
  &&  pg_ctl -D /var/lib/postgresql/data start \
  && createdb -U postgres librate \
  && createdb -U postgres librate_test

CMD ["postgres"]

EXPOSE 5432

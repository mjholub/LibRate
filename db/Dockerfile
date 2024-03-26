FROM postgres:15-alpine

WORKDIR /tmp/
COPY ./start.sh /usr/bin/start.sh
COPY ./couchdb_seed.sh /usr/bin/couchdb_seed.sh

RUN apk update && \
  apk add --no-cache \
  git \
  build-base \
  gcc \
  postgresql-dev \
  clang15 \
  clang15-dev \
  llvm15 \
  postgresql-sequential-uuids \
  postgresql15-contrib \
  curl-dev \
  curl

# create the databases in couchDB which we'll use for search
RUN chmod +x /usr/bin/couchdb_seed.sh && \
  /usr/bin/couchdb_seed.sh

# install the 'sequential-uuids' extension
WORKDIR /tmp/sequential-uuids

RUN git clone --depth=1 "https://github.com/tvondra/sequential-uuids/" . && \
  make && \
  make install && \
  if [ "$(mv sequential-uuids/* /usr/share/postgresql/extension/)" ]; then true; fi

WORKDIR /tmp/pg_http

RUN git clone --depth=1 "https://github.com/pramsey/pgsql-http" . && \
  make && \
  make install && \
  if [ "$(mv pg_http/* /usr/share/postgresql/extension/)" ]; then true; fi

WORKDIR /tmp/pg_uint
RUN apk add --no-cache \ 
  python3 && \
  git clone --depth=1 "https://github.com/petere/pguint" . && \
  make && \
  make install && \
  if [ "$(mv pg_uint/* /usr/share/postgresql/extension/)" ]; then true; fi

WORKDIR /tmp

RUN apk del \
  git \
  build-base \
  gcc \
  clang15-dev \
  llvm15 \
  postgresql-dev 

RUN chown postgres:postgres /usr/bin/start.sh && \
  chown -R postgres:postgres /var/lib/postgresql && \
  chmod -R 750 /var/lib/postgresql/data 

USER postgres
VOLUME /var/lib/postgresql/data 
CMD ["sh", "/usr/bin/start.sh"]
EXPOSE 5432


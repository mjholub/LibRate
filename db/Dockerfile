FROM postgres:15-alpine

COPY ./start.sh .

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
  postgresql15-contrib


USER postgres

# Bootstrap the database
RUN initdb -D /var/lib/postgresql/data && \
  pg_ctl -D /var/lib/postgresql/data -l /var/lib/postgresql/data/logfile start && \
  createdb -U postgres librate && \
  createdb -U postgres librate_test 

USER root

# install the 'sequential-uuids' extension
RUN git clone "https://github.com/tvondra/sequential-uuids/" && \
  cd sequential-uuids && \
  make && \
  make install

# Clean up
RUN rm -rf sequential-uuids

RUN apk del \
  git \
  build-base \
  gcc \
  clang15-dev \
  llvm15 \
  postgresql-dev

USER postgres

CMD ["sh", "start.sh"]

EXPOSE 5432


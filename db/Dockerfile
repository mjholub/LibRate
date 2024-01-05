FROM postgres:15-alpine

WORKDIR /tmp/
COPY ./start.sh /usr/bin/start.sh

# we install dev packages for pg16 since AGE requires dependencies that are used by pg16
# while pg15 uses older versions of clang or llvm
RUN --mount=type=cache,target=/var/cache/apk \
  apk update && \
  apk add --no-cache \
  git \
  build-base \
  gcc \
  glib \
  glib-dev \
  make \
  gcompat \
  clang15 \
  clang15-dev \
  llvm15 \
  postgresql-sequential-uuids \
  flex \
  readline-dev readline \
  zlib zlib-dev \
  bison \
  perl \
  libc-dev \
  postgresql16-dev \
  postgresql15-contrib

# install the 'sequential-uuids' extension
WORKDIR /tmp/sequential-uuids

RUN git clone --depth=1 "https://github.com/tvondra/sequential-uuids/" . && \
  make && \
  make install && \
  if [ "$(mv sequential-uuids/* /usr/share/postgresql/extension/)" ]; then true; fi

WORKDIR /tmp/age

RUN --mount=type=cache,target=/var/cache/make \
  wget https://github.com/apache/age/releases/download/PG15%2Fv1.4.0-rc0/apache-age-1.4.0-src.tar.gz \
  -O age.tar.gz && \
  tar x -f age.tar.gz --strip-components=1 && \
  pg_config && \
  make install && \
  rm -rf /tmp/age

WORKDIR /tmp

RUN apk del \
  git \
  build-base \
  gcc \
  clang15-dev \
  llvm15 \
  postgresql-dev \ 
  readline-dev \
  glib-dev \
  zlib-dev \
  postgresql16-dev

RUN chown postgres:postgres /usr/bin/start.sh && \
  chown -R postgres:postgres /var/lib/postgresql && \
  chmod -R 750 /var/lib/postgresql/data 
USER postgres
CMD ["sh", "/usr/bin/start.sh"]
EXPOSE 5432


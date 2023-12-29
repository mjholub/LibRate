FROM postgres:15-alpine

WORKDIR /tmp/
COPY ./start.sh /usr/bin/start.sh

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

# install the 'sequential-uuids' extension
WORKDIR /tmp/sequential-uuids

RUN git clone --depth=1 "https://github.com/tvondra/sequential-uuids/" . && \
  make && \
  make install && \
  if [ "$(mv sequential-uuids/* /usr/share/postgresql/extension/)" ]; then true; fi

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
CMD ["sh", "/usr/bin/start.sh"]
EXPOSE 5432


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


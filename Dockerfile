FROM ubuntu:noble AS app

RUN --mount=type=cache,target=/var/cache/apt \
  apt update && \
  apt upgrade -y && \
  apt -y \
  install --no-install-recommends \
  --no-install-suggests \
  'golang-1.21' \
  'ca-certificates' \
  'npm' \
  'nodejs'=18.13.0+dfsg1-1ubuntu2 \
  'libwebp-dev'=1.3.2-0.3


RUN useradd -U -m -r librate \
  -d /app

USER librate
WORKDIR /app

VOLUME /app
ENV HOME /app
ENV PATH /app/bin:$PATH
ENV GOPATH /app

WORKDIR /app/fe
COPY --chown=librate:librate ./fe /app/fe
RUN npm install && npm run build

USER root
WORKDIR /app/src
COPY . /app/src
RUN --mount=type=cache,target=/app/pkg/mod \
  --mount=type=cache,target=/var/cache/go-build \
  go mod tidy && \
  CGO_ENABLED=0 GOOS=linux go build -ldflags "-w -s" -o /app/bin/librate && \
  go install codeberg.org/mjh/lrctl@latest
WORKDIR /app
COPY --chown=librate:librate ./config.yml /app/data/config.yml
COPY --chown=librate:librate ./static/ /app/data/static
COPY --chown=librate:librate ./db/migrations/ /app/data/migrations
# TODO: change the path being used by tke app so that it doesn't hardcode relative directory
COPY --chown=librate:librate ./views/ /app/bin/views
RUN chown -R librate:librate /app/bin && \
  chmod -R 755 /app/bin/

USER librate

ENV USE_SOPS=false

EXPOSE 3000
CMD [ "/app/bin/librate", "-c", "/app/data/config.yml" ]
# [ "/usr/bin/bash" ]

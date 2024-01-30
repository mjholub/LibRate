# Build frontend
FROM ubuntu:noble AS frontend-builder 

WORKDIR /app/fe
COPY ./fe /app/fe

RUN --mount=type=cache,target=/var/cache/apt \
  apt update && \
  apt upgrade -y && \
  apt -y \
  install --no-install-recommends \
  --no-install-suggests \
  'npm'=9.2.0~ds1-2 \
  'nodejs'=18.13.0+dfsg1-1ubuntu2 && \
  npm install && npm run build

# Build backend
# TODO: try moving backend to src/
FROM golang:1.21-alpine3.19 AS backend-builder

VOLUME /app
ENV HOME /app
ENV PATH /app/bin:$PATH
ENV GOPATH /app

WORKDIR /app/src
COPY . /app/src
RUN --mount=type=cache,target=/app/pkg/mod \
  --mount=type=cache,target=/var/cache/go-build \
  go mod tidy && \
  CGO_ENABLED=0 GOOS=linux go build -ldflags "-w -s" -o /app/bin/librate && \
  go install codeberg.org/mjh/lrctl@latest

# Build final image
FROM alpine:3.19 AS app
RUN apk update && apk add --no-cache 'libwebp'=1.3.2-r0 'libwebp-dev'=1.3.2-r0 'ca-certificates' \
  && apk cache purge \
  && addgroup -S librate \ 
  && adduser -G librate -S -D librate \
  -h /app

USER librate

WORKDIR /app
COPY --from=frontend-builder --chown=librate:librate /app/fe/build /app/data/fe/build
COPY --from=backend-builder --chown=librate:librate /app/bin /app/bin
COPY --chown=librate:librate ./config.yml /app/data/config.yml
COPY --chown=librate:librate ./static/ /app/data/static
COPY --chown=librate:librate ./db/migrations/ /app/data/migrations
# TODO: change the path being used by tke app so that it doesn't hardcode relative directory
COPY --chown=librate:librate ./views/ /app/bin/views
RUN chmod -R 755 /app/bin/

USER librate

ENV USE_SOPS=false

EXPOSE 3000
CMD [ "/app/bin/librate", "-c", "/app/data/config.yml" ]
# [ "/usr/bin/bash" ]

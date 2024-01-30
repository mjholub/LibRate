# Build frontend
FROM node:lts-alpine AS frontend-builder

WORKDIR /app/fe
COPY ./fe /app/fe

RUN --mount=type=cache,target=/app/.cache \
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
RUN apk update && apk add 'ca-certificates' \
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
COPY --chown=librate:librate ./views/ /app/bin/views
COPY --chown=librate:librate ./data /app/query-builder
# large query for genre information
RUN go run /app/query-builder/main.go && \
  mv /app/data/migrations/000023-media-form-pt/6_sixth_migration.up.sql /app/data/migrations/000023-media-form-pt/7_seventh_migration.up.sql && \
  mv /app/data/query-builder/queries.sql /app/data/migrations/000023-media-form-pt/6_sixth_migration.up.sql && \
  rm -rf /app/query-builder

RUN chmod -R 755 /app/bin/

USER librate

ENV USE_SOPS=false

EXPOSE 3000
CMD [ "/app/bin/librate", "-c", "/app/data/config.yml" ]
# [ "/usr/bin/bash" ]

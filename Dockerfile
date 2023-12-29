# Build frontend
FROM node:lts-alpine AS fe-builder
RUN npm install -g "pnpm@latest"
WORKDIR /app/fe
COPY ./fe /app/fe
RUN pnpm install && pnpm run build

FROM golang:1.21-alpine AS app

RUN addgroup -S librate && adduser -S librate -G librate

WORKDIR /app
VOLUME /app
ENV HOME /app
ENV PATH /app/bin:$PATH

RUN if [ "$(addgroup -S librate $?)" != 0 ]; \
  then echo ""; \
  fi 
RUN if [ "$(adduser -S librate -G librate $?)" != 0 ]; then echo ""; fi
COPY --from=fe-builder /app/fe/build /app/bin/fe/build

WORKDIR /app
#RUN go mod tidy && \
#  CGO_ENABLED=0 GOOS=linux go build -ldflags "-w" -o /app/bin/librate
# skip compilation since it can take some time, use pre-built binaries (see
# Releases on Codeberg) instead.
# Add a directive to copy everything from cwd to /app and uncomment the line
# above if you want to compile the app yourself anyway
ENV GO_BIN=/app/bin
COPY .env /app/.env
COPY .air.toml /app/data/.air.toml
RUN go install codeberg.org/mjh/lrctl@latest
COPY ./src /app/src
WORKDIR /app/src
RUN go build -ldflags "-w -s" -o /app/bin/librate
WORKDIR /app/data
COPY ./config.yml /app/data/config.yml
COPY ./static/ /app/data/static
COPY ./src/db/migrations/ /app/data/migrations

WORKDIR /app/bin
# add live reload
# when the container is started, the source code directory must be mounted to /app
RUN apk add --no-cache inotify-tools && \
  apk add --no-cache -X "https://dl-cdn.alpinelinux.org/alpine/edge/testing" air

COPY ./fe/entrypoint.sh /app/bin/fe-entrypoint.sh
RUN chmod +x /app/bin/fe-entrypoint.sh

COPY ./entrypoint.sh /app/bin/entrypoint.sh
RUN chmod +x /app/bin/entrypoint.sh

RUN chown -R librate:librate /app

USER librate 
ENV USE_SOPS=false
#RUN go mod tidy && \
#  go build -ldflags "-w" -o /app/bin/librate && \ 
#  chmod +x /app/bin/librate

EXPOSE 3000
#CMD [ "/app/bin/librate", "-c", "env", "-e" ]
CMD [ "/app/bin/entrypoint.sh" ]

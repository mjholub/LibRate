FROM golang:1.20-alpine AS app

RUN addgroup -S librate && adduser -S librate -G librate

WORKDIR /app
VOLUME /app
ENV HOME /app
ENV PATH /app/bin:$PATH

COPY . .
COPY config_compose.yml  /app/config.yml

RUN apk add --no-cache \
  nodejs-lts \
  npm \
  just

RUN chown -R librate:librate /app
USER librate 
RUN just copy_libs tidy build_frontend && \
  go build -o /app/bin/librate && \ 
  chmod +x /app/bin/librate

# initialize the database, don't launch the database subprocess and rely solely on pg_isready, run the migrations

EXPOSE 3000
CMD ["/app/bin/librate", "migrate", "-auto", "-no-db-subprocess", "-hc-extern", "-init", "--config", "/app/config.yml"]

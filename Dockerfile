FROM golang:1.20-alpine AS app

RUN addgroup -S librate && adduser -S librate -G librate

WORKDIR /app
VOLUME /app
ENV HOME /app
ENV PATH /app/bin:$PATH

COPY ./fe /app/fe
RUN cd fe && npm install && npm run build

RUN addgroup -S librate && adduser -S librate -G librate

COPY . /app
COPY config_compose.yml /app/data/config.yml

RUN just copy_libs tidy && \
  CGO_ENABLED=0 GOOS=linux go build -o /app/bin/librate

RUN chown -R librate:librate /app
USER librate 
RUN just copy_libs tidy build_frontend && \
  go build -o /app/bin/librate && \ 
  chmod +x /app/bin/librate

# initialize the database, don't launch the database subprocess and rely solely on pg_isready, run the migrations

EXPOSE 3000
CMD ["/app/bin/librate", "-no-db-subprocess", "-hc-extern", "-init", "migrate", "-config", "/app/data/config.yml"]

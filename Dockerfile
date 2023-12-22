FROM golang:1.21-alpine AS app

RUN addgroup -S librate && adduser -S librate -G librate

WORKDIR /app
VOLUME /app
ENV HOME /app
ENV PATH /app/bin:$PATH

COPY ./fe /app/fe
RUN cd fe && npm install && npm run build

RUN addgroup -S librate && adduser -S librate -G librate

COPY . /app
COPY config.yml /app/data/config.yml

RUN just copy_libs tidy && \
  CGO_ENABLED=0 GOOS=linux go build -o /app/bin/librate

RUN chown -R librate:librate /app

USER librate 
ENV USE_SOPS=false
RUN just copy_libs tidy build_frontend && \
  go build -ldflags "-w" -o /app/bin/librate && \ 
  chmod +x /app/bin/librate

# -hc-extern tells the app not to check database connection on startup

EXPOSE 3000
CMD ["/app/bin/librate", "-hc-extern", "-c", "/app/data/config.yml"]

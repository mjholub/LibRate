FROM opensuse/leap:15 AS app

RUN --mount=type=cache,target=/var/cache/zypp \
  zypper --non-interactive \
  install --no-recommends \
  go \
  unzip 

RUN useradd -U -m -r librate \
  -d /app

USER librate
WORKDIR /app
RUN --mount=type=cache,target=/app/.cache \
  curl -fsSL https://bun.sh/install | bash
RUN source /app/.bashrc

VOLUME /app
ENV HOME /app
ENV PATH /app/bin:$PATH
ENV GOPATH /app

WORKDIR /app/fe
COPY --chown=librate:librate ./fe /app/fe
RUN /app/.bun/bin/bun install && /app/.bun/bin/bun run build

USER root
WORKDIR /app/src
COPY . /app/src
RUN --mount=type=cache,target=/app/pkg/mod \
  --mount=type=cache,target=/var/cache/go-build \
  go mod tidy && \
  CGO_ENABLED=0 GOOS=linux go build -ldflags "-w -s" -o /app/bin/librate && \
  go install codeberg.org/mjh/lrctl@latest
WORKDIR /app
COPY --chown=librate:librate .env /app/.env
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
CMD [ "/app/bin/librate", "-c", "env", "-e" ]
# [ "/usr/bin/bash" ]

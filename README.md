# LibRate
[![weblate badge](https://translate.codeberg.org/widget/librate/frontend-messages/svg-badge.svg?native=1
)](https://translate.codeberg.org/projects/librate/)
[![GoDoc](https://godoc.org/codeberg.org/mjh/LibRate?status.svg)](https://pkg.go.dev/codeberg.org/mjh/LibRate)
![Depfu](https://img.shields.io/depfu/dependencies/github/mjholub%2FLibRate)
![LiberaPay](https://img.shields.io/liberapay/receives/Librate.svg?logo=liberapay)

### Project status - May 2024

I am doing a substantial rewrite and focusing on paying off the technical debt as much as possible so don't expect many new features within the next couple months. But I really have other priorities at the moment, perhaps I will be able to tick of at least 1/5 of the milestones later in the summer.
<hr />

LibRate is a project that aims to provide a free, federated service for tracking, reviewing, sharing and discovering films, books, music, games and other culture texts.


Element, probably the only Matrix client to support publishing roomss has _temporarily_ disabled this feature, therefore we've removed the widget for it, so if you want to join us, go to
[#librate-dev:matrix.org](https://matrix.to/#/#librate-dev:matrix.org)

You can donate to it's development via [LiberaPay](https://liberapay.com/LibRate) or Monero (contact me via Matrix for the latter).

**The only public beta instance can be found at [librate.club](https://librate.club/). Federation is currently work in progress.**

# Table of contents

<!-- toc -->

- [Roadmap:](#roadmap)
  - [**Social features**](#social-features)
  - [**Media features**](#media-features)
  - [**Reviews**](#reviews)
  - [**Recommendations**](#recommendations)
  - [**Other**](#other)
- [Deploying with Docker](#deploying-with-docker)
- [Prerequisites for running natively:](#prerequisites-for-running-natively)
  - [Get the Dependencies](#get-the-dependencies)
  - [Setup secrets](#setup-secrets)

* [IMPORTANT: Updating the app and instance administration](#important-updating-the-app-and-instance-administration)
  - [Development prerequisites](#development-prerequisites)
  - [Building and installing](#building-and-installing)
* [Testing](#testing)

<!-- tocstop -->

## Roadmap:

### **Social features**

- [x] Basic registration support
- [x] User profile cards
- [x] Full profile pages
- [ ] Tagging and mentions
- [x] Advanced profile and UI customization
- [ ] Groups
- [ ] Direct messages (E2EE)
- [ ] Group chats, more group-friendly design, like Lemmy or Kbin
- [x] (**WIP**) User-generated content tagging and categorization
- [x] Following
- [ ] Sharing
- [x] (**WIP**) **ActivityPub support**, with selective federation

### **Media features**

- [x] Album cards
- [x] Carousels showing random media
- [x] Relevant DB setup
- [x] Film and series cards
  - [ ] Trailers and stills support
- [ ] Release notifications
  - [ ] Sending them as DMs to federated service accounts
- [ ] Content filters
- [ ] Book cards and pages
  - [ ] Bookwyrm federation
- [ ] Anime and manga cards/pages
- [ ] Games support
- [ ] Customizable, shareable media collections and logging
- [x] (**WIP**) Convenient submission form, with decentralized deduplication and POW-based anti-spam (a bit similar to Bookwyrm)
- [x] Automated imports from 3rd party sources
- [ ] Scrobbling
- [ ] Decentralized subtitle repository
- [ ] DRM-free audio hosting and streaming(?)
- [ ] Visual artwork galleries and ability to claim own page as an artist
- [ ] Arbitrarily defined custom media objects (like board games, anime figures or other collectible)

### **Reviews**

- [x] Basic review form
- [x] Backend logic for submission and fetching of reviews
- [x] (**WIP**) Review feed
- [ ] Commenting and voting on reviews and media items
- [ ] Importing from 3rd party sources

### **Recommendations**

- [x] Prototype logic
- [ ] Basic working implementation
  - [ ] Personalized feeds
  - [ ] Advanced algorithm

### **Other**

- [x] Extended configurability
- [x] Internationalization
- [ ] Events (like concerts or conventions)
- [ ] Federated merch and works marketplace, possibly an alternative to Bandcamp
- [ ] Mobile app

## General administration information

Adjust the configuration file *example_config.yml* to suit your needs. 

The following configuration paths will be automatically checked and not require passing a `-c` flag. These paths can either contain no or a .yml or .yaml extension.

- "./config",
-     "./config/config",
-     "/etc/librate/config",
-     "/var/lib/librate/config",
-     "/opt/librate/config",
-     "/usr/local/librate/config",
-     ~/.local/share/librate/config",
-     ~/.config/librate/config"

Then make the necessary adjustments to the privacy policy and Terms of Service templates, located in _static/templates/examples_ and move them one directory up. If your instance is country-specific, you can add another file with your prospective users' main language code as a suffix to that directory. You need to at keep at least one version of TOS and Privacy using the [BCP 47 language tag](https://en.wikipedia.org/wiki/IETF_language_tag) you've set as `defaultLanguage` in config.

## Deploying with Docker

Set up your configuration as described earlier. If something's wrong with the networking, try setting the hostnames in config for anything except the gRPC server to `0.0.0.0`.

Then, before deploying, update the following line in [postgres container start script](./db/start.sh), or use a secrets management tool of your choice:

```sh
psql -U postgres -c "ALTER USER postgres WITH PASSWORD 'CHANGE_ME';"
```

You can use compose. Just remember to execute the commands from step 5 and create the network first. 

Optionally in compose you can enable monitoring of your instance with Grafana and Prometheus. To do so, make sure to modify the password and remove the entrypoint lines for these services in the compose file.

1. Create a Docker network with

```sh
# 172.20.0.0 is the default subnet, this is needed for pg_hba.conf, though
# generally a hostname should work
docker network create --subnet 172.20.0.0/16 librate-net
```

2. Run the redis container with

```sh
 docker run -d --rm --network=librate-net --hostname "librate-redis"
 redis:alpine
```

3. Build and run the database container

```sh
cd db && \
docker build -t librate-db . && \
docker run -it --network=librate-net --hostname "librate-db" librate-db:latest
```

4. Create an config file from the [provided example](https://codeberg.org/mjh/LibRate/src/branch/main/example_config.yml). Alternatively you can configure the container to use the .env config, but .yaml is somewhat more reliable (for example, as of 30-01-2024 there is no way of setting up thumbnailing with .env).

5. Build and run the app container. In the container run
```sh
lrctl -c [path to config file or 'env'] db init
lrctl -c [...] db migrate # If the migrations fail you can copy them to the database container and apply them manually, although this shouldn't happen. Note that each migration has a corresponding rollback (down) migration.
# Something like (fish syntax) for f in (fd -t f *.up.sql .);psql -u postgres -d librate -a -f $f;end should do as a fallback, but please note thad you lose the benefit of proper rollback on any errors along the way.
```

6. Finally, set up some reverse proxy, like caddy or nginx, to resolve the app to a
   hostname (locally you can also use something like `lr.localhost`). Note that
   YOU MUST set up some downstream server with HTTPS or get your own TLS certificates and set tls to true and specify key paths in config. since the app hardcodes certain security
   headers that make it unusable without HTTPS (although self-signed certs work,
   but you're better off using caddy which will automatically generate a free
   certificate with Let's Encrypt for you).

**TIP:**
Caddy lets you really easily have zero to minimum downtime when updating the app. All you need to do is set up `lb_policy` to `first`, then when updating the app, change the config to use the second port you've set Caddy to balance the load for your domain.

Note the required healthchecks setup. Our API endpoint for that is `/api/health/check`. See [official documentation](https://caddyserver.com/docs/caddyfile/directives/reverse_proxy#load-balancing) for more details.

It also lets you have other nice things, like automatically rejecting bad actors using a crowdsec plugin.

## Prerequisites for running natively:

### Get the Dependencies

- [SOPS](https://github.com/getsops/sops) and [age](https://github.com/FiloSottile/age) for handling secrets
- A JS bundler, like pnpm or bun, for building the frontend
- working **Postgres**, **Redis** and **Melisearch** (it also supports an embedded *Bleve* search engine, but it's less reliable than Meili. I'll probably also include support for Postgresql's RUM in the future) instances. You'll also need to install the development files package for postgres since LibRate uses Postgres extensions. You may also need to manually build the [sequential UUIDs](https://github.com/tvondra/sequential-uuids/) extension

### Setup secrets


**A foreword**: 
If you experience any issues, just set `USE_SOPS` to false in your environment.

You may ask about other storage options for secrets.
Well, relying on local storage and age is the simplest way for now, but luckily
thanks to sops' versatility, we'll successively work on adding support for more ways of handling them.

Please don't hesitate to open an issue if you feel the need for support for a particular secrets storage option supported by SOPS (see their README linked in dependencies list) to be added first.

In a production environment, you're strongly advised to create a separate user for LibRate
and use that to handle the config, run the binary and store the
[age](https://github.com/FiloSottile/age) keys.

All you need to do is [generate an age X25519 identity](https://github.com/FiloSottile/age)
and then [encrypt the config file](https://github.com/getsops/sops#22encrypting-using-age)
with SOPS. Don't forget to pass the `-c` flag if you decide to save the encrypted file under a new path.

# IMPORTANT: Updating the app and instance administration

Release notes for each update should mention when the database requires
migrations for the new release to work properly, including the migration's name.

LibRate provides an utility for administration an instance, called [lrctl](https://codeberg.org/mjh/lrctl). If you use docker, this utility is bundled into the image. To run a migration, execute:

```sh
/app/bin/lrctl -c /app/data/config.yml db migrate <migration directory's base
path>
```

In the future it will ship as a module of the base app once I restructure the repo since the current approach poses some risk of dependency cycles.

## Development prerequisites

To develop the (currently not in active development) recommendations feature, you'll need:

- `protoc` and `protoc-gen-go` for generating code from the protocol buffers files.
- Rust and Go toolchains

## Building and installing

If you have installed [just](https://github.com/casey/just), you can simply run:

```sh
just first_run
```

Alternatively, edit the example config file and run:

```sh
go mod tidy  && \
cd fe && pnpm install \
&& pnpm run build && \
go build -ldflags "-w -s" -o librate
./librate
# then the lrctl bootstraping commands described earlier WHILE librate is running
```

You can then test your instance at [http://127.0.0.1:3000](127.0.0.1:3000) (or the other port you've specified)

# Testing

In order to test the database code, you should create a `librate_test` database. We currently try to move towards using mocks where possible, but there are some incompatibilities between the mocking library we use and the actual implementation.

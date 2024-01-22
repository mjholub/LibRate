# LibRate

This project aims to bring a website combining the functionality of such projects as Bookwyrm, RateYourMusic/Sonemic, IMDB and similar to the #fediverse.

This project is currently in early beta stage, bugs are expected and PRs are very welcome.

**The first public beta instance is expected to launch in Q1 2024. Bleeding edge branch can be found [here](https://codeberg.org/mjh/LibRate/src/branch/dev)**

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
- [x] (WIP) Advanced profile and UI customization
- [ ] Groups
- [ ] Direct messages (E2EE)
- [ ] Group chats, more group-friendly design, like Lemmy or Kbin
- [x] (**WIP**) User-generated content tagging and categorization
- [x] (**WIP**) Following
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
- [ ] DRM-free audio hosting and streaming(?)
- [ ] Artwork galleries for visual artists(?)

### **Reviews**

- [x] Basic review form
- [x] Backend logic for submission and fetching of reviews
- [x] (**WIP**) Review feed
- [ ] Commenting and voting on reviews and media items
- [ ] Importing from 3rd party sources

### **Recommendations**

- [x] Prototype logic
- [ ] Actual working implementation
  - [ ] Personalized feeds
  - [ ] Advanced algorithm powered by ML and **graph-like database structure - already implemented**

### **Other**

- [x] Extended configurability
- [ ] Internationalization
- [ ] Events, federating with Mobilizon
- [ ] Federated merch and works marketplace, possibly an alternative to Bandcamp
- [ ] Mobile app (although the frontend is and will be mobile friendly, but also never at the expense of desktop experience. We'll also try to make it work with Fedilab, though the number of distinctive features may make it difficult)

## Deploying with Docker

You can use compose or 

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

4. Create an .env file from the [provided example](https://codeberg.org/mjh/LibRate/src/branch/main/.env.example). Alternatively you can configure the container to use the .yml config, but .env is somewhat more reliable.

5. Build and run the app container. In the container run
```sh
lrctl -c [path to config file or 'env'] db init
lrctl -c [...] db migrate # If the migrations fail you can copy them to the database container and apply them manually, although this shouldn't happen. Note that each migration has a corresponding rollback (down) migration.
```

6. Finally, set up some reverse proxy, like caddy or nginx, to resolve the app to a
   hostname (locally you can also use something like `lr.localhost`). Note that
   YOU MUST set up some reverse proxy, since the app hardcodes certain security
   headers that make it unusable without HTTPS (although self-signed certs work,
   but you're better off using caddy which will automatically generate a free
   certificate with Let's Encrypt for you).

## Prerequisites for running natively:

### Get the Dependencies

- [SOPS](https://github.com/getsops/sops) and [age](https://github.com/FiloSottile/age) for handling secrets
- `pnpm`, `yarn` or `npm`, for building the frontend
- working **Postgres** and **Redis** instances. You'll also need to install the development files package for postgres since LibRate uses Postgres extensions. You may also need to manually build the [sequential UUIDs](https://github.com/tvondra/sequential-uuids/) extension

### Setup secrets

**A foreword**: you may ask about other storage options for secrets.
Well, relying on local storage and age is the simplest way for now, but luckily
thanks to sops' versatility, we'll successively work on adding support for more ways of handling them.

Please don't hesitate to open an issue if you feel the need for support for a particular secrets storage option supported by SOPS (see their README linked in dependencies list) to be added first.

In a production environment, you're strongly advised to create a separate user for LibRate
and use that to handle the config, run the binary and store the
[age](https://github.com/FiloSottile/age) keys.

All you need to do is [generate an age X25519 identity](https://github.com/FiloSottile/age)
and then [encrypt the config file](https://github.com/getsops/sops#22encrypting-using-age)
with SOPS. Don't forget to pass the `-c` flag if you decide to save the encrypted file under a new path.

The following config paths will be automatically checked and not require passing a `-c` flag. These paths can either contain no or a .yml or .yaml extension.

- "./config",
-     "./config/config",
-     "/etc/librate/config",
-     "/var/lib/librate/config",
-     "/opt/librate/config",
-     "/usr/local/librate/config",
-     ~/.local/share/librate/config",
-     ~/.config/librate/config"

# IMPORTANT: Updating the app and instance administration

Release notes for each update should mention when the database requires
migrations for the new release to work properly, including the migration's name.

LibRate provides an utility for administration an instance, called [lrctl](https://codeberg.org/mjh/lrctl). If you use docker, this utility is bundled into the image. To run a migration, execute:

```sh
/app/bin/lrctl -c /app/data/config.yml db migrate <migration directory's base
path>
```

## Development prerequisites

To develop the recommendations feature, you'll need:

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
go run . -init -exit && \
go run . migrate -auto-migrate
```

For subsequent runs of course you shouldn't use the `init` flag.

You can then test your instance at [http://127.0.0.1:3000](127.0.0.1:3000) (or the other port you've specified)

# Testing

In order to test the database code, you should create a `librate_test` database.

If you set the `$CLEANUP_TEST_DB` variable to 0, the test database will not be cleaned up by the deferred function in the database initialization unit test.

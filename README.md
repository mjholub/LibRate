# LibRate

This project aims to bring a website combining the functionality of such projects as Bookwyrm, RateYourMusic/Sonemic, IMDB and similar to the #fediverse. 

This project is currently in early alpha stage, bugs are expected and PRs are very welcome.

**The first public beta instance is expected to launch by the end of September 2023.**

**NOTE**: we're using the term *member* to stress inclusivity and openness, as opposed to the term *user* which is controversial, because it was borrowed from the term drug dealers use to refer to their customers.

## Roadmap:

### **Social features**:
  - [x] Basic registration support
  - [x] Member cards with profile info
  - [x] (**WIP**)Full member profile pages
  - [ ] Tagging and mentions
  - [ ] Advanced profile and UI customization
  - [ ] Groups
  - [ ] Direct messages (E2EE)
  - [ ] Group chats, more group-friendly design, like Lemmy or Kbin
  - [x] (**WIP**) Member-generated content tagging and categorization
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
  - [ ] Automated imports from 3rd party sources
  - [ ] DRM-free audio hosting and streaming, federation with Funkwhale
  - [ ] Artwork galleries for visual artists(?)
  ### **Reviews**
  - [x] Basic review form
  - [?] Backend logic for submission and fetching of reviews
  - [x] (**WIP**) Review feed
  - [ ] Commenting and voting on reviews and media items
  - [ ] Importing from 3rd party sources

### **Recommendations**
  - [x] Prototype logic
  - [ ] Actual working implementation
    - [ ] Personalized feeds
    - [ ] Advanced algorithm powered by ML and **graph-like database structure - already implemented**


### **Other**
  - [ ] Extended configurability
  - [ ] Signed builds and security mechanisms preventing federation with modified versions of LibRate
  - [ ] Admin panel
  - [ ] Events, federating with Mobilizon
  - [ ] Federated merch and works marketplace, possibly an alternative to Bandcamp
  - [ ] Mobile app (although the frontend is and will be mobile friendly, but also never at the expense of desktop experience. We'll also try to make it work with Fedilab, though the number of distinctive features may make it difficult)

## Prerequisites:

- `pnpm`, `yarn` or `npm`, for building the frontend
- Python 3 for setting up the uint Postgres extension
- working **Postgres** and **Redis** instances. You'll also need to install the development files package
  since LibRate uses Postgres extensions

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
go run . -init 
```

For subsequent runs of course you shouldn't use the `init` flag.

Additionally, for now you'll also have to run each of the migrations in the _db/migrations_ folder.

You can then test your instance at [http://127.0.0.1:3000](127.0.0.1:3000)

# Testing

In order to test the database code, you should create a `librate_test` database.

If you set the `$CLEANUP_TEST_DB` variable to 0, the test database will not be cleaned up by the deferred function in the database initialization unit test.

## Legal notice

All images included in this repository are assumed to be fair use.

If you are the copyright holder of an image which you want to be removed, 
please [contact the maintaner](mailto:1a6f1a@riseup.net).

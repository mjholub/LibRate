# LibRate

This project aims to bring a website combining the functionality of such projects as Bookwyrm, RateYourMusic/Sonemic, IMDB and similar to the #fediverse. 

This project is currently in early alpha stage, bugs are expected and PRs are very welcome. 

## Roadmap:

### **Social features**:
  - [x] Basic registration support
  - [x] (**WIP**) Member cards with profile info
  - [ ] Full member profile pages
  - [ ] Groups
  - [ ] Direct messages, follows, sharing...
  - **ActivityPub support**
### **Media features**
  - [x] Album cards
  - [x] Carousels showing random media
  - [x] Relevant DB setup
  - [ ] Film and series cards
  - [ ] Book cards
  - [ ] Anime and manga cards/pages
  ### **Reviews**
  - [x] Basic review form
  - [?] Backend logic for submission and fetching of reviews
  - [x] (**WIP**) Review feed

### **Recommendations**
  - [x] Prototype logic
  - [ ] Actual working implementation

###**Other**
  - [ ] Extended configurability
  - [ ] Admin panel

## Prerequisites:

- `pnpm`, `yarn` or `npm`, for building the frontend
- Python 3 for setting up the uint Postgres extension
- a working Postgres instance. You'll also need to install the development files package
  since LibRate uses Postgres extensions

## Development prerequisites

To develop the recommendations feature, you'll need:

- `protoc` and `protoc-gen-go` for generating the protobufs
- Rust and Go toolchains

## Building and installing

```
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

# LibRate

This project aims to bring a website combining the functionality of such projects as Bookwyrm, RateYourMusic/Sonemic, IMDB and similar to the #fediverse. 

This project is currently in early alpha stage, bugs are expected and PRs are very welcome. 

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

You can then test your instance at [http://localhost:3000](localhost:3000)

# Testing

In order to test the database code, you should create a `librate_test` database.

If you set the `$CLEANUP_TEST_DB` variable to 0, the test database will not be cleaned up by the deferred function in the database initialization unit test.

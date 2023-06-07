# LibRate

This project aims to bring a website combining the functionality of such projects as Bookwyrm, RateYourMusic/Sonemic, IMDB and similar to the #fediverse. 

This project is currently in early alpha stage, bugs are expected and PRs are very welcome. 

## Prerequisites:

- `pnpm`, `yarn` or `npm`, for building the frontend
- a working Postgres instance

## Development prerequisites

To develop the recommendations feature, you'll need:

- `protoc` and `protoc-gen-go` for generating the protobufs
- Rust toolchain

## Building and installing

```
go mod tidy  && \
cd fe && pnpm install \
&& pnpm run build && \
go run . -init 
```

For subsequent runs of course you shouldn't use the `init` flag.

You can then test your instance at [http://localhost:3000](localhost:3000)

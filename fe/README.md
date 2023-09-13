# LibRate frontend

This directory contains, as the name suggests, the frontend source code for LibRate. It uses the static adapter 
(more on it's quirks and limitations and a few hacks to overcome them in [this great blog post](https://khromov.se/the-missing-guide-to-understanding-adapter-static-in-sveltekit/)).

## Why the static adapter

TL;DR: simplicity – we've tried using Bun and it's kind of a PITA to rewrite API calls and get CORS working with other adapters such as Bun. 
Besides, the projects strives to follow the KISS ( _Keep it simple, stupid_ ) philosophy – that's also why we chose Svelte in the first place,
since it doesn't use Virtual DOM, unlike React or Vue and has greater performance than these frameworks while being much simpler to develop than 
WebAssembly. The lack of an additional JS server in addition to the Go backend reduces the resource usage and improves security by reducing the attack surface.

While we get that some things might be easier to achieve when using a bundler, we believe that a static adapter will suffice despite some challenges it may pose.

## Alternatives

In addition to this frontend, work is underway to create a noscript alternative based on the ported Django templating engine.

# Building

Just run `pnpm run build`


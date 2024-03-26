# Tags, branching

Create your PRs against the upstream dev branch, where they will be tested and if everything is fine, merged with main.

Once you branch from dev and you're ready to submit a PR, ensure you have all tags using `git fetch -a --tags`. Then, tag your version using semantic versioning that summarizes the nature of your PR, or just use the name of your feature branch.

# Commits

For your commits on the feature branch, follow **atomic and conventional commits**. These will later undergo a squash merge to clean up the reflog, but for your feature branch you should use the them to make the review process easier and also to make it easier to undo things if anything's wrong.

# Testing

For small changes you might just do some manual testing, for larger changes provide unit tests and for really large ones, integration tests.
As a maintainer I can help you if anything feels unclear or overwhelming during the process of testing any larger components.

**TIP**: First, define or extend existing interfaces, then use a tool like _mockgen_ to generate mocks.
Although if you want to be 100% sure when contributing code that affects the database, you can create a test db with the relevant tables, 
or (although the data types might not always exactly match), an in-memory sqlite database.

# Style, concurrent code

Your editor should have support for _eslint_, svelte LSP and _golangci-lint_ installed and enabled. For _golangci-lint_ though, you can also use it as a CLI tool, although a language server version is more convenient.

As you probably see, LibRate's codebase relies heavily on the dependency injection pattern for anything public. 

This also implies other related pattern which are rather idiomatic in Go, such as constructors and method receivers.

Your code should be [referentially transparent](https://en.wikipedia.org/wiki/Referential_transparency). Channels are not, so if you want to do concurrency, only use them if it's 
really necessary, like long-running background tasks that actually require inter-procedure communication. See [this blog post](https://www.jtolio.com/2016/03/go-channels-are-bad-and-you-should-feel-bad/). 

Also when it comes to concurrency, before deciding to write anything concurrent, please write two copies of your function, one sequential and the other not, then benchmark them.

You'd be surprised how often the goroutine overhead exceeds it's benefits you'd expect based on your initial assumptions.

Also, please do not use busy wait. It's highly unidiomatic in Go.

  It can be self-documenting but if it's more complex, add some short comment.

In the _controllers_ pacakge you'll find examples of how to write Swagger-compatible comments for APIs and in _models_, in structs tagged with the `example` tag, how to document your models. For more details please see the wiki.

# Dependencies

When introducing a new dependency, consider the following:

- Is the dependency actively maintained and stable?
- Would it be a significant improvement and/or time saver compared to doing the same thing from scratch
- Popularity: sometimes there are good libraries out there which few projects use, but generally popular libraries are a bit more trustworthy and less likely to be vulnerable to supply chain attacks
- Avoiding bloat: use tools like Github GLOC, tokei, scc or library size info on npm website to assess and compare codebase sizes of various libraries

## Go backend

The topmost functions, like those in the _main_ and _routes_ packages, inject the dependencies (like logger and database connection) into the functions in _models_ and/or _controllers_ packages.  

Resources intensive operations called inside other functions, should be parallelized with goroutines, but make sure they are properly synchronized if needed. For greater predictabilty you can use RxGo, which is a ReactiveX implementation for Go, allowing you to use stuff such as the observer pattern.
 Note however, that it is preferable to use smaller, dedicated functions with limited or no side effects instead of unnecessarily large ones.

Long, more complex functions should include a brief comment summarizing their mechanism of action.

For error messages, prefer to use the status codes provided by fiber, to avoid unnecessary overlap with _net/http_.

Also, a general advice: if you're not sure how to implement something in the frontend, 
just submit a PR with the backend functionality, as long as it returns the expected data 
in the unit tests or to `curl`.

## Frontend

An easy and welcome way to contribute is writing CSS stylesheets to be used as themes. Please put them in ./static/themes folder. SCSS or other stuff that hasn't gone through a preprocessor won't work, only CSS files can be used by the frontend as custom themes source.

Avoid doing too much computation on the client's side. If something can be done by an API endpoint in the Go backend, then do it this way.

Keep all type definitions in _fe/src/types_ and the more complex functions for managing state in _fe/src/stores_

The goal is an accessible, customizable site which can be feature rich but must not be a potential source of information overload and which works fast thanks to heavy use of SSR.

# Tags, branching

Create your PRs against the upstream dev branch, where they will be tested and if everything is fine, merged with main.

Once you branch from dev and you're ready to submit a PR, ensure you have all tags using `git fetch -a --tags`. Then, tag your version using semantic versioning that summarizes the nature of your PR, or just use the name of your feature branch.

# Commits

For your commits on the feature branch, follow **atomic and conventional commits**. These will later undergo a squash merge to clean up the reflog, but for your feature branch you should use the them to make the review process easier and also to make it easier to undo things if anything's wrong.

# Testing

Although striving towards 100% test coverage is irrational, to say the least, provide unit tests for your code. Test driven development is advisable especially for new features. Test the frontend with jest, and for the backend just run `go test -v. /...`

  **TIP**: First, define or extend existing interfaces, then use a tool like _mockgen_ to generate mocks. Although if you want to be 100% sure when contributing code that affects the database, you can create a test db with the relevant tables, or (although the data types might not always exactly match), an in-memory sqlite database.

# Style

Your editor should have support for _eslint_, svelte LSP and _golangci-lint_ installed and enabled. For _golangci-lint_ though, you can also use it as a CLI tool, although a language server version is more convenient.

## Go backend

The topmost functions, like those in the _main_ and _routes_ packages,   inject the dependencies (like logger and database connection) into the functions in _models_ and/or _controllers_ packages.  

Resources intensive operations called inide other functions, should be parallelized with goroutines, but make sure they are properly synchronized if needed. Note however, that it is preferrable to use smaller, dedicated functions with limited or no side effects instead of unnecessarily large ones.

Long, more complex functions should include a brief comment summarizing their mechanism of action.

For error messages, prefer to use the status codes provided by fiber, to avoid unnecessary overlap with _net/http_

## Frontend

Avoid doing too much computation on the client's side. If something can be done by an API endpoint in the Go backend, then do it this way.

Keep all type definitions in _fe/src/types_ and the more complex functions for managing state in _fe/src/stores_

The goal is an accessible, customizable site which can be feature rich but must not be a potential source of information overload and which works fast thanks to heavy use of SSR.

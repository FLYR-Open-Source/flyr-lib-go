# Contributing Guidelines

The flyr-lib-go project accepts contributions via GitHub pull requests. This document outlines the process
to help get your contribution accepted.

## Pre-commits

Wehn you raise a PR, it will also try to run pre-commits. We use the [pre-commit framework](https://github.com/pre-commit/pre-commit), that you can install via brew:

```bash
brew install pre-commit
```

Then run:

```bash
make git-hooks
```

## Lint

```bash
brew install golangci-lint
```

```bash
make lint
```

### TBD

Add more details for contributions

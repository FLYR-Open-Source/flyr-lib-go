# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## What this is

`github.com/FLYR-Open-Source/flyr-lib-go` — a public, MIT-licensed Go library (not a service) providing OpenTelemetry-based observability (`logger`, `monitoring`) and test fakes (`pkg/testhelpers`) for FLYR services. Consumers import it; there is no `main`.

## Commands

```bash
make lint            # golangci-lint run --enable gofmt,testifylint,misspell (no .golangci.yml; flags live in the Makefile)
make test            # gotestsum -- -short --cover over $(TEST_DIRS)
make tidy            # go mod tidy in EVERY go.mod (root + each example)
make add-license     # add MIT headers; CI fails without them on every .go file
make check-license
make coverage        # per-package coverage over $(COVER_PKGS); prints funcs below 100% + total, writes coverage.out / coverage.html
make coverage-check  # make coverage, then fail if total < COVERAGE_MIN (default 90)
make benchmark
make pre-commit-hook # installs scripts/license-pre-commit.sh as a git hook

go test ./monitoring/tracer -run TestName -v   # single test
```

- `make test` deliberately **excludes** `examples/`, `pkg/` and `scripts/` (`EXCLUDED_TOP_DIRS`), so `pkg/testhelpers` tests are not run by it. CI instead runs `gotestsum ... ./...` from the root, which does include `pkg/`. Use `go test ./...` to match CI.
- `make coverage` scope: all root-module packages except `examples/**`; `pkg/testhelpers` is included. Per-package (no `-coverpkg`), same as CI's `coverage.out` for Sonar. Threshold: `make coverage-check COVERAGE_MIN=80` overrides the default 90.
- `golangci-lint`, `gotestsum` and `addlicense` are not vendored; install them separately (`brew install golangci-lint`; `go install github.com/google/addlicense@latest`).
- Every `.go` file needs the MIT header with `FLYR, Inc` (see any existing file; `make add-license` writes it). Package clauses also carry import comments, e.g. `package tracer // import "github.com/FLYR-Open-Source/flyr-lib-go/monitoring/tracer"` — keep that on new packages.

## Releases and commits

Releases are automated by semantic-release (`.releaserc.js`) from **Conventional Commits** on `main` (also `release/N.x`, `next`, `alpha`, `beta`, `rc`). The commit type decides the version bump, so use `fix:`/`feat:`/`feat!:` accurately. Dependency bumps arrive from Renovate as `fix(deps): ...`.

PR titles (and so squash-merge commit subjects) put the Jira ticket right after the type: `<type>(<scope>): <TICKET> <summary>`, e.g. `ci: CPE-3645 pin semantic-release plugins to fix Release Validation`. Bot PRs (Renovate/Dependabot) are exempt.

## Architecture

Three layers; public packages depend on `internal/`, never the reverse.

- **`internal/config`** — env-var config (via `caarlos0/env/v11`) for monitoring (`OTEL_*`) and logger (`LOG_LEVEL`). Each is a **process-wide singleton behind `sync.Once`**; `ResetMonitoringConfig()`-style helpers and the `withEnvironment` option exist so tests can inject env and re-parse. `commontags.go` holds the shared OTel semantic-convention attribute names used by both spans and logs.
- **`internal/span`, `internal/logger`, `internal/utils`, `internal/version`** — the implementation behind the public wrappers.
- **Public packages**
  - `logger/` — wraps `log/slog` with a JSON handler plus a `TracingHandler` (`internal/logger`) chained via `slog-multi`. Package-level functions (`logger.Info(ctx, ...)`, `logger.Error(ctx, msg, err, ...)`) require `InitLogger()` first.
  - `monitoring/tracer`, `monitoring/meter` — initialise the global OTel Tracer/Meter providers from config (OTLP gRPC/HTTP or stdout when `OTEL_EXPORTER_OTLP_TEST=true`). Before initialisation, the default tracer is a **noop**, so tracing calls are always safe. Always pair with `ShutdownTracerProvider` / `ShutdownMeterProvider`.
  - `monitoring/http`, `monitoring/pubsub` (**deprecated**, use `monitoring/pubsub/v2`), `monitoring/rabbitmq`, `monitoring/middleware` (Gin + Chi) — trace propagation for each transport.
  - `pkg/testhelpers` — `monitoring` (FakeTracer/FakeSpan/fake OTLP servers), `grpc`, `gcp/pubsub` (emulator helpers) for consumer tests.

### The logger ↔ span coupling (cross-file behaviour)

Logging with a `ctx` that carries a span has side effects that live in different files:
- `internal/logger/tracehandler.go` adds `trace_id`/`span_id` to every record **except Debug**.
- `logger/attrs.go` (the `Attribute` builder, `WithOutInjectingAttrsToSpan` for Debug) calls into `logger/span.go`, which copies log attributes onto the active span (flattening nested JSON into dotted tags).
- `logger.Error` additionally marks the span errored and records the error as a span event.

`InitLogger()` installs the handler chain as `slog.Default()`, and the package-level functions log through it.

Pass the ctx returned by span creation into log calls, or this correlation is silently lost.

The `Span` returned by this library's tracer (`internal/span`) wraps the OTel span and adds `EndWithError(err)` and `EndSuccessfully()`; spans must always be ended.

### Multi-module layout

The root `go.mod` covers the library. `examples/**` are **each their own Go module** (8 of them) wired to the local library via `replace github.com/FLYR-Open-Source/flyr-lib-go => ../../..`. Consequences: `go build ./...` / `go test ./...` at the root do not touch examples; after changing a public API, build the affected examples separately (`cd examples/<path> && go build ./...`) and run `make tidy` if dependencies changed.

## Docs

Per-package READMEs (`logger/README.md`, `monitoring/README.md`) hold the env-var tables and behavioural docs; update the relevant table when adding or changing a config env var in `internal/config`.

---
name: pre-pr
description: Run the CI-equivalent checks (lint, license headers, tests, example builds) before opening a PR
disable-model-invocation: true
allowed-tools: Bash(make lint), Bash(make check-license), Bash(go test *), Bash(make coverage-check), Bash(go build *), Bash(find examples *), Bash(git status *), Bash(git log *), Bash(git diff *)
---

# pre-pr

Run these in order and report a pass/fail line per step. Keep going after a failure so every problem is reported at once; do not fix anything unless the user asks.

1. **Lint** — `make lint`. CI runs golangci-lint with gofmt, testifylint and misspell.
2. **License headers** — `make check-license`. On failure, `make add-license` fixes it (ask first, it edits files).
3. **Tests** — `go test ./...` from the repo root. Use this rather than `make test`: CI runs `./...`, which includes `pkg/testhelpers`, while `make test` skips `pkg/`.
4. **Coverage** — `make coverage-check`. Fails if total coverage (all root-module packages except `examples/**`, `pkg/testhelpers` included) is below `COVERAGE_MIN` (default 90). The per-function report lists what is under 100%.
5. **Examples** — each `examples/**` directory with a `go.mod` is a separate module that CI does not build, so build each one:
   ```bash
   for m in $(find examples -name go.mod -exec dirname {} \;); do
     (cd "$m" && go build ./... ) || echo "FAILED: $m"
   done
   ```
   Skip this step if the diff touches no exported API (check `git diff main...HEAD --stat`).
6. **Commit messages** — `git log main..HEAD --format=%s`. Every subject must be a Conventional Commit (`fix:`, `feat:`, `feat!:`, `docs:`...), because semantic-release derives the version bump from them. Flag any that are not, and flag a `feat`/`fix` type that does not match the change.

If `golangci-lint` or `addlicense` is not installed, say so and tell the user how to install it (`brew install golangci-lint`, `go install github.com/google/addlicense@latest`) instead of skipping silently.

End with a one-line verdict: ready for a PR, or the list of failing steps.

GO := go

EXCLUDED_TOP_DIRS := examples pkg scripts benchmark
ALL_GO_DIRS := $(shell find . -type f -name '*.go' -exec dirname {} \; | sort -u)
EXCLUDED_PATHS := $(addprefix ./,$(EXCLUDED_TOP_DIRS))
TEST_DIRS := $(filter-out $(foreach d,$(EXCLUDED_PATHS),$(d)% $(d)), $(ALL_GO_DIRS))
BENCHMARK_DIRS := $(filter-out $(foreach d,$(EXCLUDED_PATHS),$(d)% $(d)), $(ALL_GO_DIRS))

# Minimum total statement coverage (percent) enforced by `make coverage-check`.
COVERAGE_MIN ?= 90
# Root-module packages minus examples/** (demo programs). pkg/testhelpers is included.
COVER_PKGS := $(shell $(GO) list ./... | grep -v /examples/)


.PHONY: tidy
tidy:
	@find . -name go.mod -execdir go mod tidy \;

.PHONY: lint
lint:
	golangci-lint run  --timeout 5m --enable gofmt,testifylint,misspell -v

.PHONY: test
test:
	gotestsum -- -short --cover $(TEST_DIRS)

.PHONY: test-coverage
test-coverage:
	go test -coverprofile coverage.out $(TEST_DIRS) -json > unit-test-results.json

.PHONY: coverage
coverage:
	$(GO) test -short -covermode=atomic -coverprofile=coverage.out $(COVER_PKGS)
	@echo "Functions below 100% coverage:"
	@$(GO) tool cover -func=coverage.out | awk '$$1 == "total:" || $$NF != "100.0%"'
	@$(GO) tool cover -html=coverage.out -o coverage.html
	@echo "HTML report: coverage.html"

.PHONY: coverage-check
coverage-check: coverage
	@total=$$($(GO) tool cover -func=coverage.out | awk '$$1 == "total:" { sub(/%/, "", $$NF); print $$NF }'); \
	if awk -v t="$$total" -v m="$(COVERAGE_MIN)" 'BEGIN { exit !(t + 0 >= m + 0) }'; then \
		echo "coverage $$total% >= $(COVERAGE_MIN)%"; \
	else \
		echo "coverage check FAILED: total coverage $$total% is below the required $(COVERAGE_MIN)% (override with COVERAGE_MIN=<percent>)" >&2; \
		exit 1; \
	fi

.PHONY: benchmark
benchmark: $(BENCHMARK_DIRS:%=benchmark/%)

benchmark/%:
	@echo "Benchmarking $* ..."
	@cd $* && $(GO) list ./... \
		| grep -v third_party \
		| xargs $(GO) test -run=xxxxxMatchNothingxxxxx -bench=. -benchmem || true

.PHONY: check-license
check-license:
	@find . -type f -name "*.go" | xargs addlicense -check -l mit -f ./LICENSE -c "FLYR, Inc"

.PHONY: add-license
add-license:
	@find . -type f -name "*.go" | xargs addlicense -l mit -f ./LICENSE -c "FLYR, Inc"

.PHONY: pre-commit-hook
pre-commit-hook:
	cp ./scripts/license-pre-commit.sh .git/hooks/pre-commit
	chmod +x .git/hooks/pre-commit

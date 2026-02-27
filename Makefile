GO := go

EXCLUDED_TOP_DIRS := examples pkg scripts benchmark
ALL_GO_DIRS := $(shell find . -type f -name '*.go' -exec dirname {} \; | sort -u)
EXCLUDED_PATHS := $(addprefix ./,$(EXCLUDED_TOP_DIRS))
TEST_DIRS := $(filter-out $(foreach d,$(EXCLUDED_PATHS),$(d)% $(d)), $(ALL_GO_DIRS))
BENCHMARK_DIRS := $(filter-out $(foreach d,$(EXCLUDED_PATHS),$(d)% $(d)), $(ALL_GO_DIRS))


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

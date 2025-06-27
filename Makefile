# The files that will be included in the test coverage report
TEST_FILES=$(shell go list ./... | grep -v ./examples | grep -v ./pkg)

TOOLS_MOD_DIR := ./pkg
EXAMPLES_MOD_DIR := ./examples

ALL_DOCS := $(shell find . -name '*.md' -type f | sort)
ALL_GO_MOD_DIRS := $(shell find . -type f -name 'go.mod' -exec dirname {} \; | sort)
OTEL_GO_MOD_DIRS := $(filter-out $(TOOLS_MOD_DIR), $(EXAMPLES_MOD_DIR), $(ALL_GO_MOD_DIRS))

GO = go

.PHONY: tidy
tidy:
	go mod tidy

.PHONY: lint
lint:
	golangci-lint run  --timeout 5m --enable gofmt,testifylint,misspell -v

.PHONY: test
test:
	gotestsum -- -short --cover $(TEST_FILES)

.PHONY: test-coverage
test-coverage:
	go test -coverprofile coverage.out $(TEST_FILES) -json > unit-test-results.json

.PHONY: test-benchmark
test-benchmark:
	@for pkg in $(TEST_FILES); do \
		pkg_file=$$(echo $$pkg | sed 's|/|_|g'); \
		go test -run=xxxxxMatchNothingxxxxx -bench=. -benchtime=20s -benchmem -cpu=1,2,4,12 -timeout 20m \
			-o /dev/null \
			-cpuprofile="benchmark/$$pkg_file.cpu.prof" \
			-memprofile="benchmark/$$pkg_file.mem.prof" $$pkg; \
	done

.PHONY: benchmark
benchmark: $(OTEL_GO_MOD_DIRS:%=benchmark/%)
benchmark/%:
	@echo "$(GO) test -run=xxxxxMatchNothingxxxxx -bench=. -benchmem $*..." \
		&& cd $* \
		&& $(GO) list ./... \
		| grep -v third_party \
		| xargs $(GO) test -run=xxxxxMatchNothingxxxxx -bench=.


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

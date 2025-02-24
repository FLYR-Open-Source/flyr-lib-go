# The files that will be included in the test coverage report
TEST_FILES=$(shell go list ./... | grep -v ./examples | grep -v ./pkg)

.PHONY: tidy
tidy:
	go mod tidy

.PHONY: docs
docs:
	$(GOBIN)/godoc -http=:6060
	# Visit: http://localhost:6060/pkg/github.com/FlyrInc/flyr-lib-go/

.PHONY: lint
lint:
	golangci-lint run  --timeout 5m --enable gofmt,testifylint,misspell -v

.PHONY: test
test:
	gotestsum -- -short --cover $(TEST_FILES)

.PHONY: test-coverage
test-coverage:
	go test -coverprofile coverage.out $(TEST_FILES) -json > unit-test-results.json

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

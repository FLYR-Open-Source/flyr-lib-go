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
	gotestsum -- -short --cover ./...

.PHONY: test-coverage
test-coverage:
	go test -coverprofile=cover/coverage.out ./...
	go tool cover -html=cover/coverage.out

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

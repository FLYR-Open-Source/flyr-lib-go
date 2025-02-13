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

.PHONY: git-hooks
git-hooks:
	pre-commit install

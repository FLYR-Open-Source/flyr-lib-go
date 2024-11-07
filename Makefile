tidy:
	go mod tidy

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

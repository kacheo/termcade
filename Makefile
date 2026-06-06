.PHONY: lint test coverage build clean update-golden test-integration test-regression test-all

lint:
	golangci-lint run ./...

test:
	go test -race -covermode=atomic -coverprofile=coverage.out ./...

coverage: test
	go tool cover -html=coverage.out

build:
	go build -o main ./cmd

update-golden:
	go test -run TestGoldenRender -args -update ./games/...
	go test -tags=regression -run TestRegression -args -update ./cmd/...

clean:
	rm -f coverage.out main

test-integration:
	@echo "tmvgs has no subprocess CLI; integration tests not applicable. See tests/regression/ for cmd-level goldens."
	@exit 0

test-regression:
	go test -tags=regression -timeout=5m -v ./cmd/...

test-all: test test-regression
	@echo "All test suites completed"

.PHONY: lint test coverage build clean

lint:
	golangci-lint run ./...

test:
	go test -race -covermode=atomic -coverprofile=coverage.out ./...

coverage: test
	go tool cover -html=coverage.out

build:
	go build -o main ./cmd

clean:
	rm -f coverage.out main

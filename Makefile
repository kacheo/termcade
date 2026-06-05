.PHONY: lint test coverage build clean update-golden

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

clean:
	rm -f coverage.out main

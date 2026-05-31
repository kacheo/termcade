.PHONY: build test test-coverage clean

build:
	go build ./...

test:
	go test ./...

test-coverage:
	@go test ./... -coverprofile=coverage.out -covermode=atomic
	@coverage=$$(go tool cover -func=coverage.out | grep "^total:" | awk '{print $$3}' | tr -d '%') && \
	echo "Coverage: $$coverage%" && \
	result=$$(awk -v c="$$coverage" 'BEGIN {if (c < 80) print "fail"; else print "pass"}') && \
	if [ "$$result" = "fail" ]; then \
		echo "ERROR: Coverage must be at least 80%"; \
		rm -f coverage.out; \
		exit 1; \
	fi && \
	rm -f coverage.out && \
	echo "Coverage check passed (80%+ required)"

clean:
	rm -f coverage.out

lint:
	go vet ./...

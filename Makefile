.PHONY: test test-cov lint build clean bench

test:
	go test ./...

test-cov:
	go clean -testcache
	go test -coverprofile=coverage.out ./internal/... 2>/dev/null || true
	@if [ -f coverage.out ]; then \
		go tool cover -func=coverage.out; \
		rm coverage.out; \
	fi

lint:
	golangci-lint run ./...

build:
	go build -o 0xfer ./cmd/server

clean:
	rm -f 0xfer coverage.out
	go clean

bench:
	go test -bench=. -benchtime=1x ./...
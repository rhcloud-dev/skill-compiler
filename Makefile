VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
LDFLAGS := -ldflags "-s -w -X main.version=$(VERSION)"

.PHONY: build test lint clean

build: clean
	go build $(LDFLAGS) -o dist/sc ./cmd/sc/

test:
	go test ./...

lint:
	golangci-lint run ./...

clean:
	rm -rf dist

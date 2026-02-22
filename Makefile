VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
LDFLAGS := -ldflags "-s -w -X main.version=$(VERSION)"

.PHONY: build test clean

build: clean
	go build $(LDFLAGS) -o sc ./cmd/sc/

test:
	go test ./...

clean:
	rm -f sc

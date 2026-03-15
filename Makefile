BINARY := hydramancer
MODULE := github.com/cederikdotcom/hydramancer
VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo dev)
LDFLAGS := -ldflags "-X $(MODULE)/internal/cli.Version=$(VERSION)"

.PHONY: build run clean vet fmt

build:
	go build $(LDFLAGS) -o bin/$(BINARY) ./cmd/hydramancer

run: build
	./bin/$(BINARY) serve --dev

clean:
	rm -rf bin/

vet:
	go vet ./...

fmt:
	gofmt -w .

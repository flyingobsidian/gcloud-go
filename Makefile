default: build

BINARY := gcloud-go
VERSION := $(shell git describe --tags 2>/dev/null || echo dev)
LDFLAGS := -s -w -X github.com/flyingobsidian/gcloud-go/cmd.Version=$(VERSION)

.PHONY: build test test-v clean

build:
	CGO_ENABLED=0 go build -ldflags="$(LDFLAGS)" -o bin/$(BINARY) .

test:
	go test ./...

test-v:
	go test ./... -v

clean:
	rm -f bin/$(BINARY)

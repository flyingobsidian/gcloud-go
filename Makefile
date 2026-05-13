default: build

BINARY := gcloud-go
MODULE := github.com/flyingobsidian/gcloud-golang-cli
LDFLAGS := -s -w

.PHONY: build test test-v test-integration clean

build:
	CGO_ENABLED=0 go build -ldflags="$(LDFLAGS)" -o bin/$(BINARY) .

test:
	go test ./...

test-v:
	go test ./... -v

test-integration:
	go test -tags=integration -v

clean:
	rm -f bin/$(BINARY)

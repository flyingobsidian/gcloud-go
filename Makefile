BINARY := gcloud-go
MODULE := github.com/flyingobsidian/gcloud-golang-cli
LDFLAGS := -s -w
GO := /home/ashley/go1.24.4/bin/go

.PHONY: build test test-v test-integration clean

build:
	CGO_ENABLED=0 $(GO) build -ldflags="$(LDFLAGS)" -o bin/$(BINARY) .

test:
	$(GO) test ./...

test-v:
	$(GO) test ./... -v

test-integration:
	$(GO) test -tags=integration -v

clean:
	rm -f bin/$(BINARY)

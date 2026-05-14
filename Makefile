default: build

BINARY := gcloud-go
LDFLAGS := -s -w

.PHONY: build test test-v clean

build:
	CGO_ENABLED=0 go build -ldflags="$(LDFLAGS)" -o bin/$(BINARY) .

test:
	go test ./...

test-v:
	go test ./... -v

clean:
	rm -f bin/$(BINARY)

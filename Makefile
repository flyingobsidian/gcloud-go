default: build

BINARY := gcloud-go
VERSION := $(shell git describe --tags 2>/dev/null || echo dev)
LDFLAGS := -s -w -X github.com/flyingobsidian/gcloud-go/cmd.Version=$(VERSION)

# Limit go commands to half the CPUs and half the memory, without swap.
CPUS := $(shell n=$$(( $$(nproc) / 2 )); echo $$(( n > 0 ? n : 1 )))
CPU_QUOTA := $(shell echo $$(( $$(nproc) * 50 )))%
MEM := $(shell echo $$(( $$(free -b | awk '/Mem:/{print $$2}') / 2 )))
GO := systemd-run --user --scope --quiet -p CPUQuota=$(CPU_QUOTA) -p MemoryMax=$(MEM) -p MemorySwapMax=0 \
	env GOMAXPROCS=$(CPUS) GOMEMLIMIT=$(MEM) GOFLAGS=-p=$(CPUS) go

.PHONY: build test test-v clean

build:
	CGO_ENABLED=0 $(GO) build -ldflags="$(LDFLAGS)" -o bin/$(BINARY) .

fmt:
	$(GO) fmt ./...

test:
	$(GO) test ./...

test-v:
	$(GO) test ./... -v

vet:
	$(GO) vet ./...

clean:
	rm -f bin/$(BINARY)

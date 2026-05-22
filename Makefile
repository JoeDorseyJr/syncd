.PHONY: build test integration-test lint clean install uninstall

PREFIX ?= /usr/local
VERSION := $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")

build:
	CGO_ENABLED=0 go build -ldflags "-X main.version=$(VERSION)" -o bin/syncd ./cmd/syncd

test:
	go test ./...

integration-test:
	SYNCD_INTEGRATION=1 go test -tags integration ./test/...

lint:
	go vet ./...

clean:
	rm -rf bin/

install: build
	mkdir -p $(PREFIX)/bin
	cp bin/syncd $(PREFIX)/bin/syncd

uninstall:
	rm -f $(PREFIX)/bin/syncd

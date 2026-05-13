.PHONY: build test lint clean

build:
	CGO_ENABLED=0 go build -o bin/syncd ./cmd/syncd

test:
	go test ./...

lint:
	go vet ./...

clean:
	rm -rf bin/

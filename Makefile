.PHONY: build test integration-test lint clean

build:
	CGO_ENABLED=0 go build -o bin/syncd ./cmd/syncd

test:
	go test ./...

integration-test:
	SYNCD_INTEGRATION=1 go test -tags integration ./test/...

lint:
	go vet ./...

clean:
	rm -rf bin/

.PHONY: build test lint clean
BINARY := bin/notes

build:
	@mkdir -p bin
	go build -o $(BINARY) ./cmd/notes

test:
	go test ./...

lint:
	go vet ./...
	# TODO: enable golangci-lint once dev-deps decision lands

clean:
	rm -rf bin

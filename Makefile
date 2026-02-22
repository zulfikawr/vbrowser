.PHONY: build test lint clean

build:
	go build -o vbrowser ./cmd/vbrowser

test:
	go test -v ./...

lint:
	go fmt ./...
	go run github.com/golangci/golangci-lint/cmd/golangci-lint@latest run ./...

clean:
	rm -f vbrowser
	rm -rf dist/

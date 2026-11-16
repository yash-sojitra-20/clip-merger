.PHONY: all build run test test-v clean tidy lint

APP_NAME = clip-merger
GO ?= /usr/local/go/bin/go

all: test build

build:
	@mkdir -p bin
	$(GO) build -o bin/$(APP_NAME) ./cmd/clip-merger

run:
	$(GO) run ./cmd/clip-merger

test:
	$(GO) test ./...

test-v:
	$(GO) test -v -race ./...

tidy:
	$(GO) mod tidy

clean:
	@rm -rf bin/ tmp/ data/


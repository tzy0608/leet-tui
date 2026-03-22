.PHONY: build run test clean generate

BINARY := leet-tui
BUILD_DIR := bin

build:
	go build -o $(BUILD_DIR)/$(BINARY) ./cmd/leet-tui

run:
	go run ./cmd/leet-tui

test:
	go test ./... -v

clean:
	rm -rf $(BUILD_DIR)

generate:
	sqlc generate

lint:
	golangci-lint run ./...

.DEFAULT_GOAL := build

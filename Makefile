.PHONY: all build test vet lint tidy clean run install-tools

VERSION ?= dev
LDFLAGS := -X nea-ai/internal/app.Version=$(VERSION)
BIN_DIR := dist
BIN_NAME := nea-ai
ifeq ($(OS),Windows_NT)
	BIN := $(BIN_DIR)/$(BIN_NAME).exe
else
	BIN := $(BIN_DIR)/$(BIN_NAME)
endif

all: vet test build

build:
	@mkdir -p $(BIN_DIR)
	go build -ldflags "$(LDFLAGS)" -o $(BIN) ./cmd/nea-ai

test:
	go test ./... -count=1

vet:
	go vet ./...

lint:
	golangci-lint run --timeout=5m

tidy:
	go mod tidy

clean:
	rm -rf $(BIN_DIR)

run: build
	$(BIN) $(ARGS)

install-tools:
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@v1.64.5

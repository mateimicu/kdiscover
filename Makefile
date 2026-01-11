.PHONY: all build test lint coverage clean golden-update check

# Binary name
BINARY_NAME=kdiscover

# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOTEST=$(GOCMD) test
GOMOD=$(GOCMD) mod
GOFMT=$(GOCMD) fmt

# Build flags
VERSION ?= dev
COMMIT ?= $(shell git rev-parse --short HEAD 2>/dev/null || echo "none")
DATE ?= $(shell date -u +"%Y-%m-%dT%H:%M:%SZ")
LDFLAGS=-ldflags "-X main.version=$(VERSION) -X main.commit=$(COMMIT) -X main.date=$(DATE)"

all: check build

## build: Build the binary
build:
	$(GOBUILD) $(LDFLAGS) -o $(BINARY_NAME) main.go

## test: Run all tests with race detection
test:
	$(GOTEST) -race -v ./...

## test-short: Run tests without verbose output
test-short:
	$(GOTEST) -race ./...

## lint: Run golangci-lint
lint:
	golangci-lint run ./...

## coverage: Generate test coverage report
coverage:
	$(GOTEST) -race -coverprofile=coverage.txt -covermode=atomic ./...
	$(GOCMD) tool cover -html=coverage.txt -o coverage.html
	@echo "Coverage report generated: coverage.html"

## coverage-func: Show coverage by function
coverage-func:
	$(GOTEST) -race -coverprofile=coverage.txt -covermode=atomic ./...
	$(GOCMD) tool cover -func=coverage.txt

## golden-update: Update golden test files
golden-update:
	$(GOTEST) -v ./... -update

## check: Run lint and tests
check: lint test

## fmt: Format code
fmt:
	$(GOFMT) ./...

## tidy: Tidy go modules
tidy:
	$(GOMOD) tidy

## clean: Clean build artifacts
clean:
	rm -f $(BINARY_NAME)
	rm -f coverage.txt
	rm -f coverage.html

## help: Show this help message
help:
	@echo "Usage: make [target]"
	@echo ""
	@echo "Targets:"
	@sed -n 's/^## //p' $(MAKEFILE_LIST) | column -t -s ':'

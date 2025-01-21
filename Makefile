BINARY_NAME := vhicmd
SRC := ./...

# Build flags
BUILD_FLAGS := -ldflags="-s -w"
CGO_ENABLED := 0

# Targets
.PHONY: all build clean test

all: build

build: ## Build the project with CGO disabled and debugging off
	CGO_ENABLED=$(CGO_ENABLED) go build $(BUILD_FLAGS) -o bin/$(BINARY_NAME) main.go

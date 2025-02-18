BINARY_NAME := vhicmd
SRC := ./...
CGO_ENABLED := 0

BUILD_TIME := $(shell date -u +"%Y%m%d-%H%M%S")
LDFLAGS := -s -w -X github.com/jessegalley/vhicmd/cmd.buildTime=${BUILD_TIME}
BUILD_FLAGS := -ldflags="$(LDFLAGS)"
GODOC := $(HOME)/go/bin/godoc

.PHONY: all build clean docs docs-html

all: build

build:
	CGO_ENABLED=$(CGO_ENABLED) go build $(BUILD_FLAGS) -o bin/$(BINARY_NAME) main.go

clean:
	rm -f bin/$(BINARY_NAME)
	rm -f api/doc.go
	rm -rf docs

# Creates a docs/api.html in Solarized style
docs-html:
	@echo "Generating HTML documentation with Solarized theme..."
	@mkdir -p docs
	@go run tools/docgen/main.go
	@echo "Documentation written to docs/api.html"

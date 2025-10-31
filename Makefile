BINARY := interchange

# If VERSION isn't provided, use the latest git commit hash
VERSION ?= $(shell git rev-parse --short HEAD)

build:
	@echo "Building $(BINARY) with version $(VERSION)"
	@go build -ldflags="-X 'templates.ServerVersionString=$(BINARY)/$(VERSION)' -X 'main.Version=$(BINARY)/$(VERSION)'" -o $(BINARY) .

clean:
	rm -f $(BINARY)

help:
	@echo "Usage:"
	@echo "  make build VERSION=v1.2.3   Build with specific version"
	@echo "  make build                  Build using latest git commit hash"
	@echo "  make clean                  Remove built binary"
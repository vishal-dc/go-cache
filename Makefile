# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
BINARY_NAME=go-cache
BINARY_LINUX=$(BINARY_NAME)_linux

# All target: build the binary
all: clean test build build-linux docker-build

# Build the binary
build:
	$(GOBUILD) -o ./bin/$(BINARY_NAME) -v ./cmd

# Run tests
test:
	$(GOTEST) -v ./...

# Clean build files
clean:
	$(GOCLEAN)
	rm -f ./bin/$(BINARY_NAME)
	rm -f ./bin/$(BINARY_LINUX)

# Run the application
run:
	$(GOBUILD) -o ./bin/$(BINARY_NAME) -v ./cmd
	./bin//$(BINARY_NAME)

# Install dependencies
# deps:
# 	$(GOGET) -u github.com/stretchr/testify
# 	$(GOGET) -u go.uber.org/zap

# Cross compilation for Linux
build-linux:
	GOOS=linux GOARCH=amd64 $(GOBUILD) -o ./bin/$(BINARY_LINUX) -v ./cmd

# Docker build
docker-build:
	docker build -t $(BINARY_NAME) .
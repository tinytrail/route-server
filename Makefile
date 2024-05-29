# Go parameters
GOCMD = go
GOBUILD = $(GOCMD) build
GOCLEAN = $(GOCMD) clean
GOTEST = $(GOCMD) test
GOGET = $(GOCMD) get
PROTOC = protoc

# Main package name
PACKAGE_NAME = server

# Output binary name
BINARY_NAME = routeguide

# Build target
build:
	buf generate

# Clean target
clean:
	$(GOCLEAN)
	rm -rf gen
	rm -f $(BINARY_NAME)

# Test target
test:
	$(GOTEST) -v ./...

# Default target
default: build
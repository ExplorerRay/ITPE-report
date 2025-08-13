# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOMOD=$(GOCMD) mod
GOCLEAN=$(GOCMD) clean

BINARY_NAME=itpe-report

BINARY_DIR=bin
MAIN_GO_PATH=./cmd/itpe-report

all: build clean deps

## Build the application
.PHONY: build
build: deps
	mkdir -p $(BINARY_DIR)
	$(GOBUILD) -o $(BINARY_DIR)/$(BINARY_NAME) $(MAIN_GO_PATH)

## Install the application dependencies
.PHONY: deps
deps:
	$(GOMOD) tidy
	$(GOMOD) verify

## Clean the build artifacts
.PHONY: clean
clean:
	$(GOCLEAN)
	rm -rf $(BINARY_DIR)

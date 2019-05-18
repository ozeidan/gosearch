# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
GOBASE := $(shell pwd)
SERVER_BINARY_NAME=goSearchServer
CLIENT_BINARY_NAME=goSearchClient

all: build

build:
	cd $(GOBASE)/cmd/server; $(GOBUILD) -o $(GOBASE)/$(SERVER_BINARY_NAME) -v
	cd $(GOBASE)/cmd/client; $(GOBUILD) -o $(GOBASE)/$(CLIENT_BINARY_NAME) -v
test:
	$(GOTEST) -v ./...
clean:
	cd $(GOBASE)/cmd/server; $(GOCLEAN)
	cd $(GOBASE)/cmd/client; $(GOCLEAN)
	rm -f $(SERVER_BINARY_NAME)
	rm -f $(CLIENT_BINARY_NAME)
# run:
# 	$(GOBUILD) -o $(BINARY_NAME) -v ./...
# 	./$(SERVER_BINARY_NAME) &
deps:
	$(GOGET)


# Cross compilation
build-linux:
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 $(GOBUILD) -o $(BINARY_UNIX) -v

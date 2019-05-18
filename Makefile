# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
GOINSTALL=$(GOCMD) install
GOBASE := $(shell pwd)
SYSTEMD_SERVICE_FILE=./init/goSearch.service
SERVER_BINARY_NAME=goSearchServer
CLIENT_BINARY_NAME=goSearchClient

all: build

build-server:
	cd $(GOBASE)/cmd/server; $(GOBUILD) -o $(GOBASE)/$(SERVER_BINARY_NAME) -v

build-client:
	cd $(GOBASE)/cmd/client; $(GOBUILD) -o $(GOBASE)/$(CLIENT_BINARY_NAME) -v

build: build-server build-client

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
#
#

deps:
	$(GOGET)

install: build-server move

move:
	$(GOINSTALL) -v $(GOBASE)/cmd/client
	sudo mv $(SERVER_BINARY_NAME) /usr/bin
	sudo cp $(SYSTEMD_SERVICE_FILE) /etc/systemd/system/
	sudo systemctl daemon-reload
	sudo systemctl enable goSearch
	sudo systemctl start goSearch


# Cross compilation
build-linux:
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 $(GOBUILD) -o $(BINARY_UNIX) -v

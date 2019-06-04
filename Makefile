# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
GOINSTALL=$(GOCMD) install
GOBASE := $(shell pwd)
SYSTEMD_SERVICE_FILE=./init/gosearch.service
SERVER_BINARY_NAME=gosearchServer
CLIENT_BINARY_NAME=gosearch

all: build

build-server:
	cd $(GOBASE)/cmd/server; $(GOBUILD) -v -o $(GOBASE)/$(SERVER_BINARY_NAME)

build-client:
	cd $(GOBASE)/cmd/client; $(GOBUILD) -v -o $(GOBASE)/$(CLIENT_BINARY_NAME)

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

install: build-server build-client move

move:
	sudo mv $(SERVER_BINARY_NAME) /usr/bin
	sudo mv $(CLIENT_BINARY_NAME) /usr/bin
	sudo cp $(SYSTEMD_SERVICE_FILE) /etc/systemd/system/
	sudo systemctl daemon-reload
	sudo systemctl enable gosearch
	sudo systemctl stop gosearch
	sudo systemctl start gosearch


# Cross compilation
build-linux:
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 $(GOBUILD) -o $(BINARY_UNIX) -v

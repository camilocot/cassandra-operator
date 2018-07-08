KUBERNETES_CONFIG ?= /etc/kubernetes/admin.conf
PKGS := $(shell go list ./... | grep -v /vendor)
IMAGE ?= camilocot/operator:v0.0.1

# Go parameters
GOCMD=go
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
BIN_DIR := $(GOPATH)/bin
GOMETALINTER := $(BIN_DIR)/gometalinter

all: deps lint test build

test:
	KUBERNETES_CONFIG=$(KUBERNETES_CONFIG) $(GOTEST) -v -timeout 120s -short $(PKGS)

clean:
	$(GOCLEAN)

deps:
	$(GOGET) -u github.com/golang/dep/cmd/dep
	dep ensure -v

$(GOMETALINTER):
	$(GOGET) -u github.com/alecthomas/gometalinter
	gometalinter --install

lint: $(GOMETALINTER)
	gometalinter -d --fast --disable gosimple --disable staticcheck --deadline=240s --exclude=zz --vendor --tests ./...

build:
	./tmp/build/build.sh
	IMAGE=$(IMAGE) ./tmp/build//docker_build.sh

.PHONY: all build test lint deps

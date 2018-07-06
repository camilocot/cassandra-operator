KUBERNETES_CONFIG ?= /etc/kubernetes/admin.conf
PKGS := $(shell go list ./... | grep -v /vendor)

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
	go get -u github.com/alecthomas/gometalinter
	gometalinter --install

lint: $(GOMETALINTER)
	gometalinter --disable=aligncheck --disable=unconvert --disable=gotype --disable=errcheck --disable=varcheck --disable=structcheck --disable gosimple --disable staticcheck --disable interfacer --deadline=20s --exclude=zz --vendor --tests ./...

install:
	go install github.com/operator-framework/operator-sdk/commands/operator-sdk

build: install
	operator-sdk build camilocot/operator:v0.0.1

.PHONY: all build test lint deps

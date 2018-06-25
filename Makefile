KUBERNETES_CONFIG ?= /etc/kubernetes/admin.conf
PKGS := $(shell go list ./... | grep -v /vendor)

# Go parameters
GOCMD=go
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get

.PHONY: test
test: lint
	KUBERNETES_CONFIG=$(KUBERNETES_CONFIG) $(GOTEST) -v -timeout 120s -short $(PKGS)

clean:
	$(GOCLEAN)

deps:
	$(GOGET) -u github.com/golang/dep/cmd/dep
	dep ensure

BIN_DIR := $(GOPATH)/bin
GOMETALINTER := $(BIN_DIR)/gometalinter

$(GOMETALINTER):
	go get -u github.com/alecthomas/gometalinter
	gometalinter --install &> /dev/null

.PHONY: lint
lint: $(GOMETALINTER)
	gometalinter --disable=aligncheck --disable=unconvert --disable=gotype --disable=errcheck --disable=varcheck --disable=structcheck --disable gosimple --disable staticcheck --disable interfacer --deadline=20s --exclude=zz --vendor --tests ./...

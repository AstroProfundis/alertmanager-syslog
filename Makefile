.PHONY: check server
.DEFAULT_GOAL := default

GOOS    := $(if $(GOOS),$(GOOS),linux)
GOARCH  := $(if $(GOARCH),$(GOARCH),amd64)
GOENV   := GO111MODULE=on CGO_ENABLED=0 GOOS=$(GOOS) GOARCH=$(GOARCH)
GO      := $(GOENV) go
GOBUILD := $(GO) build $(BUILD_FLAG)

COMMIT    := $(shell git describe --no-match --always --dirty)

PKG := github.com/AstroProfundis/alertmanager-syslog
LDFLAGS := -w -s
LDFLAGS += -X "$(PKG)/pkg/version.GitHash=$(COMMIT)"

default: all

all: check server

server:
	$(GOBUILD) -ldflags '$(LDFLAGS)' -o bin/alertmanager-syslog cmd/*.go

lint:
	@golangci-lint run

vet:
	$(GO) vet ./...

check: vet lint

clean:
	@rm -rf bin


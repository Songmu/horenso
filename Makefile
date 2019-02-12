VERSION = $(shell gobump show -r)
CURRENT_REVISION = $(shell git rev-parse --short HEAD)
BUILD_LDFLAGS = "-s -w -X github.com/Songmu/horenso.revision=$(CURRENT_REVISION)"
ifdef update
  u=-u
endif

GO ?= GO111MODULE=on go

devel-deps:
	GO111MODULE=off go get ${u} golang.org/x/lint/golint \
	  github.com/mattn/goveralls         \
	  github.com/motemen/gobump          \
	  github.com/Songmu/goxz/cmd/goxz    \
	  github.com/Songmu/ghch/cmd/ghch    \
	  github.com/tcnksm/ghr

test:
	$(GO) test

lint: devel-deps
	$(GO) vet
	golint -set_exit_status

cover: devel-deps
	goveralls

build:
	$(GO) build -ldflags=$(BUILD_LDFLAGS) ./cmd/horenso

bump: devel-deps
	_tools/releng

crossbuild: devel-deps
	GO111MODULE=on goxz -pv=v$(VERSION) -build-ldflags=$(BUILD_LDFLAGS) \
	  -os=linux,darwin,freebsd -arch=386,amd64 \
	  -d=./dist/v$(VERSION) ./cmd/horenso

upload:
	ghr v$(VERSION) dist/v$(VERSION)

release: bump crossbuild upload

.PHONY: deps devel-deps test lint cover build bump crossbuild upload release

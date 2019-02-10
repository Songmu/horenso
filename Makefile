VERSION = $(shell gobump show -r)
CURRENT_REVISION = $(shell git rev-parse --short HEAD)
BUILD_LDFLAGS = "-s -w -X github.com/Songmu/horenso.revision=$(CURRENT_REVISION)"
ifdef update
  u=-u
endif

deps:
	go get ${u} github.com/golang/dep/cmd/dep
	dep ensure

devel-deps: deps
	go get ${u} golang.org/x/lint/golint \
	  github.com/mattn/goveralls         \
	  github.com/motemen/gobump          \
	  github.com/Songmu/goxz/cmd/goxz    \
	  github.com/Songmu/ghch/cmd/ghch    \
	  github.com/tcnksm/ghr

test: deps
	go test

lint: devel-deps
	go vet
	golint -set_exit_status

cover: devel-deps
	goveralls

build: deps
	go build -ldflags=$(BUILD_LDFLAGS) ./cmd/horenso

bump: devel-deps
	_tools/releng

crossbuild: devel-deps
	goxz -pv=v$(VERSION) -build-ldflags=$(BUILD_LDFLAGS) \
	  -os=linux,darwin,freebsd -arch=386,amd64 \
	  -d=./dist/v$(VERSION) ./cmd/horenso

upload:
	ghr v$(VERSION) dist/v$(VERSION)

release: bump crossbuild upload

.PHONY: deps devel-deps test lint cover build bump crossbuild upload release

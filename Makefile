VERSION = $(shell godzil show-version)
CURRENT_REVISION = $(shell git rev-parse --short HEAD)
BUILD_LDFLAGS = "-s -w -X github.com/Songmu/horenso.revision=$(CURRENT_REVISION)"
ifdef update
  u=-u
endif

export GO111MODULE=on

devel-deps:
	GO111MODULE=off go get ${u}
	  golang.org/x/lint/golint            \
	  github.com/mattn/goveralls          \
	  github.com/Songmu/goxz/cmd/goxz     \
	  github.com/Songmu/godzil/cmd/godzil \
	  github.com/tcnksm/ghr

test:
	go test

lint: devel-deps
	go vet
	golint -set_exit_status

cover: devel-deps
	goveralls

build:
	go build -ldflags=$(BUILD_LDFLAGS) ./cmd/horenso

bump: devel-deps
	godzil releaes

crossbuild: devel-deps
	goxz -pv=v$(VERSION) -build-ldflags=$(BUILD_LDFLAGS) \
	  -os=linux,darwin,freebsd,windows -arch=386,amd64 \
	  -d=./dist/v$(VERSION) ./cmd/horenso

upload:
	ghr v$(VERSION) dist/v$(VERSION)

release: bump crossbuild upload

.PHONY: deps devel-deps test lint cover build bump crossbuild upload release

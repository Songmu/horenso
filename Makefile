VERSION = $(shell godzil show-version)
CURRENT_REVISION = $(shell git rev-parse --short HEAD)
BUILD_LDFLAGS = "-s -w -X github.com/Songmu/horenso.revision=$(CURRENT_REVISION)"
ifdef update
  u=-u
endif

.PHONY: deps
deps:
	go get ${u} -d
	go mod tidy

.PHONY: devel-deps
devel-deps:
	go install github.com/Songmu/godzil/cmd/godzil@latest
	go install github.com/tcnksm/ghr@latest

.PHONY: test
test:
	go test

.PHONY: build
build:
	go build -ldflags=$(BUILD_LDFLAGS) ./cmd/horenso

.PHONY: release
release: devel-deps
	godzil release

.PHONY: crossbuild
crossbuild: devel-deps
	godzil crossbuild -pv=v$(VERSION) -build-ldflags=$(BUILD_LDFLAGS) \
	  -os=linux,darwin,freebsd,windows -arch=386,amd64 \
	  -d=./dist/v$(VERSION) ./cmd/horenso

.PHONY: upload
upload:
	ghr -body="$$(./godzil changelog --latest -F markdown)" v$(VERSION) dist/v$(VERSION)

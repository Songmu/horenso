CURRENT_REVISION = $(shell git rev-parse --short HEAD)
BUILD_LDFLAGS = "-s -w -X github.com/Songmu/horenso.revision=$(CURRENT_REVISION)"
ifdef update
  u=-u
endif

.PHONY: deps
deps:
	go get ${u}
	go mod tidy

.PHONY: devel-deps
devel-deps:
	go install github.com/Songmu/gocredits/cmd/gocredits@v1.1.0

.PHONY: test
test:
	go test

.PHONY: build
build:
	go build -ldflags=$(BUILD_LDFLAGS) ./cmd/horenso

.PHONY: prepare-release
prepare-release: devel-deps
	go mod tidy
	gocredits -w
	git update-index --add --remove -- go.mod go.sum CREDITS

CREDITS: deps devel-deps go.sum
	gocredits -w

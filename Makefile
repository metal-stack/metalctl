GOOS := linux
GOARCH := amd64
CGO_ENABLED := 1
TAGS := -tags 'netgo'
BINARY := metalctl-$(GOOS)-$(GOARCH)

SHA := $(shell git rev-parse --short=8 HEAD)
GITVERSION := $(shell git describe --long --all)
# gnu date format iso-8601 is parsable with Go RFC3339
BUILDDATE := $(shell date --iso-8601=seconds)
VERSION := $(or ${VERSION},$(shell git describe --tags --exact-match 2> /dev/null || git symbolic-ref -q --short HEAD || git rev-parse --short HEAD))

ifeq ($(CGO_ENABLED),1)
ifeq ($(GOOS),linux)
	LINKMODE := -linkmode external -extldflags '-static -s -w'
	TAGS := -tags 'osusergo netgo static_build'
endif
endif

LINKMODE := $(LINKMODE) \
		 -X 'github.com/metal-stack/v.Version=$(VERSION)' \
		 -X 'github.com/metal-stack/v.Revision=$(GITVERSION)' \
		 -X 'github.com/metal-stack/v.GitSHA1=$(SHA)' \
		 -X 'github.com/metal-stack/v.BuildDate=$(BUILDDATE)'

.PHONY: all
all: build test lint-structs markdown

.PHONY: build
build:
	go build \
		$(TAGS) \
		-ldflags \
		"$(LINKMODE)" \
		-o bin/$(BINARY) \
		github.com/metal-stack/metalctl

	sha512sum bin/$(BINARY) > bin/$(BINARY).sha512

.PHONY: test
test: build
	go test -cover ./...

.PHONY: test-in-docker
test-in-docker:
	docker build -f Dockerfile.test --progress plain .

.PHONY: lint-structs
lint-structs:
	@golangci-lint run --enable exhaustruct ./cmd --tests=false || \
		echo "certain structs in metalctl should always be initialized completely! \
		 \notherwise fields would be missing in the edit command or – depending on the metal-api implementation – it could empty existing fields \
		 \n(this was probably caused by new fields in metal-go)"

.PHONY: markdown
markdown:
	rm -rf docs
	mkdir -p docs
	bin/$(BINARY) markdown

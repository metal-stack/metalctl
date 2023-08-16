BINARY := metalctl
MAINMODULE := github.com/metal-stack/metalctl
# the builder is at https://github.com/metal-stack/builder
COMMONDIR := $(or ${COMMONDIR},../builder)

-include $(COMMONDIR)/Makefile.inc

.PHONY: all
all:: markdown lint-structs

release:: all

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
	bin/metalctl markdown

.PHONY: build
build:
	$(GO) build \
		-tags netgo \
		-ldflags \
		"$(LINKMODE)" \
		-o bin/$(BINARY) \
		$(MAINMODULE)

.PHONY: lint
lint:
	docker run --rm -v $(PWD):/app -w /app golangci/golangci-lint golangci-lint run -v

.PHONY: build-platforms
build-platforms:
	docker build --no-cache -t platforms-binaries --target platforms .

.PHONY: extract-binaries
extract-binaries: build-platforms
	mkdir -p tmp
	mkdir -p result
	docker cp $(shell docker create platforms-binaries):/work/bin tmp
	mv tmp/bin/metalctl-linux-amd64 result
	mv tmp/bin/metalctl-windows-amd64 result
	mv tmp/bin/metalctl-darwin-amd64 result
	mv tmp/bin/metalctl-darwin-arm64 result
	md5sum result/metalctl-linux-amd64 > result/metalctl-linux-amd64.md5
	md5sum result/metalctl-windows-amd64 > result/metalctl-windows-amd64.md5
	md5sum result/metalctl-darwin-amd64 > result/metalctl-darwin-amd64.md5
	md5sum result/metalctl-darwin-arm64 > result/metalctl-darwin-arm64.md5
	ls -lh result

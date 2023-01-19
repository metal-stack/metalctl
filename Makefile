BINARY := metalctl
MAINMODULE := github.com/metal-stack/metalctl
COMMONDIR := $(or ${COMMONDIR},../builder)

include $(COMMONDIR)/Makefile.inc

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
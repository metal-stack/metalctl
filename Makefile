BINARY := metalctl
MAINMODULE := github.com/metal-stack/metalctl
COMMONDIR := $(or ${COMMONDIR},../builder)

include $(COMMONDIR)/Makefile.inc

.PHONY: all
all:: markdown

release:: all

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

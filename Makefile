BINARY := metalctl
MAINMODULE := github.com/metal-stack/metalctl
COMMONDIR := $(or ${COMMONDIR},../builder)

include $(COMMONDIR)/Makefile.inc

release:: all

markdown: all
	rm -rf docs; \
	mkdir -p docs ; \
	bin/metalctl markdown

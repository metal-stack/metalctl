BINARY := metalctl
MAINMODULE := git.f-i-ts.de/cloud-native/metal/metalctl
COMMONDIR := $(or ${COMMONDIR},../common)
SWAGGER_VERSION := $(or ${SWAGGER_VERSION},v0.19.0)

include $(COMMONDIR)/Makefile.inc

release:: all

markdown: all
	rm -rf docs; \
	mkdir -p docs ; \
	bin/metalctl markdown

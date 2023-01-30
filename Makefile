.PHONY: all test test-edition test-fog test-edge

uname_s := $(shell uname -s)
uname_p := $(shell uname -p)

test:
ifeq ($(uname_s),Linux)
ifeq ($(uname_p),$(filter $(uname_p), aarch64 arm))
	go test -v -count 1 $(ARGS) -tags linux,arm64,fog_edition ./test
endif
ifeq ($(uname_p),x86_64)
	go test -v -count 1 $(ARGS) -tags linux,amd64,fog_edition ./test
endif
endif
ifeq ($(uname_s),Darwin)
ifeq ($(uname_p),$(filter $(uname_p), aarch64 arm))
	go test -v -count 1 $(ARGS) -tags darwin,arm64,fog_edition ./test
endif
ifeq ($(uname_p),i386)
	go test -v -count 1 $(ARGS) -tags darwin,amd64,fog_edition ./test
endif
endif


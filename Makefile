.PHONY: all test test-edition test-fog test-edge

uname_s := $(shell uname -s)
uname_m := $(shell uname -m)

test:
ifeq ($(uname_s),Linux)
ifeq ($(uname_m),$(filter $(uname_m), aarch64 arm64))
	go test -v -count 1 $(ARGS) -tags=fog_edition ./test
	go test -v -count 1 $(ARGS) -tags=edge_edition ./test
endif
ifeq ($(uname_m),x86_64)
	go test -v -count 1 $(ARGS) -tags=fog_edition ./test
	go test -v -count 1 $(ARGS) -tags=edge_edition ./test
endif
endif
ifeq ($(uname_s),Darwin)
ifeq ($(uname_m),$(filter $(uname_m), aarch64 arm64))
	go test -v -count 1 $(ARGS) -tags=fog_edition ./test
	go test -v -count 1 $(ARGS) -tags=edge_edition ./test
endif
ifeq ($(uname_m),i386)
	go test -v -count 1 $(ARGS) -tags=fog_edition ./test
	go test -v -count 1 $(ARGS) -tags=edge_edition ./test
endif
endif


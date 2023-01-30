.PHONY: all test test-edition test-fog test-edge

uname_s := $(shell uname -s)
uname_p := $(shell uname -p)

test-fog:
	go test -v -count 1 $(ARGS) -tags fog_edition ./test

test-edge:
	go test -v -count 1 $(ARGS) -tags edge_edition ./test

test:
	make test-edge

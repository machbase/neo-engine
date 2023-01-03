.PHONY: all test

uname_s := $(shell uname -s)
uname_p := $(shell uname -p)

test:
	@go test -v -count 1 $(ARGS) ./test
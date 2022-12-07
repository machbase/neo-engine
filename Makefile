.PHONY: all

targets := $(shell ls main)

all:
	@for tg in $(targets) ; do \
		make $$tg; \
	done

cleanpackage:
	@rm -rf packages/*

test:
	@go test $(ARGS) ./server/test

test-all:
	@make -f Makefile ARGS="-cover -v -count 1" test

package:
	@./docker-package.sh machgo

package-all:
	@for tg in $(targets) ; do \
		make package-$$tg; \
	done

releases:
	@./docker-package.sh machgo linux amd64
	@./docker-package.sh machgo linux arm64/v7

package-%:
	@./scripts/package.sh $*  linux    amd64
#	@./scripts/package.sh $*  linux    arm64
#	@./scripts/package.sh $*  darwin   arm64
#	@./scripts/package.sh $*  darwin   amd64

protos := $(basename $(shell cd proto && ls *.proto))

regen-all:
	@for tg in $(protos) ; do \
		make regen-$$tg; \
	done

regen-%:
	@./regen.sh $*

%:
	@./scripts/build.sh $@

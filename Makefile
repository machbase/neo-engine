.PHONY: all

targets := $(shell ls main)

all:
	@for tg in $(targets) ; do \
		make $$tg; \
	done

cleanpackage:
	@rm -rf packages/

test:
	@go test $(ARGS) \
		mods/shqd

test-all:
	@make -f Makefile ARGS="-cover -v -count 1" test

package:
	@./docker-package.sh caud

package-all:
	@for tg in $(targets) ; do \
		make package-$$tg; \
	done

docker-package:
	@./docker-package.sh machgo linux arm64
	@./docker-package.sh machgo linux amd64

package-%:
	@./scripts/package.sh $*  linux    amd64
#	@./scripts/package.sh $*  linux    arm64

%:
	@./scripts/build.sh $@

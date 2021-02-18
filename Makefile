# This Makefile is intended only to facilitate muscle memory for things like
# `make test`, `make lint`, and other common operations so the developer doesn't
# have to do `./scripts/lint.sh.`  It is not expected that we will start using
# complicated Make logic.

# TODO: make help smarter
help:
	@echo Known targets:
	@egrep '^[a-z]+:' Makefile | cut -d: -f1

lint:
	docker run --rm -v $(PWD):/app -w /app \
      golangci/golangci-lint:v1.35.2 golangci-lint run \
	  -E golint,goimports,misspell

test:
	go test -cover -v ./...

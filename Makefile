VERSION ?= $(shell git describe --abbrev=4 --dirty --always --tags)
TIME := $(shell date '+%Y-%m-%d_%H:%M:%S')
UPSTREAM_BRANCH ?= origin/master

PHONY: lint
lint:
	go run github.com/golangci/golangci-lint/cmd/golangci-lint@latest run
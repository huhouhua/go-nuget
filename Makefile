# Copyright (c) 2025 Kevin Berger <huhouhuam@outlook.com>. All rights reserved.
# Use of this source code is governed by a MIT style
# license that can be found in the LICENSE file.

# Build all by default, even if it's not first
.DEFAULT_GOAL := all

.PHONY: all
all: tools tidy verify-copyright format lint cover

SHELL := /bin/bash
GO := go
ROOT_DIR=.

ROOT_PACKAGE=github.com/huhouhua

ifeq ($(origin OUTPUT_DIR),undefined)
OUTPUT_DIR := $(ROOT_DIR)/_output
$(shell mkdir -p $(OUTPUT_DIR))
endif

# Linux command settings
FIND := find .
XARGS := xargs -r

# Minimum test coverage
ifeq ($(origin COVERAGE),undefined)
# prod
#COVERAGE := 60
# develop
COVERAGE := 0.0
endif

include scripts/tools.mk

## lint: Check syntax and styling of go sources.
.PHONY: lint
lint: tools.verify.golangci-lint
	@echo "===========> Run golangci to lint source codes"
	@golangci-lint run -c $(ROOT_DIR)/.golangci.yml $(ROOT_DIR)/...

## test: Run unit test.
.PHONY: test
test: tools.verify.go-junit-report
	@echo "===========> Run unit test"
	@set -o pipefail;$(GO) test ./ -race -cover -coverprofile=$(OUTPUT_DIR)/coverage.out \
		-timeout=10m -shuffle=on -short \

	@sed -i '/mock_.*.go/d' $(OUTPUT_DIR)/coverage.out # remove mock_.*.go files from test coverage
	@$(GO) tool cover -html=$(OUTPUT_DIR)/coverage.out -o $(OUTPUT_DIR)/coverage.html

## cover: Run unit test and get test coverage.
.PHONY: cover
cover: test
	@$(GO) tool cover -func=$(OUTPUT_DIR)/coverage.out | \
		awk -v target=$(COVERAGE) -f $(ROOT_DIR)/scripts/coverage.awk

## format: Gofmt (reformat) package sources (exclude vendor dir if existed).
.PHONY: format
format: tools.verify.golines tools.verify.goimports
	@echo "===========> Formating codes"
	@$(FIND) -type f -name '*.go' | $(XARGS) gofmt -s -w
	@$(FIND) -type f -name '*.go' | $(XARGS) goimports -w -local $(ROOT_PACKAGE)
	@$(FIND) -type f -name '*.go' | $(XARGS) golines -w --max-len=120 --reformat-tags --shorten-comments --ignore-generated .
	@$(GO) mod edit -fmt

## verify-copyright: Verify the boilerplate headers for all files.
.PHONY: verify-copyright
verify-copyright: tools.verify.licctl
	@echo "===========> Verifying the boilerplate headers for all files"
	@licctl --check -f $(ROOT_DIR)/scripts/boilerplate.txt $(ROOT_DIR) --skip-dirs=_output

## add-copyright: Ensures source code files have copyright license headers.
.PHONY: add-copyright
add-copyright: tools.verify.licctl
	@licctl -v -f $(ROOT_DIR)/scripts/boilerplate.txt $(ROOT_DIR) --skip-dirs=_output

## tools: install dependent tools.
.PHONY: tools
tools:
	@$(MAKE) tools.install

.PHONY: tidy
tidy:
	@$(GO) mod tidy

.PHONY: help
help: ## Display this help
	@awk 'BEGIN {FS = ":.*##"; printf "\nUsage:\n  make \033[36m<target>\033[0m\n"} /^[a-zA-Z_0-9-]+:.*?##/ { printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2 } /^##@/ { printf "\n\033[1m%s\033[0m\n", substr($$0, 5) } ' $(MAKEFILE_LIST)

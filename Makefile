# Copyright (c) 2025 Kevin Berger <huhouhuam@gmail.com>. All rights reserved.
# Use of this source code is governed by a MIT style
# license that can be found in the LICENSE file.

# Build all by default, even if it's not first
.DEFAULT_GOAL := all

.PHONY: all
all: tools tidy verify-copyright format lint cover

# ==============================================================================
# Build options

SHELL := /bin/bash
GO := go
GOLANG_CI_LINT_VERSION ?= 2.9.0
ROOT_PACKAGE=github.com/huhouhua/go-nuget
COMMON_SELF_DIR := $(dir $(lastword $(MAKEFILE_LIST)))

ifeq ($(origin ROOT_DIR),undefined)
ROOT_DIR := $(abspath $(shell cd $(COMMON_SELF_DIR) && pwd -P))
endif

ifeq ($(origin OUTPUT_DIR),undefined)
OUTPUT_DIR := $(ROOT_DIR)/_output
$(shell mkdir -p $(OUTPUT_DIR))
endif

ifeq ($(origin BIN_DIR),undefined)
BIN_DIR := $(ROOT_DIR)/bin
$(shell mkdir -p $(BIN_DIR))
endif

# Linux command settings
FIND := find .
XARGS := xargs -r

# Minimum test coverage
ifeq ($(origin COVERAGE),undefined)
COVERAGE := 60
endif

# Copy githook scripts when execute makefile
COPY_GITHOOK:=$(shell cp -f $(ROOT_DIR)/githooks/* $(ROOT_DIR)/.git/hooks/)

# ==============================================================================
# Includes

include scripts/Makefile.tools.mk

# ==============================================================================
# Targets

## lint: Check syntax and styling of go sources.
.PHONY: lint
lint: tools.verify.local.golangci-lint
	@echo "===========> Run golangci to lint source codes"
	@$(BIN_DIR)/golangci-lint run -c $(ROOT_DIR)/.golangci.yml $(ROOT_DIR)/...

## test: Run unit test.
.PHONY: test
test: tools.verify.go-junit-report
	@echo "===========> Run unit test"
	@set -o pipefail;$(GO) test -tags=test $(shell go list ./... | grep -v '^$(ROOT_PACKAGE)/e2e' | grep -v '^$(ROOT_PACKAGE)/examples') -race -cover -coverprofile=$(OUTPUT_DIR)/coverage.out \
		-timeout=10m -shuffle=on -short \
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
	@licctl --check -f $(ROOT_DIR)/scripts/boilerplate.txt $(ROOT_DIR) --skip-dirs=_output,testdata,.github

## add-copyright: Ensures source code files have copyright license headers.
.PHONY: add-copyright
add-copyright: tools.verify.licctl
	@licctl -v -f $(ROOT_DIR)/scripts/boilerplate.txt $(ROOT_DIR) --skip-dirs=_output,testdata,.github,.idea

## tools: Install dependent tools.
.PHONY: tools
tools:
	@$(MAKE) tools.install

## check-updates: Check outdated dependencies of the go projects.
.PHONY: check-updates
check-updates: tools.verify.go-mod-outdated
	@$(GO) list -u -m -json all | go-mod-outdated -update -direct

## tidy: Clean up go.mod and go.sum by removing unused dependencies and adding missing ones
.PHONY: tidy
tidy:
	@$(GO) mod tidy

## help: Show this help info.
.PHONY: help
help: Makefile
	@printf "\nUsage: make <TARGETS> <OPTIONS> ...\n\nTargets:\n"
	@sed -n 's/^##//p' $< | column -t -s ':' | sed -e 's/^/ /'
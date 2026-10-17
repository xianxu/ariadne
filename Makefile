# Canonical repo name from git remote (portable across worktrees and containers)
REPO_NAME := $(shell git remote get-url origin 2>/dev/null | sed 's|.*/||; s|\.git$$||')

# Share compiled packages across worktrees while retaining configured Go flags.
GOFLAGS := $(shell go env GOFLAGS)
ifeq ($(filter -trimpath,$(GOFLAGS)),)
GOFLAGS += -trimpath
endif
export GOFLAGS

# This project nests issues and history under workshop/
WF_ISSUES_DIR = workshop/issues
WF_HISTORY_DIR = workshop/history

# The sharded suite (#253, ~12 processes) takes both full-suite slots, so two
# ariadne suites never run at once (#346).
WF_TEST_WEIGHT ?= 2

# Owner build declarations are available before generated workflow helpers.
.DEFAULT_GOAL := help
-include Makefile.workflow
-include Makefile.local

.PHONY: help
help: $(WF_HELP_TARGETS)
	@true

# Build this layer's explicit tools from tracked sources. weave itself is a
# distributed gateway, built separately for development/release.
.PHONY: tools sdlc-build datatype-build vocabulary-build doc-review-build tl-worker-build tl-lead-build weave-build vocab-embed
tools: sdlc-build datatype-build vocabulary-build doc-review-build tl-worker-build tl-lead-build

# Each publishes with one rename (WF_GO_PUBLISH, #419): agents run :0's bin/
# while the TL rebuilds it.
sdlc-build datatype-build vocabulary-build doc-review-build tl-worker-build tl-lead-build:
	@mkdir -p bin
	@echo "go build ./cmd/$(@:-build=) → bin/$(@:-build=)"
	@$(WF_GO_PUBLISH) bin/$(@:-build=) ./cmd/$(@:-build=)

weave-build:
	@mkdir -p bin
	@echo "go build ./cmd/weave → bin/weave"
	@$(WF_GO_PUBLISH) bin/weave ./cmd/weave

# Explicit regeneration guard; never a prerequisite of owner tools.
vocab-embed: vocabulary-build
	PATH="$(CURDIR)/bin:$$PATH" go generate ./pkg/vocab/...
	@git diff --exit-code -- pkg/vocab

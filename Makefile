# Canonical repo name from git remote (portable across worktrees and containers)
REPO_NAME := $(shell git remote get-url origin 2>/dev/null | sed 's|.*/||; s|\.git$$||')

# This project nests issues and history under workshop/
WF_ISSUES_DIR = workshop/issues
WF_HISTORY_DIR = workshop/history

# Owner build declarations are available before generated workflow helpers.
.DEFAULT_GOAL := help
-include Makefile.workflow
-include Makefile.local

.PHONY: help
help: $(WF_HELP_TARGETS)
	@true

# Build this layer's explicit tools from tracked sources. weave itself is a
# distributed gateway, built separately for development/release.
.PHONY: tools sdlc-build datatype-build vocabulary-build doc-review-build weave-build vocab-embed
tools: sdlc-build datatype-build vocabulary-build doc-review-build

sdlc-build datatype-build vocabulary-build doc-review-build:
	@mkdir -p bin
	go build -o bin/$(@:-build=) ./cmd/$(@:-build=)

weave-build:
	@mkdir -p bin
	go build -o bin/weave ./cmd/weave

# Explicit regeneration guard; never a prerequisite of owner tools.
vocab-embed: vocabulary-build
	PATH="$(CURDIR)/bin:$$PATH" go generate ./pkg/vocab/...
	@git diff --exit-code -- pkg/vocab

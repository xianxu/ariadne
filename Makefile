# Canonical repo name from git remote (portable across worktrees and containers)
REPO_NAME := $(shell git remote get-url origin 2>/dev/null | sed 's|.*/||; s|\.git$$||')

# This project nests issues and history under workshop/
WF_ISSUES_DIR = workshop/issues
WF_HISTORY_DIR = workshop/history

# Assemble sub-Makefiles
include Makefile.workflow
# -include .openshell/Makefile  # Layer 2: sandbox (coming later)
# -include Makefile.construct   # Layer 3: construct management (coming later)

.PHONY: help

help: help-workflow
	@true

# Canonical repo name from git remote (portable across worktrees and containers)
REPO_NAME := $(shell git remote get-url origin 2>/dev/null | sed 's|.*/||; s|\.git$$||')

# This project nests issues and history under workshop/
WF_ISSUES_DIR = workshop/issues
WF_HISTORY_DIR = workshop/history

# Assemble sub-Makefiles (Makefile.workflow already includes .openshell/Makefile)
include Makefile.workflow
-include Makefile.local

.PHONY: help

help: help-workflow help-sandbox
	@true

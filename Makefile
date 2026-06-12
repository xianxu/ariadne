# Canonical repo name from git remote (portable across worktrees and containers)
REPO_NAME := $(shell git remote get-url origin 2>/dev/null | sed 's|.*/||; s|\.git$$||')

# This project nests issues and history under workshop/
WF_ISSUES_DIR = workshop/issues
WF_HISTORY_DIR = workshop/history

# Assemble sub-Makefiles (Makefile.workflow already includes .openshell/Makefile)
include Makefile.workflow
-include Makefile.local

.PHONY: help

# help-sandbox, help-tart, and help-colima are defined by .openshell/Makefile,
# .tart/Makefile, and .colima/Makefile respectively, all included via
# Makefile.workflow's -include lines. Every consumer that vendors the ariadne
# base layer ships all three fragments (see construct/base.manifest), so these
# targets always resolve. If a consumer ever drops .openshell, .tart, or
# .colima from its manifest, the corresponding help-X line would need to come
# out. Transient window: a consumer that pulls this updated Makefile before
# running setup.sh to materialize the new .colima/Makefile symlink will get a
# "No rule to make target 'help-colima'" until setup runs — same accepted
# fragility the other two already carry.
help: help-workflow help-sandbox help-tart help-colima
	@true

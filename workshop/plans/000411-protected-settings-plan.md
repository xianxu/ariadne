# Protected settings at planning gates — implementation plan

> **For agentic workers:** Consult AGENTS.md Section 3 (Subagent Strategy). Execute bounded ownership work with a subagent; main handles the warm-context gate wiring. Steps use checkboxes.

**Goal:** A protected Claude settings file cannot fail planning, and stale settings remain visible until an operator refreshes them.

**Architecture:** Choose the Spec's operator-run-publication alternative. Gate refresh explicitly defers Claude settings; ordinary shell refresh --substrate publishes them. A read-only settings check runs even when Git is already current. Preserve the independent rebuilt-binary rerun after Git HEAD advancement.

**Tech Stack:** Go, Cobra, existing weavefs/settingsx/managed ownership, local Git fixtures and native sandbox-exec conformance.

## Core concepts

| Pure entity | Lives in | Status |
|---|---|---|
| DeferredOutput action | cmd/weave/internal/plan/action.go | new |
| Managed output ownership | cmd/weave/internal/plan/ownership.go | modified |

DeferredOutput declares one produced path whose publication belongs to the operator. Its previous inventory entries remain unchanged; it cannot acquire desired-output ownership. Existing matching/retirement/displaced-copy rules remain authoritative (ARCH-DRY).

| Integration | Lives in | Status | Wraps |
|---|---|---|---|
| Settings check/defer | cmd/weave/settings.go | new | existing walk/planner, settingsx merge/compare, weavefs reads |
| Gate refresh mode | cmd/weave/refresh.go | modified | existing refresh/compile transaction |
| Planning settings observation | cmd/sdlc/catchup.go | modified | installed weave process |

ARCH-PURE: reuse pure settingsx merge/equality; keep reads/output in the IO shell. No permission inference by environment or file mode; a gate never attempts the protected write. Existing real Git fixtures and filesystem doubles are used; native deny controls qualify the production OS seam (ARCH-MOCK).

## State and limits

- Current settings: check emits no stale notice. Missing/different: emit exactly one line naming .claude/settings.json and run weave refresh --substrate from your shell. Operator refresh publishes, next check clears notice.
- Deferred ownership: desired source changes do not authorize settings displacement, publication or retirement; preserve prior records/backups. Ordinary apply later converges and retires by existing rules (ARCH-ORDER/FUNERAL).
- Git catchup stays separate: changed HEAD still requires rerun with rebuilt sdlc. Settings never adds a failed gate or unsandboxed retry.
- Read-only check starts no acquisition, dependency install, tools build or generators. It walks only declared present layers and reads existing manifest/settings inputs. Use existing process cancellation limits; no new persistent artifact or growing ledger.
- Trust: source/local JSON use existing parser; invalid/unreadable inputs remain visible errors, never inferred current. No secrets are created or printed; notice contains only path/action (ARCH-SECURE).

## Tasks

- [x] Reproduce native denial before code; record first/second-run root cause and preserve #303 rerun safety.
- [x] Test prepareManaged/commitManaged through ApplyManaged with repeated/interrupted ownership inputs; assert identity and recovery evidence remains unchanged until actual publication or retirement.
- [x] Test deferredClaudeSettings and buildSettings against adversarial JSON/filesystem states; observe exact notice counts and deny all protected mutations. Add read-only observation and explicit deferred-refresh mode.
- [x] Test catchUpBeforeWork/startPlanBranch with stateful installed-weave fixtures; assert settings observation cannot fail planning or obscure genuine catchup errors. Wire gate deferral and retain the moved-HEAD rerun.
- [x] Run focused weave and SDLC catchup tests, negative controls and native denial conformance.
- [x] Update README and atlas; push and record the full suite.

After all verification completes, close through SDLC, land on SHIP and report hub.

## Verification

RED already observed on installed weave: native deny file-write* for exact .claude/settings.json makes compile exit 1 at rename with operation not permitted; control touch is denied, old settings retained. Temporary fixture was removed.
Focused commands: go test -trimpath ./cmd/weave/... ; go test -trimpath ./cmd/sdlc -run 'TestStartPlan.*|TestGateSettings.*' -count=1 -timeout 10m. Native conformance uses a scratch-only sandbox with an explicit denied-write control. Full suite via sdlc local-check run --name suite on pushed head. No unsandboxed retry of a failed settings write.

## Revisions

### 2026-10-16 — PQ-1/PQ-2

The single recovery action is run weave refresh --substrate from your shell: default refresh refuses normal ahead/divergent issue branches before compile (refresh.go), whereas --substrate preserves the consumer branch. Verify clearing through this real command on an issue branch, not by hand-copying expected bytes. Function-level adversarial strategies replace lossy scenario lists; tests own the detailed enumeration.

### 2026-10-16 — verification bookkeeping

Separate the final validation checkbox from its following close/landing actions; no change to deliverables.

### 2026-10-16 — recorded execution versus passing landing validation

Two full suite runs completed with unrelated failures (shell timeout then reproduced main flake #413); all feature tests pass. Record the execution step complete and submit those exact outcomes to mandatory close review while an unchanged retry queues. Passing local-suite evidence remains required at merge.

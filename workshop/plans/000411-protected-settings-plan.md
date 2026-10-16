# Protected settings at planning gates — implementation plan

> **For agentic workers:** Consult AGENTS.md Section 3 (Subagent Strategy). Execute bounded ownership work with a subagent; main handles the warm-context gate wiring. Steps use checkboxes.

**Goal:** A protected Claude settings file cannot fail planning, and stale settings remain visible until an operator refreshes them.

**Architecture:** Choose the Spec's operator-run-publication alternative. Gate refresh explicitly defers Claude settings; ordinary shell refresh publishes them. A read-only settings check runs even when Git is already current. Preserve the independent rebuilt-binary rerun after Git HEAD advancement.

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

- Current settings: check emits no stale notice. Missing/different: emit exactly one line naming .claude/settings.json and run weave refresh from your shell. Operator refresh publishes, next check clears notice.
- Deferred ownership: desired source changes do not authorize settings displacement, publication or retirement; preserve prior records/backups. Ordinary apply later converges and retires by existing rules (ARCH-ORDER/FUNERAL).
- Git catchup stays separate: changed HEAD still requires rerun with rebuilt sdlc. Settings never adds a failed gate or unsandboxed retry.
- Read-only check starts no acquisition, dependency install, tools build or generators. It walks only declared present layers and reads existing manifest/settings inputs. Use existing process cancellation limits; no new persistent artifact or growing ledger.
- Trust: source/local JSON use existing parser; invalid/unreadable inputs remain visible errors, never inferred current. No secrets are created or printed; notice contains only path/action (ARCH-SECURE).

## Tasks

- [ ] Reproduce native denial before code; record first/second-run root cause and preserve #303 rerun safety.
- [ ] Add DeferredOutput tests, watch red, implement retention through managed apply, verify recovery/retirement/backups.
- [ ] Add read-only settings command and explicit deferred-refresh option; tests stale/current/missing and no protected effects.
- [ ] Wire start-plan to read-only observation when current, gate refresh to defer settings; test first successful planning, recurring notice, operator correction clearing notice, and unchanged moved-HEAD rerun.
- [ ] Run focused weave and SDLC catchup tests, mutation/negative controls and native denial conformance.
- [ ] Update README and atlas; push, record full suite, close review, land on SHIP and report hub.

## Verification

RED already observed on installed weave: native deny file-write* for exact .claude/settings.json makes compile exit 1 at rename with operation not permitted; control touch is denied, old settings retained. Temporary fixture was removed.
Focused commands: go test -trimpath ./cmd/weave/... ; go test -trimpath ./cmd/sdlc -run 'TestStartPlan.*|TestGateSettings.*' -count=1 -timeout 10m. Native conformance uses a scratch-only sandbox with an explicit denied-write control. Full suite via sdlc local-check run --name suite on pushed head. No unsandboxed retry of a failed settings write.

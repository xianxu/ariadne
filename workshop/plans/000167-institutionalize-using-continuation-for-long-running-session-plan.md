# Proactive Session Continuity Policy Implementation Plan

> **For agentic workers:** Consult AGENTS.md Section 3 (Subagent Strategy) to determine the appropriate execution approach: use superpowers-subagent-driven-development (if subagents are suitable per AGENTS.md) or superpowers-executing-plans to implement this plan. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Make every Ariadne agent proactively checkpoint a long-running session through the canonical continuation flow when active context exceeds 60%.

**Architecture:** Add one declarative trigger policy to the exported base constitution. Route from that policy to the existing continuation datatype for handoff content and to the available writer for persistence/restart; do not duplicate either downstream contract. Guard the ownership boundaries with a focused semantic test and map the lifecycle in the atlas.

**Tech Stack:** Markdown constitution and atlas; Go standard-library contract test; existing `weave` base-layer composition.

---

## Core concepts

### Pure entities

| Name | Lives in | Status |
|------|----------|--------|
| `SessionContinuityPolicy` | `AGENTS.base.md` | new |

- **SessionContinuityPolicy** — declarative decision rule mapping observable context pressure to a continuation checkpoint.
  - **Relationships:** N:1 from every composed harness entry file to the one exported base policy; 1:1 route from the policy to the canonical continuation datatype procedure.
  - **DRY rationale:** `AGENTS.base.md` is the one always-loaded source for **when** to checkpoint; `construct/datatype/continuation.md` remains the one source for **how** to checkpoint (`ARCH-DRY`).
  - **Future extensions:** The trigger can gain additional harness-neutral pressure signals, but model-window catalogs and polling remain outside this policy.

The policy is prose rather than executable Go, but its observable contract is pure: given either utilization above 60% or a context-pressure/compaction warning when percentage is unavailable, it directs the agent to checkpoint before another substantial unit of work. `cmd/datatype/continuation_policy_test.go` guards those stable semantics without pinning the full wording.

### Integration points

| Name | Lives in | Status | Wraps |
|------|----------|--------|-------|
| `BaseConstitutionExport` | `AGENTS.base.md` | modified | existing `construct/base.manifest` prose fan-out to Claude, Codex, and Gemini entry files |
| `ContinuationWriterBoundary` | `AGENTS.base.md` | new | producer-provided continuation writer and its restart behavior |

- **BaseConstitutionExport** — delivers the policy to every supported agent through the existing weave composition path.
  - **Injected into:** Harness entry files generated from the base prose; no harness-specific copy is added.
  - **Future extensions:** New harness faces inherit the rule by joining the existing prose fan-out.
- **ContinuationWriterBoundary** — tells the agent to finalize through the available writer and not perform a second restart.
  - **Injected into:** The agent's checkpoint procedure after it selects the continuation datatype. If no writer exists, the datatype's existing stop behavior remains authoritative.
  - **Future extensions:** Additional producers may supply their own writer/restart implementation without changing the constitution.

## Chunk 1: Constitution contract

### Task 1: Add the failing session-continuity policy guard

**Files:**
- Create: `cmd/datatype/continuation_policy_test.go`
- Read: `AGENTS.base.md`
- Read: `construct/datatype/continuation.md`

- [x] **Step 1: Write the failing semantic contract test**

Create `cmd/datatype/continuation_policy_test.go`:

```go
package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestSessionContinuityPolicyRoutesToDatatypeAndWriter(t *testing.T) {
	root := filepath.Join("..", "..")
	raw, err := os.ReadFile(filepath.Join(root, "AGENTS.base.md"))
	if err != nil {
		t.Fatal(err)
	}
	agents := string(raw)
	start := strings.Index(agents, "### 14. Session Continuity")
	if start < 0 {
		t.Fatal("AGENTS.base.md missing Session Continuity policy")
	}
	policy := agents[start:]
	if end := strings.Index(policy, "\n## Core Design Principles"); end >= 0 {
		policy = policy[:end]
	}

	for name, marker := range map[string]string{
		"threshold":        "60%",
		"fallback signal":  "compaction warning",
		"canonical route":  "`continuation` datatype",
		"writer boundary":  "writer owns the restart",
		"no-writer behavior": "no writer",
	} {
		if !strings.Contains(policy, marker) {
			t.Errorf("Session Continuity policy missing %s marker %q", name, marker)
		}
	}

	proto, err := os.ReadFile(filepath.Join(root, "construct", "datatype", "continuation.md"))
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(proto), "60%") {
		t.Error("60% trigger belongs only in AGENTS.base.md, not the continuation procedure")
	}
}
```

- [x] **Step 2: Run the focused test and verify it fails**

Run: `go test ./cmd/datatype -run TestSessionContinuityPolicyRoutesToDatatypeAndWriter -count=1`

Expected: FAIL with `AGENTS.base.md missing Session Continuity policy`.

### Task 2: Add the exported constitution policy

**Files:**
- Modify: `AGENTS.base.md`
- Test: `cmd/datatype/continuation_policy_test.go`

- [x] **Step 1: Add the policy after “Model User Intention”**

Add this section before `## Core Design Principles`:

```markdown
### 14. Session Continuity
- When the harness reports that the active context is more than 60% full, proactively checkpoint the session before starting another substantial unit of work. Finish the current atomic action and update its durable issue/plan/log state first. If an exact percentage is unavailable, treat a harness context-pressure or compaction warning as the trigger.
- Apply the canonical **`continuation` datatype** for the checkpoint; it owns what to preserve and how to finalize the handoff. Use the available continuation writer. The writer owns the restart after a successful durable write, so don't separately restart the session. If no writer is available, follow the datatype's existing no-writer stop behavior.
```

This is intentionally a route, not a restatement of the continuation skeleton (`ARCH-DRY`). It is harness-neutral and does not name Pair commands (`ARCH-PURPOSE`). Restart remains an integration effect owned by the writer (`ARCH-PURE`).

- [x] **Step 2: Run the focused test and verify it passes**

Run: `go test ./cmd/datatype -run TestSessionContinuityPolicyRoutesToDatatypeAndWriter -count=1`

Expected: PASS.

- [x] **Step 3: Run the datatype package tests**

Run: `go test ./cmd/datatype -count=1`

Expected: PASS.

- [x] **Step 4: Commit the constitution contract**

```bash
git add AGENTS.base.md cmd/datatype/continuation_policy_test.go
git commit -m "#167: institutionalize proactive session continuity" \
  -m "Route context pressure through the canonical continuation datatype while leaving durable restart mechanics with the writer." \
  -m "Co-Authored-By: OpenAI Codex <noreply@openai.com>"
```

## Chunk 2: Lifecycle map and verification

### Task 3: Map the proactive lifecycle without duplicating procedure

**Files:**
- Modify: `atlas/workflow/data-artifacts.md`
- Verify: `atlas/index.md`

- [x] **Step 1: Extend the continuation lifecycle paragraph**

After the existing continuation paragraph, add a concise mapping such as:

```markdown
`AGENTS.base.md`'s **Session Continuity** policy is the canonical proactive trigger: under material context pressure, the agent finishes its current atomic action, updates durable work state, and routes into the continuation datatype before beginning another substantial unit. The datatype remains the procedure source; a producer-provided writer owns persistence and any automatic restart. This keeps the trigger visible to every harness without copying the continuation skeleton or Pair-specific mechanics into the constitution.
```

Do not repeat the section skeleton or writer command. Confirm `atlas/index.md` already links `workflow/data-artifacts.md`; no index edit is expected.

- [x] **Step 2: Verify the atlas pointer exists**

Run: `rg -n 'workflow/data-artifacts\.md' atlas/index.md`

Expected: at least one matching link.

- [x] **Step 3: Run the base-layer composition tests**

Run: `go test ./cmd/weave/... -count=1`

Expected: PASS, proving exported prose still fans out to the supported harness entry files.

- [x] **Step 4: Run repository verification**

Run: `go test ./... -count=1`

Expected: PASS.

Run: `git diff --check`

Expected: no output and exit 0.

- [x] **Step 5: Commit the lifecycle map**

```bash
git add atlas/workflow/data-artifacts.md
git commit -m "#167: map proactive continuation lifecycle" \
  -m "Document the constitution-to-datatype-to-writer ownership chain without duplicating the procedure." \
  -m "Co-Authored-By: OpenAI Codex <noreply@openai.com>"
```

- [x] **Step 6: Inspect the committed diff against the issue purpose**

Run: `git diff main...HEAD -- AGENTS.base.md cmd/datatype/continuation_policy_test.go atlas/workflow/data-artifacts.md`

Expected: one trigger policy, one semantic guard, and one atlas mapping.

Run: `git diff --name-only main...HEAD`

Expected: exactly `AGENTS.base.md`, `atlas/workflow/data-artifacts.md`, and `cmd/datatype/continuation_policy_test.go`; no datatype-prototype, Pair, or harness-specific generated entry-file changes.

### Task 4: Reconcile execution records and close through SDLC

**Files:**
- Modify: `workshop/issues/000167-institutionalize-using-continuation-for-long-running-session.md`
- Modify: `workshop/plans/000167-institutionalize-using-continuation-for-long-running-session-plan.md`

- [x] **Step 1: Tick the completed implementation checkboxes and record verification**

Update the issue and durable plan so every completed implementation/verification step is checked. Append the verification outcome to the issue's `## Log`; do not rewrite prior log entries. Before the record commit below, mark Steps 1–3 of this task complete as part of the same reconciliation change. The SDLC close itself is deliberately outside the checkbox list so the close gate never depends on a self-referential unchecked “close” step.

- [x] **Step 2: Validate records and working tree**

Run: `sdlc issue validate workshop/issues/000167-institutionalize-using-continuation-for-long-running-session.md`

Expected: `[ok] ... conforms`.

Run: `git diff --check`

Expected: no output and exit 0.

- [x] **Step 3: Commit the reconciled records**

```bash
git add workshop/issues/000167-institutionalize-using-continuation-for-long-running-session.md workshop/plans/000167-institutionalize-using-continuation-for-long-running-session-plan.md
git commit -m "#167: record session continuity verification" \
  -m "Keep the issue and durable plan aligned with the verified implementation before the close boundary." \
  -m "Co-Authored-By: OpenAI Codex <noreply@openai.com>"
```

### Close the single atomic review boundary

After every checkbox above is committed, run `sdlc actual --issue 167`, inspect the measured value, and substitute it for `<measured-hours>` below:

```bash
sdlc close --issue 167 --actual <measured-hours> --verified 'Focused datatype and weave tests plus go test ./... passed; git diff --check clean; scoped diff confirms AGENTS.base is the sole trigger policy, continuation procedure unchanged, and atlas current.'
```

Expected: the deterministic gates pass, the binary-dispatched fresh-context review returns an acceptable verdict, and the issue closes. Use the precise `--no-atlas`/other bypass only if a named gate genuinely does not apply and record why in `--verified`; do not use `--force` for routine completion.

After a successful close, commit the close-generated issue state (the durable plan is already fully checked and committed):

```bash
git add workshop/issues/000167-institutionalize-using-continuation-for-long-running-session.md
git commit -m "#167: close proactive session continuity" \
  -m "Record the measured actual and successful boundary review." \
  -m "Co-Authored-By: OpenAI Codex <noreply@openai.com>"
```

## Revisions

### 2026-07-12T22:31:00-07:00 — align the durable-plan filename with SDLC discovery

`sdlc change-code` resolves a separate plan as
`workshop/plans/<issue-basename>-plan.md`. The original shorter filename was not
delivered to the plan-quality reviewer, so the plan moved to the issue-basename
form and its self-references were updated. Scope and implementation steps are
unchanged.

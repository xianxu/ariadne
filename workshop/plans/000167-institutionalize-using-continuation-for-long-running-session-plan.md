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

### 2026-07-12T22:44:00-07:00 — reclassify the policy contract and test real fan-out

The first whole-issue close review returned REWORK. The original **Pure
entities** section is superseded: `SessionContinuityPolicy` is a declarative
INTEGRATION contract consumed by agent harnesses, not a PURE code entity. Its
repository contract test necessarily reads the exported prose and composition
manifest. `BaseConstitutionExport` is also existing/unmodified infrastructure in
`construct/base.manifest` and `cmd/weave/internal/plan`; this issue changes only
its `AGENTS.base.md` input.

Revised integration concepts:

| Name | Lives in | Status | Wraps |
|------|----------|--------|-------|
| `SessionContinuityPolicy` | `AGENTS.base.md` | new | agent behavior under context pressure |
| `ContinuationWriterBoundary` | `AGENTS.base.md` | new | producer-provided continuation writer and restart behavior |

The existing `BaseConstitutionExport` composes the policy into `CLAUDE.md`,
`AGENTS.md`, and `GEMINI.md`; it remains unmodified and is exercised rather than
reimplemented (`ARCH-DRY`, `ARCH-PURPOSE`). No new PURE entity is introduced.

#### Revision task 1: Replace the marker-only test with a weave integration guard

**Files:**
- Delete: `cmd/datatype/continuation_policy_test.go`
- Create: `cmd/weave/session_continuity_test.go`

- [x] Assert the exact trigger semantics (`more than 60% full`, checkpoint
  before another substantial unit), both fallback signals, datatype routing,
  no-double-restart wording, and existing no-writer behavior.
- [x] Assert the live `construct/base.manifest` exports `AGENTS.base.md`, compile
  the real base fragment through a minimal end-to-end weave fixture, and verify
  the distinctive policy appears in all three harness entry files.
- [x] Prove the regression guard by mutating the threshold direction and export
  visibility one at a time, observing focused failures, then restoring both and
  observing a pass.

#### Revision task 2: Record the review lesson and reverify

**Files:**
- Modify: `workshop/lessons.md`
- Modify: `workshop/issues/000167-institutionalize-using-continuation-for-long-running-session.md`
- Modify: this plan

- [x] Record that plan entity classification must match the test boundary, and
  that prose contract tests pin semantics plus derived consumers rather than
  token presence alone.
- [x] Run `go test ./cmd/weave -run TestSessionContinuityPolicyFansOutToEveryHarness -count=1`,
  `go test ./cmd/weave/... -count=1`, `go test ./... -count=1`, issue validation,
  and `git diff --check`.
- [x] Reconcile all issue/plan checkboxes, commit the remediation with the prior
  `Review-Verdict: REWORK` and `Review-Window: 6eeb64d..792fc3f` trailers, then
  re-run `sdlc close --issue 167 --actual 1.05 ...`.

### 2026-07-12T22:52:00-07:00 — harden policy-section and active-export assertions

The second whole-issue review returned FIX-THEN-SHIP. Behavior and architecture
passed, but the integration guard could accept (1) a policy marker moved outside
the Session Continuity section and (2) a commented-out base-manifest export.
The test-hardening delta is:

- [x] Slice `AGENTS.base.md` from `### 14. Session Continuity` through
  `## Core Design Principles` before checking every policy semantic marker.
- [x] Parse the live manifest with `intent.ParseManifest` and require an active
  `intent.Prose` / `intent.Export` entry whose source is `AGENTS.base.md`; use
  that validated row to build the synthetic fixture.
- [x] Prove the guards with a moved-marker mutant and a commented-export mutant,
  restore both, then re-run the focused, weave, and full repository suites.
- [x] Reconcile records and commit with `Review-Verdict: FIX-THEN-SHIP` and
  `Review-Window: 6eeb64d..5d112ae` before the final close rerun.

### 2026-07-12T23:07:00-07:00 — derive the consumer sweep from the harness registry

The next whole-issue review confirmed the policy and prior hardening, but found
that the test's literal `CLAUDE.md` / `AGENTS.md` / `GEMINI.md` list duplicated
the canonical registry. This cheap `ARCH-DRY` / `ARCH-PURPOSE` remediation keeps
the “every harness” promise true as the registry evolves:

- [x] Replace the literal consumer slice with `plan.TargetAll.EntryFiles()`.
- [x] Re-run the focused guard, full repository suite, issue validation, and
  whitespace verification before rerunning the close boundary.

### 2026-07-12T23:16:00-07:00 — require complete policy fan-out

The following whole-issue review confirmed the registry-derived sweep, then
found that checking only the heading and threshold could miss truncation of the
datatype route or writer boundary. To make the propagation proof match Done-when:

- [x] Require each generated `plan.TargetAll.EntryFiles()` consumer to contain
  the complete scoped Session Continuity policy.
- [x] Re-run the focused guard, full repository suite, issue validation, and
  whitespace verification before the next close boundary.

### 2026-07-12T23:23:00-07:00 — pin fallback and checkpoint ordering semantics

The next review confirmed complete consumer propagation but found two source
semantics could still be reversed or deleted without failing the marker guard.

- [x] Assert `If an exact percentage is unavailable`, `Finish the current atomic
  action`, `update its durable issue/plan/log state first`, and the successful
  durable-write precondition independently.
- [x] Re-run the focused guard, full repository suite, issue validation, and
  whitespace verification before the next close boundary.

### 2026-07-12T23:30:00-07:00 — bind clauses and verify relative order

The next review correctly distinguished phrase presence from relational
semantics. The guard now proves the relationships, not merely their vocabulary:

- [x] Assert the complete unavailable-percentage → warning-trigger clause.
- [x] Assert the complete atomic-action → durable-state-first clause and verify
  it appears before the continuation route.
- [x] Re-run the focused guard, full repository suite, issue validation, and
  whitespace verification before the next close boundary.

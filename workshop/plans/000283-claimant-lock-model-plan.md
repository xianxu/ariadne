# Claimant lock model (#283) Implementation Plan

> **For agentic workers:** Consult AGENTS.md Section 3 (Subagent Strategy) to determine the appropriate execution approach: use superpowers-subagent-driven-development (if subagents are suitable per AGENTS.md) or superpowers-executing-plans to implement this plan. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Make the claimant the owner and the lock, and status lifecycle only. `claim` records an owner without changing status, and `start-plan` starts the lifecycle (`open → working`).

**Architecture:** The issue vocabulary (`construct/vocabulary/issue.cue`) gains an `ownership` block, the second axis beside `lifecycle`. It names the lock field, the statuses that can hold an owner, and the ownership events. `pkg/vocab` exposes it, and the claim, start-plan, reclaim and set-status decisions derive from it (ARCH-PURPOSE: every consumer derives, nothing restates the list). The decisions stay pure functions over card bytes plus a `Claimant`. The IO glue (tracker CAS, branch preparation) is unchanged in shape.

**Tech Stack:** Go (`cmd/sdlc`, `pkg/vocab`), CUE (`construct/vocabulary/issue.cue`, exported to `pkg/vocab/issue.json` by `make vocab-embed`).

**Scope boundary:** Tracker repositories only. Legacy repositories (no `workshop/issue-tracker.json`) keep claim-as-status-flip; they have no claimant. Not in this issue: the unclaim/takeover verbs and claiming several issues (#284), the transfer guard (#285), boundary pushes and abandon (#286), reconciling merges (#287).

---

## Design decisions

- **D1 — `claim` changes only the owner.** `open`+unowned → owner = me, `updated` bumped, and `started` stamped if empty. **Keeping the `started` stamp at claim is deliberate:** the operator decided shaping time counts toward the issue (#283 Log, round 2). Gap-truncation already bounds idle claim→start-plan time. Anchoring precisely when one session holds several claims belongs with #284.
- **D2 — `start-plan` starts the lifecycle.** It requires that this workspace owns the card. An `open` card is flipped to `working` by CAS after the branch is prepared. Branch first, then card: a failed CAS leaves an idempotent re-run, never a `working` card with no branch. An already-active owned card is a no-op on the card, as today.
- **D3 — `start-plan` carries the claimed issue's own details edits.** Shaping on a resting branch under a claim is the intended flow, so uncommitted edits confined to *this issue's* details file no longer block branch creation. `git switch -c` carries them onto the new branch. Any other dirty tracked file still refuses. This is the friction hit while designing #283 (Log, round 1).
- **D4 — `change-code` refuses an unstarted (`open`) card**, with next action `sdlc start-plan --issue N`. Close's tracker completion already refuses `open` (`completeop.go:93`).
- **D5 — `reclaim` accepts an owned `open` card** (a shaping claim held by a slot that is gone). The holdable-status set comes from the model.
- **D6 — Lifecycle event rename.** `open → working` is event `start` with guard `owned`, no longer `claim`. `claim` becomes an ownership event. Nothing in Go keys on the event name `claim` (checked: `grep '"claim"' pkg cmd/sdlc`).
- **D7 — `set-status open → working`** stays legal. It already refuses a foreign owner and records the setter. It is the manual spelling of `start`.

## Core concepts

### Pure entities

| Name | Lives in | Status |
|------|----------|--------|
| `ownership` (CUE block) | `construct/vocabulary/issue.cue` | new |
| `Ownership` / `OwnershipEvent` | `pkg/vocab/vocab.go` | new |
| `IssueModel.CanHoldOwner` | `pkg/vocab/vocab.go` | new |
| `claimDecision` | `cmd/sdlc/claimdecision.go` | modified |
| `startDecision` | `cmd/sdlc/startdecision.go` | new |
| `reclaimDecision` | `cmd/sdlc/reclaim.go` | modified |
| `planningDirtyAllowed` | `cmd/sdlc/planningbranch.go` | new |

- **ownership (CUE)** — the lock axis. `lock: "claimant"`. `holdable` = open ∪ active statuses. Terminal statuses keep the owner as attribution only. `events`: `claim` (none→me), `unclaim` (me→none), `reclaim` (other→me, with reason), `move` (me here→me there). Each event has `statuses` (where it applies) and a `when` gloss. No event names a status change. A law enforces it (events carry no `to` status) and keeps ownership event names disjoint from lifecycle events.
  - **Relationships:** orthogonal to `lifecycle`; both are keyed by `#Status`.
  - **DRY rationale:** replaces the status literals in `claimDecision` (`IsOpen`), `reclaimDecision` (`"working","blocked","codecomplete"`) and `adoptDecision`.
  - **Future extensions:** #284 adds `unclaim` and `takeover` verbs that read the same events; a per-status `owner: required` would hang here.
- **startDecision(card, me) → (next []byte, changed bool, err)** — pure. Owned `open` → `status: working`, `updated`, `started` if empty. Owned and already `working`/`blocked`/`codecomplete` → unchanged. Unowned → refuse toward claim. Foreign → refuse naming the owner. Terminal → refuse. The status edge is checked against `vocab.Issue().TransitionForEvent(status, "start")`.
- **planningDirtyAllowed(porcelainZ, detailPath) → (blocking []string)** — pure. Splits dirty tracked entries into those allowed (exactly this issue's details file) and those that block.

### Integration points

| Name | Lives in | Status | Wraps |
|------|----------|--------|-------|
| `startPlanBranch` | `cmd/sdlc/startplan.go` | modified | tracker CAS + git branch |
| `preparePlanningBranch` | `cmd/sdlc/planningbranch.go` | modified | git status/switch |
| `runClaim` | `cmd/sdlc/claim.go` | modified | tracker CAS |
| change-code ownership gate | `cmd/sdlc/changecode*.go` | modified | tracker read |

- **startPlanBranch** — prepares the branch, then `env.repo.UpdateCard` with `startDecision`'s output when `changed`. Tested through the existing real-git tracker fixtures (`startplan_test.go`, `claimant_test.go` helpers), not mocks.

## Chunk 1: Model

### Task 1: `ownership` block in the issue vocabulary

**Files:**
- Modify: `construct/vocabulary/issue.cue`
- Regenerate: `pkg/vocab/issue.json` (`make vocab-embed`)
- Modify: `pkg/vocab/vocab.go`, `pkg/vocab/lifecycle.go` (types)
- Test: `pkg/vocab/vocab_test.go`, `pkg/vocab/conformance_test.go`

- [ ] **Step 1: Failing tests** in `pkg/vocab/vocab_test.go`:

```go
func TestOwnershipAxis(t *testing.T) {
	m := Issue()
	if got := m.Ownership().Lock; got != "claimant" {
		t.Fatalf("lock = %q, want claimant", got)
	}
	for _, s := range []string{"open", "working", "blocked", "codecomplete"} {
		if !m.CanHoldOwner(s) {
			t.Errorf("%s must be holdable", s)
		}
	}
	for _, s := range []string{"done", "wontfix", "punt"} {
		if m.CanHoldOwner(s) {
			t.Errorf("%s must not be holdable (attribution only)", s)
		}
	}
	for _, ev := range []string{"claim", "unclaim", "reclaim", "move"} {
		if m.OwnershipEvent(ev) == nil {
			t.Errorf("ownership event %q missing", ev)
		}
		if m.FirstTransitionForEvent(ev) != nil {
			t.Errorf("%q is an ownership event; it must not be a lifecycle event", ev)
		}
	}
	if tr := m.TransitionFor("open", "working"); tr == nil || tr.Event != "start" {
		t.Fatalf("open→working must be event start, got %+v", tr)
	}
}
```

(Add `FirstTransitionForEvent`/`TransitionFor` on `IssueModel` only if they don't already exist; `conformance_test.go:121` uses `TransitionFor`.)

- [ ] **Step 2:** `go test ./pkg/vocab/ -run TestOwnershipAxis`, expect FAIL (no `Ownership`).
- [ ] **Step 3: CUE.** In `issue.cue`, change the lifecycle row to `{from: "open", to: "working", event: "start", guards: ["owned"]}`. Add:

```cue
// ── ownership: the lock axis (#283). The claimant on the card is the owner AND
// the lock; status is lifecycle only. Ownership events never change status. ──
#OwnershipEvent: {
	event:    string
	owner:    "none→me" | "me→none" | "other→me" | "me→me"
	statuses: [...#Status]
	when:     string & !=""
}
ownership: {
	lock:     "claimant"
	holdable: list.Concat([categories.open, categories.active])
	// terminal statuses keep the last owner as attribution, never as a lock
	events: [...#OwnershipEvent] & [
		{event: "claim", owner: "none→me", statuses: holdable, when: "take the lock: shaping (open) or taking over unowned started work"},
		{event: "unclaim", owner: "me→none", statuses: holdable, when: "release the lock; started work stays started and is open for takeover"},
		{event: "reclaim", owner: "other→me", statuses: holdable, when: "operator-directed transfer, with a reason"},
		{event: "move", owner: "me→me", statuses: categories.active, when: "relocate the owner's own work between worktrees on one machine"},
	]
}
```

Add laws: `"ownership-disjoint"` (no ownership event name in `_events` of the lifecycle) and `"holdable-nonterminal"` (no terminal status in `holdable`). Follow the existing `laws` pattern with `list.Contains(...) & false`.

- [ ] **Step 4:** `make vocab-embed`. It regenerates `issue.json`; the trailing `git diff --exit-code` fails by design on a changed export, so check the diff by eye. Add Go types:

```go
type OwnershipEvent struct {
	Event    string   `json:"event"`
	Owner    string   `json:"owner"`
	Statuses []string `json:"statuses"`
	When     string   `json:"when"`
}
type Ownership struct {
	Lock     string           `json:"lock"`
	Holdable []string         `json:"holdable"`
	Events   []OwnershipEvent `json:"events"`
}
```

Add the field `Own Ownership `json:"ownership"`` on `IssueModel`, plus `Ownership()`, `CanHoldOwner(s)` (= `contains(m.Own.Holdable, s)`) and `OwnershipEvent(name) *OwnershipEvent`.

- [ ] **Step 5:** `go test ./pkg/vocab/... && cue vet` (via the `vocabulary` tool's own test), expect PASS. Run `go test ./...` once to surface every consumer of the event rename.
- [ ] **Step 6: Commit** `#283: vocab: ownership axis; open→working is start`.

## Chunk 2: Verbs

### Task 2: `claimDecision` records the owner only

**Files:** Modify `cmd/sdlc/claimdecision.go`, `cmd/sdlc/claim.go` (messages). Test: `cmd/sdlc/claim_test.go` (pure decision table).

- [ ] **Step 1: Failing table test** `TestClaimDecisionOwnerOnly`, covering every status × {none, me, other}:
  - open/none → status stays `open`, claimant = me, `started` set.
  - open/me → `errAlreadyMine`.
  - open/other → error containing `reclaim` and the owner's worktree.
  - working|blocked|codecomplete/me → `errAlreadyMine`.
  - …/other → refuse naming the owner.
  - …/none → refuse toward `--adopt` (unchanged; #284 folds it).
  - terminal → refuse.

  Build cards with the existing test helpers in `claim_test.go`/`claimant_test.go` (`issue.SetCardClaimant` on a fixture card).
- [ ] **Step 2:** Run it; expect FAIL (status becomes `working`).
- [ ] **Step 3: Implement.** Gate on `vocab.Issue().CanHoldOwner(status)`. Read the recorded claimant for **every** holdable status, not only `working` (`ownedBy` already explains mine/foreign/unknown). Only an unowned `open` card is written: drop the `SetField(fm, "status", "working")`. Keep `updated` and the `started`-if-empty stamp (D1). Update `ownedBy`'s messages to not say "is working". Use the status: "is open, claimed by …".
- [ ] **Step 4:** In `claim.go`, change the success line to `Issue #N claimed on issue-tracker: this workspace owns it; status stays open — `sdlc start-plan --issue N` starts it.` Also update the relocation branch's `status == "working"` check to `CanHoldOwner(status)`.
- [ ] **Step 5:** `go test ./cmd/sdlc/ -run 'Claim'`; PASS. Commit `#283: claim: record the owner, leave status alone`.

### Task 3: `startDecision` + `start-plan` starts the lifecycle

**Files:** Create `cmd/sdlc/startdecision.go`, `cmd/sdlc/startdecision_test.go`. Modify `cmd/sdlc/startplan.go` (`startPlanBranch`), `cmd/sdlc/planningbranch.go`.

- [ ] **Step 1: Failing table test** `TestStartDecision`, status × owner:
  - open/me → changed, `working`, `started` kept if present, else set.
  - working/me → unchanged, no error.
  - open/none → error containing `sdlc claim --issue`.
  - open/other, working/other → error naming the owner.
  - done → error.
- [ ] **Step 2:** Run it; FAIL.
- [ ] **Step 3: Implement**:

```go
// startDecision is start-plan's pure core (#283): the owner starts the
// lifecycle. Ownership is the lock; this only moves status, and only along the
// model's `start` edge.
func startDecision(card []byte, id string, me issue.Claimant, today, started string) ([]byte, bool, error) {
	fm, _, err := issue.Parse(string(card))
	if err != nil {
		return nil, false, err
	}
	status, _ := issue.GetField(fm, "status")
	if !vocab.Issue().CanHoldOwner(status) {
		return nil, false, fmt.Errorf("#%s is %s; there is nothing to plan", id, status)
	}
	recorded, has, err := issue.CardClaimant(card)
	if err != nil {
		return nil, false, err
	}
	if !has {
		return nil, false, fmt.Errorf("#%s has no owner; `sdlc claim --issue %s` takes it before planning", id, issue.CLIRef(id))
	}
	if issue.MatchClaimant(&recorded, me) != issue.OwnershipMine {
		return nil, false, fmt.Errorf("#%s is owned by %s; planning belongs to its owner", id, describeClaimant(recorded))
	}
	if vocab.Issue().TransitionForEvent(status, "start") == nil {
		return card, false, nil // already started
	}
	out, err := issue.SetCardField(card, "status", "working")
	if err == nil {
		out, err = issue.SetCardField(out, "updated", today)
	}
	if err == nil {
		if cur, _ := issue.GetField(fm, "started"); strings.TrimSpace(cur) == "" {
			out, err = issue.SetCardField(out, "started", started)
		}
	}
	return out, err == nil, err
}
```

  Add `TransitionForEvent` on `IssueModel` if missing (`transitionForEvent` exists in `lifecycle.go`).
- [ ] **Step 4: Rewire `startPlanBranch`.**
  - Remove the `status != "working"` refusal.
  - Keep `requireCardOwnership` for its relocation and adopt hints, but call it only when the card is not mine: `startDecision` already covers the basic cases.
  - Order: `startDecision` (validate, no write) → `preparePlanningBranch` → if `changed`, `env.repo.UpdateCard(card, next, operationToken("start"), nil-check)` → `cok("#N started (open → working)")`.
  - On `tracker.ErrCardChanged`, say to re-run `start-plan`, which is idempotent.
- [ ] **Step 5: D3.** In `preparePlanningBranch`, replace the raw `status --porcelain` dirty check with `status --porcelain=v1 -z --untracked-files=no` → `gitx.ParseStatusZ` (lessons.md: never slice porcelain) → `planningDirtyAllowed(entries, detailPath)`. Refuse only on the blocking remainder. Pure unit test `TestPlanningDirtyAllowed`:
  - only the details file → none blocking;
  - details + `cmd/x.go` → `cmd/x.go` blocking;
  - another issue's details → blocking.
- [ ] **Step 6: Real-git test** in `startplan_test.go`, using the existing tracker fixture:
  1. claim → card `open` with an owner;
  2. edit the details on the resting branch;
  3. `start-plan` → branch created, edit carried, card `working`, `started` unchanged from claim;
  4. re-run `start-plan` → no-op.
- [ ] **Step 7:** `go test ./cmd/sdlc/ -run 'StartPlan|StartDecision|PlanningDirty'`; PASS. Commit `#283: start-plan: owner starts the lifecycle`.

### Task 4: change-code refuses unstarted; reclaim and set-status derive from the model

**Files:** `cmd/sdlc/changecode*.go` (where `requireCardOwnership` is called), `cmd/sdlc/reclaim.go`, `cmd/sdlc/setstatus.go`; tests beside each.

- [ ] **Step 1: Failing tests.**
  - change-code on an owned `open` card → error containing `sdlc start-plan --issue`.
  - `reclaimDecision` on an owned `open` card with reason + expect → succeeds.
  - `reclaimDecision` on an unowned `open` card → refuses toward claim.
- [ ] **Step 2: Implement.**
  - change-code: after the ownership gate, read the card status. `vocab.Issue().IsOpen(status)` → refuse.
  - reclaim: replace the status `switch` with `CanHoldOwner`. Unowned (any holdable status) → the `has` check already refuses; reword it to "`sdlc claim`" for `open`.
  - set-status: no logic change (D7). Update the refusal text that says "claimed before #277 … `--adopt`" only if it now misleads.
- [ ] **Step 3:** Run the tests; PASS. Commit `#283: change-code/reclaim: lifecycle vs lock`.

### Task 5: Fix e2e fixtures that assumed claim → working

- [ ] **Step 1:** `go test ./cmd/sdlc/...`. Expect failures in the 12 files that run `claim` then a verb expecting `working`: `closetracker_test.go`, `flow_e2e_test.go`, `leftover_e2e_test.go`, `observe_test.go`, `claimremote_test.go`, `issuemigrate_test.go`, … .
- [ ] **Step 2:** Insert `start-plan --issue N` after `claim` where the test exercises started work. Update assertions that read `status: working` straight after claim to expect `open` plus a claimant. Don't weaken any assertion; if a test asserted claim → working *as its subject*, rewrite it to assert the new contract.
- [ ] **Step 3:** `go test ./...` green. Commit `#283: tests: claim then start-plan`.

## Chunk 3: Words

### Task 6: Help text, atlas, terminology

**Files:**
- `cmd/sdlc/helptext/claim.md`
- `cmd/sdlc/helptext/root.md` (BEFORE WORK)
- `cmd/sdlc/helptext/start-plan.md`
- `cmd/sdlc/helptext/set-status.md`
- `cmd/sdlc/helptext/reclaim.md`
- `cmd/sdlc/main.go:136` (claim short)
- `atlas/workflow/issue-lifecycle.md` (§Claim)
- `atlas/workflow/issue-tracker.md` (line 23 and the verb table at line 203)
- `atlas/workflow/sdlc-binary.md:66`
- `atlas/workflow/vocabulary.md` (new "Ownership axis" section)
- `atlas/index.md` if a file is added
- `atlas/workflow/process-manual.md` (regenerate with `sdlc process-manual`)
- AGENTS.md §2 "Claim early" sentence (base layer: one clause)

- [ ] **Step 1:** Rewrite claim help:
  - claim takes the lock and leaves status alone;
  - shaping under a claim;
  - `start-plan` starts the lifecycle;
  - a slot can hold an `open` claim.
- [ ] **Step 2:** Add a **Terminology** paragraph to `vocabulary.md`, linked from `issue-tracker.md`:
  - **owner** = the claimant = a slot (repository + machine + worktree), the durable workspace of one agent's thread, which outlives sessions;
  - **operator** = the supervising human, descriptive, the escalation point;
  - **issue branch** = the derived branch, never called "owner".
- [ ] **Step 3: Sweep.** `grep -rn "owner" cmd/sdlc/helptext atlas/workflow` and `grep -rn "open → working" …`. Fix every hit where "owner" means a branch or a person, and every hit where claim is described as open → working. Transfer-guard *code* strings are #285's; list them in the Log and leave them.
- [ ] **Step 4:** `go test ./cmd/sdlc/helptext/... ./pkg/vocab/...` (prose-drift tests); regenerate the process manual. Commit `#283: docs: claimant is the lock; terminology`.

### Task 7: Verify and close

- [ ] `go test ./...`; `make vocab-embed` clean.
- [ ] Manual smoke in a scratch tracker fixture (or this repo with `--dry-run`): `claim` → card `open` + claimant; `start-plan` → `working` + branch.
- [ ] Log the outcome; `sdlc close --issue 283 --verified '…'`.

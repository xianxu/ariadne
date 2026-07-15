# no-verdict single-pass recovery — Implementation Plan

> **For agentic workers:** Consult AGENTS.md Section 3 (Subagent Strategy) to determine the appropriate execution approach: use superpowers-subagent-driven-development (if subagents are suitable per AGENTS.md) or superpowers-executing-plans to implement this plan. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** `sdlc close` accepts the issue-close boundary review as covering *trailing* milestones that never got their own `milestone-close`, so single-pass work with legacy Mx tags closes without `--no-verdict`; refusals for the remaining cases name the sanctioned recovery (AGENTS.md §3), and the plan-quality judge flags over-split Mx plans at design time.

**Architecture:** The verdict gate in `computeClose` currently refuses whenever any Mx row lacks a `Review-Verdict:` commit. The fix partitions the missing set (pure function) into **mid-stream** misses (a later milestone *did* close with a review — the review-before-next-boundary invariant was genuinely violated) and **trailing** misses (no later reviewed boundary — the imminent issue-close review, whose window is branch-point→HEAD, covers their work; the close *is* their boundary). Trailing misses are accepted with a loud info line — but only when the close review will actually run (`--no-judge` voids the coverage premise, so trailing misses then refuse). All refusal text gains the §3 citation + concrete recovery. Forward fix: a new failure-mode bullet in the plan-quality prompt catches over-split plans before the trap is set.

**Tech Stack:** Go (cmd/sdlc), existing test seams: `closeRepo`/`stubJudge` (closereview_test.go), `expectDie` (die_test.go).

**Why coverage is real, not fictional (the acceptance rationale):** `boundaryWindowBase(issueStr, "", issuePath)` for a whole-issue close is `merge-base(main, HEAD)` (or the issue's branch start) — the review dispatched by `sdlc close` diffs the *entire branch*, including every unclosed milestone's commits. For a trailing miss, no boundary was ever crossed without review — the close review arrives at the first (and only) boundary. For a mid-stream miss, a later `milestone-close` boundary *was* crossed without evidence for the earlier row — that stays a refusal (retroactive `sdlc judge milestone-review` is the recovery).

---

## Core concepts

### Pure entities

| Name | Lives in | Status |
|------|----------|--------|
| `partitionMissingVerdicts` | `cmd/sdlc/close.go` | new |
| `findMilestonesMissingVerdict` | `cmd/sdlc/close.go` | modified |
| `formatMissingVerdicts` | `cmd/sdlc/close.go` | modified |
| `formatTrailingVerdictAccepted` | `cmd/sdlc/close.go` | new |
| `formatTrailingNeedsJudge` | `cmd/sdlc/close.go` | new |

- **`partitionMissingVerdicts(ordered, missing []string) (midstream, trailing []string)`** — splits the missing set by plan position relative to the *last* verdict-carrying milestone. Missing rows before it are mid-stream; after it (or all of them, when no row carries a verdict) are trailing. Pure: slices in, slices out, plan order preserved.
  - **Relationships:** consumes `findMilestonesMissingVerdict`'s two outputs; 1:1 with the verdict-gate block.
  - **DRY rationale:** first occurrence; keeps the gate's branch logic testable without git (ARCH-PURE).
  - **Future extensions:** if we later accept mid-stream misses covered by a later milestone's window, this is where the classification widens.
- **`findMilestonesMissingVerdict`** — signature widens to `(ordered, missing []string, err error)` so the caller can partition. Enumeration/git-probe behavior unchanged.
- **`formatMissingVerdicts`** — refusal text gains the §3 citation ("an Mx tag is a review boundary — single-pass work should use plain checkboxes") and the fold-to-plain-checkboxes recovery alongside the existing per-row `sdlc judge milestone-review` next-actions. MUST keep the closing line `Or pass --no-verdict (or --force); record the reason in --verified.` — `internal/processmanual/gatesig.go:90` pins `RefusalPat: Or pass --no-verdict \(or --force\); record` for friction measurement (#172).
- **`formatTrailingVerdictAccepted(trailing []string)`** — the loud acceptance info line: names the accepted milestones, states the issue-close review's window covers them (#175), and hints §3 (plain checkboxes for single-pass work).
- **`formatTrailingNeedsJudge(issueStr string, trailing []string)`** — refusal when trailing misses exist but `--no-judge` skips the close review (coverage premise gone). Names the recovery: drop `--no-judge`, run `sdlc judge milestone-review` per row, or pass `--no-verdict`. Ends with the same gatesig-pinned closing line so friction measurement keys on one signature.

### Integration points

| Name | Lives in | Status | Wraps |
|------|----------|--------|-------|
| verdict-gate block | `cmd/sdlc/close.go:477-489` (computeClose) | modified | git log (via existing probe) + stderr |
| plan-quality prompt | `cmd/sdlc/internal/judge/prompts/plan-quality.md` | modified | LLM judge at `sdlc change-code` |

- **verdict-gate block** — thin dispatch over the pure partition: skip-ack (unchanged, gatesig `AckPat` preserved) → mid-stream refusal → trailing+`--no-judge` refusal → trailing acceptance info. Wiring is pinned by integration tests through `runCloseWithReview` (per the lessons.md "pure helper can be silently un-wired" rule).
  - **Injected into:** n/a — it injects the partition results into `die`/`cinfo`.
- **plan-quality prompt** — new failure-mode bullet: every `## Plan` row Mx-tagged for work that will land in one pass = over-split; each Mx commits to its own `milestone-close` (AGENTS.md §3); suggest plain checkboxes unless ≥2 genuinely separate boundaries.
  - **Injected into:** `sdlc change-code`'s plan-quality dispatch (existing plumbing; content-only change, pinned by a prompt-content test in the judge package).

**Test surface:** `partitionMissingVerdicts` gets a colocated table test (no IO). The gate wiring gets three integration tests on the existing `closeRepo` + `stubJudge` + `expectDie` seams. `judge.Run` call-count assertions pin *where* the refusal fires (mutation-check discipline, lessons.md #63).

---

## Chunk 1: pure core + gate rewiring

### Task 1: `partitionMissingVerdicts` (pure, TDD)

**Files:**
- Modify: `cmd/sdlc/close.go` (next to `findMilestonesMissingVerdict`)
- Test: `cmd/sdlc/close_test.go`

- [ ] **Step 1: Write the failing table test**

```go
// TestPartitionMissingVerdicts pins the trailing-vs-midstream split (#175):
// a missing milestone BEFORE the last verdict-carrying one is a genuine
// skipped-review violation (midstream); missing milestones after it (or all,
// when none carries a verdict — the single-pass case) are trailing and
// covered by the imminent issue-close review.
func TestPartitionMissingVerdicts(t *testing.T) {
	cases := []struct {
		name               string
		ordered, missing   []string
		wantMid, wantTrail []string
	}{
		{"single-pass: all missing → all trailing",
			[]string{"M1", "M2", "M3"}, []string{"M1", "M2", "M3"},
			nil, []string{"M1", "M2", "M3"}},
		{"midstream: miss before a verdict-carrying row",
			[]string{"M1", "M2", "M3"}, []string{"M2"},
			[]string{"M2"}, nil}, // M3 has a verdict → M2's boundary was crossed unreviewed
		{"mixed: M1 missing before M2's verdict, M3 trailing",
			[]string{"M1", "M2", "M3"}, []string{"M1", "M3"},
			[]string{"M1"}, []string{"M3"}},
		{"none missing",
			[]string{"M1", "M2"}, nil, nil, nil},
		{"only last missing → trailing",
			[]string{"M1", "M2"}, []string{"M2"}, nil, []string{"M2"}},
		// The reopened-issue shape: prior milestones all reviewed, one new
		// trailing Mx added by the reopen — second most likely real-world hit.
		{"reopened issue: new trailing row after all-reviewed history",
			[]string{"M1", "M2", "M3"}, []string{"M3"}, nil, []string{"M3"}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			mid, trail := partitionMissingVerdicts(tc.ordered, tc.missing)
			if strings.Join(mid, ",") != strings.Join(tc.wantMid, ",") {
				t.Errorf("midstream = %v, want %v", mid, tc.wantMid)
			}
			if strings.Join(trail, ",") != strings.Join(tc.wantTrail, ",") {
				t.Errorf("trailing = %v, want %v", trail, tc.wantTrail)
			}
		})
	}
}
```

- [ ] **Step 2: Run to verify it fails** — `go test ./cmd/sdlc/ -run TestPartitionMissingVerdicts` → FAIL (undefined `partitionMissingVerdicts`).

- [ ] **Step 3: Implement**

```go
// partitionMissingVerdicts splits the missing-verdict milestones by plan
// position relative to the LAST verdict-carrying milestone (#175). Missing
// rows before it are "midstream" — a later boundary was crossed with no
// review evidence for them, a genuine §3 violation. Missing rows after it
// (or every row, when nothing carries a verdict — the single-pass case)
// are "trailing" — no reviewed boundary follows them, so the imminent
// issue-close boundary review (window branch-point→HEAD) is their review.
// Pure; plan order is preserved in both outputs.
func partitionMissingVerdicts(ordered, missing []string) (midstream, trailing []string) {
	missingSet := make(map[string]bool, len(missing))
	for _, tag := range missing {
		missingSet[tag] = true
	}
	last := -1 // index of the last verdict-carrying milestone
	for i, tag := range ordered {
		if !missingSet[tag] {
			last = i
		}
	}
	for i, tag := range ordered {
		if !missingSet[tag] {
			continue
		}
		if i < last {
			midstream = append(midstream, tag)
		} else {
			trailing = append(trailing, tag)
		}
	}
	return midstream, trailing
}
```

- [ ] **Step 4: Run to verify it passes** — same command, PASS.
- [ ] **Step 5: Commit** — `git add cmd/sdlc/close.go cmd/sdlc/close_test.go && git commit -m "#175: partitionMissingVerdicts — pure trailing-vs-midstream split"`

### Task 2: widen `findMilestonesMissingVerdict` to return plan order

**Files:**
- Modify: `cmd/sdlc/close.go:1400` (`findMilestonesMissingVerdict`), `close.go:478` (caller)
- Test: `cmd/sdlc/close_test.go` (5 existing tests call it: `_NoMilestones`, `_NoPlanSection`, `_Integration`, `_AllPresent`, `_SpaceBeforeColonSubject`)

- [ ] **Step 1:** Change signature to `(ordered, missing []string, err error)`; return the already-computed `ordered` slice (nil in the no-plan/no-milestones early returns). Update doc comment.
- [ ] **Step 2:** Update the 5 existing test call sites (`ordered, missing, err :=`); in `_Integration`, additionally assert `ordered == {M1, M2, M3}` (pins the new return).
- [ ] **Step 3:** Update the caller at `close.go:478` (compile fix only for now — `ordered, missing, err :=`, ignore `ordered` with `_` until Task 4).
- [ ] **Step 4:** `go test ./cmd/sdlc/ -run TestFindMilestonesMissingVerdict` → PASS; `go vet ./cmd/sdlc/` clean.
- [ ] **Step 5: Commit** — `#175: findMilestonesMissingVerdict returns plan order`

### Task 3: message formatters (pure, TDD)

**Files:**
- Modify: `cmd/sdlc/close.go` (`formatMissingVerdicts` + two new formatters, colocated)
- Test: `cmd/sdlc/close_test.go` (extend `TestFormatMissingVerdicts_ContractElements`, add two new contract tests)

- [ ] **Step 1: Write/extend failing contract tests**

```go
// extend TestFormatMissingVerdicts_ContractElements's want list with:
//   "AGENTS.md §3",                        // don't-over-split citation
//   "plain checkboxes",                    // fold recovery
//   "Or pass --no-verdict (or --force); record",  // gatesig-pinned closing line

func TestFormatTrailingVerdictAccepted_ContractElements(t *testing.T) {
	msg := formatTrailingVerdictAccepted([]string{"M1", "M2"})
	for _, w := range []string{
		"M1, M2",                 // names the accepted milestones
		"issue-close boundary review", // what covers them
		"#175",                   // provenance
		"plain checkboxes",       // §3 hint for next time
	} {
		if !strings.Contains(msg, w) {
			t.Errorf("formatTrailingVerdictAccepted missing %q in:\n%s", w, msg)
		}
	}
}

func TestFormatTrailingNeedsJudge_ContractElements(t *testing.T) {
	msg := formatTrailingNeedsJudge("31", []string{"M1"})
	for _, w := range []string{
		"M1",
		"--no-judge",            // names the premise-killer
		"sdlc judge milestone-review --issue 31 --milestone M1",
		"Or pass --no-verdict (or --force); record", // same gatesig signature
	} {
		if !strings.Contains(msg, w) {
			t.Errorf("formatTrailingNeedsJudge missing %q in:\n%s", w, msg)
		}
	}
}
```

- [ ] **Step 2: Run to verify failures** — undefined formatters / missing strings.
- [ ] **Step 3: Implement.** `formatMissingVerdicts` inserts, after the existing "Each milestone close must carry…" paragraph:

```
  An `Mx` tag in ## Plan is a review boundary, not a task label (AGENTS.md
  §3) — single-pass work should use plain checkboxes. If this plan was
  over-split, the sanctioned recovery is to fold the never-closed Mx rows
  into plain checkboxes (append a ## Revisions note saying why); otherwise
  land the per-row review evidence:
```

  New formatters (shapes; exact prose at implementer's discretion within the tested contract):

```go
// formatTrailingVerdictAccepted builds the loud acceptance line for trailing
// unclosed milestones (#175): the issue-close boundary review's window is
// branch-point→HEAD, so their work is inside the diff this close is about to
// review — the close IS their review boundary. Pure.
func formatTrailingVerdictAccepted(trailing []string) string {
	return fmt.Sprintf("milestones %s never had their own milestone-close; accepted — the "+
		"issue-close boundary review (window branch-point→HEAD) covers their work (#175). "+
		"Next time: single-pass work takes plain checkboxes, not Mx tags (AGENTS.md §3).",
		strings.Join(trailing, ", "))
}

// formatTrailingNeedsJudge builds the refusal for trailing unclosed
// milestones when --no-judge skips the issue-close review — the coverage
// premise behind the #175 acceptance is gone. Pure. Ends with the same
// closing line gatesig.go pins for the no-verdict gate.
func formatTrailingNeedsJudge(issueStr string, trailing []string) string { … }
```

- [ ] **Step 4: Run** — `go test ./cmd/sdlc/ -run 'TestFormat.*Verdict|TestFormatTrailing'` → PASS.
- [ ] **Step 5: Commit** — `#175: verdict-gate messages — §3 citation, fold recovery, trailing acceptance/needs-judge`

### Task 4: rewire the gate block + integration tests

**Files:**
- Modify: `cmd/sdlc/close.go:477-489` (verdict-gate block in computeClose)
- Test: `cmd/sdlc/close_finalize_test.go` (or closereview_test.go — wherever `closeRepo`/`stubJudge` are visible)

- [ ] **Step 1: Write the three failing integration tests** on the existing seams. Repo fixture: `closeRepo(t, 31)` then rewrite the issue file's Plan to Mx-tagged ticked rows and commit. All three must satisfy the *other* gates without touching verdict/judge flags: `Verified` set, `NoActual`, `NoAtlas`, `NoProject`, `NoReclose` as needed (mirror what nearby finalize tests pass) — never `Force`.

```go
// #175: single-pass work with legacy Mx tags closes WITHOUT --no-verdict —
// the trailing misses are covered by the issue-close boundary review.
func TestClose_TrailingUnclosedMilestones_AcceptedByCloseReview(t *testing.T) {
	// plan rows: "- [x] **M1 — all the work**" (no milestone-close commit exists)
	// stubJudge SHIP; flags WITHOUT NoVerdict/NoJudge/Force
	// assert: close finalizes — issue file status: codecomplete (#160: close
	//         never writes `done`; merge/push do) + the closed log line with
	//         "review verdict: SHIP" (same pin as closereview_test.go:254),
	//         stderr contains "M1" + "issue-close boundary review" (acceptance line),
	//         judge call count == 1 (the review actually ran — the premise).
	// fixture note: readIssue (closereview_test.go:364) hardcodes 000069-x.md —
	// use closeRepo(t, 69) to reuse it, or read the issue file directly.
}

// #175: a midstream miss (M1 unreviewed, M2 closed WITH trailer) still refuses,
// BEFORE the review dispatch, and the refusal names the §3 recovery.
func TestClose_MidstreamMissingVerdict_Refuses(t *testing.T) {
	// commit "#31 M2: close — tick" with Review-Verdict trailer touching the issue file
	// expectDie; assert message names M1 (and NOT M2), contains "AGENTS.md §3",
	//            judge call count == 0 (refused at the gate, not later — #63 discipline).
}

// #175: trailing miss + --no-judge → the coverage premise is gone → refuse.
func TestClose_TrailingMissingVerdict_NoJudgeRefuses(t *testing.T) {
	// same fixture as the accepted test but f.NoJudge = true
	// expectDie; assert message contains "--no-judge" and the gatesig closing line.
}
```

- [ ] **Step 2: Run to verify all three fail** (first: no acceptance — gate refuses; others: message/dispatch assertions).
- [ ] **Step 3: Rewire the gate block:**

```go
if mode == "issue" {
	ordered, missing, err := findMilestonesMissingVerdict(body, issueStr, issuePath)
	if err != nil {
		cwarn(stderr, fmt.Sprintf("milestone-verdict check skipped: %v", err))
	} else if len(missing) > 0 {
		midstream, trailing := partitionMissingVerdicts(ordered, missing)
		switch {
		case f.skip("verdict"):
			cwarn(stderr, fmt.Sprintf("--no-verdict (or --force): skipping Review-Verdict check for %d milestone(s): %s",
				len(missing), strings.Join(missing, ", ")))
		case len(midstream) > 0:
			explainMissingVerdicts(stderr, issueStr, midstream)
			exitWithCode(1)
		case f.skip("judge"):
			// Trailing-only, but --no-judge skips the very review that
			// would cover them — the #175 acceptance premise is gone.
			die(stderr, formatTrailingNeedsJudge(issueStr, trailing))
			exitWithCode(1)
		default:
			cinfo(stderr, formatTrailingVerdictAccepted(trailing))
		}
	}
}
```

- [ ] **Step 4: Run** — the three new tests PASS; full `go test ./cmd/sdlc/...` green.
- [ ] **Step 5: Mutation-check the accepted test's teeth** (lessons.md #63): temporarily make `partitionMissingVerdicts` return everything as midstream → `TestClose_TrailingUnclosedMilestones_AcceptedByCloseReview` must go RED; restore.
- [ ] **Step 6: Commit** — `#175: verdict gate accepts trailing unclosed milestones (issue-close review covers them)`

## Chunk 2: forward fix + docs

### Task 5: plan-quality prompt — over-split failure mode

**Files:**
- Modify: `cmd/sdlc/internal/judge/prompts/plan-quality.md`
- Test: `cmd/sdlc/internal/judge/` (colocate with existing prompt tests; check `golden_test.go`/`judge_test.go` for the established pattern of pinning prompt content)

- [ ] **Step 1: Failing test** — extend `TestBuildPrompt_PlanQuality_HasContract` (judge_test.go:301, the established prompt-content pin via `BuildPrompt(PlanQuality, …)` + `strings.Contains`) with the over-split key phrases: `Over-split milestones`, `review boundary`, `plain checkboxes`.
- [ ] **Step 2:** Add the bullet to the failure-modes list in `plan-quality.md`:

```
  - Over-split milestones — every ## Plan row is tagged Mx for work that will
    plainly land in one pass. An Mx tag is a review boundary (AGENTS.md §3):
    each one commits to its own `sdlc milestone-close` + review. Single-pass
    atomic work takes plain checkboxes; flag the plan unless it genuinely has
    ≥2 boundaries the author will close separately.
```

- [ ] **Step 3:** `go test ./cmd/sdlc/internal/judge/` → PASS.
- [ ] **Step 4: Commit** — `#175: plan-quality judge flags over-split Mx plans (forward fix)`

### Task 6: atlas + issue bookkeeping

**Files:**
- Modify: `atlas/workflow/sdlc-binary.md` (the close-gates section around the `--no-verdict` mention at :384 and the Review-Verdict discussion at :638)
- Modify: `workshop/issues/000175-no-verdict-single-pass-recovery.md` (tick Plan, Log)

- [ ] **Step 1:** Document the trailing-acceptance rule in the atlas close-gates prose: verdict gate = per-milestone evidence, EXCEPT trailing unclosed milestones are covered by the issue-close review (and `--no-judge` voids that). One short paragraph; map, don't over-specify.
- [ ] **Step 2:** Update issue `## Log` with the design rationale (cite ARCH-PURE for the partition split, ARCH-PURPOSE for shipping all three spec candidates); tick Plan rows.
- [ ] **Step 3: Commit** with the close bookkeeping.

### Verification (Done-when check)

- [ ] `go test ./cmd/sdlc/...` green; `go vet ./cmd/sdlc/...` clean (build with `-o /dev/null` — lessons.md stray-binary tripwire).
- [ ] Dogfood (lessons.md "dogfood on real data"): `TestClose_TrailingUnclosedMilestones_AcceptedByCloseReview` is the automated proof of Done-when line 1; additionally run `sdlc close --dry-run` semantics mentally against a real historical single-pass issue shape if one is at hand.
- [ ] Done-when line 2 ("re-measure via-bypass drop") is a *lagging* metric — note in the Log that it's measurable only after future closes accumulate; the gate-level test is the leading proof.

### Notes for the implementer

- The `exitWithCode(1)` after `explainMissingVerdicts`/`die` is belt-and-suspenders (die already exits); keep the existing pattern.
- `gatesig.go:88-90` pins BOTH the ack pattern and the refusal closing line — do not reword `--no-verdict (or --force): skipping Review-Verdict check` nor drop `Or pass --no-verdict (or --force); record` from any refusal path. Run `go test ./cmd/sdlc/internal/processmanual/` too.
- ARCH-DRY: reuse `explainMissingVerdicts`/`die` for the new refusal; do not duplicate the milestone-review next-action lines — extract a small shared helper if both formatters need them.

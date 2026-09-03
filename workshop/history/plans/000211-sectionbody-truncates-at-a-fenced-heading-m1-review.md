# Boundary Review — ariadne#211 (milestone M1)

| field | value |
|-------|-------|
| issue | 211 — SectionBody truncates at a fenced heading |
| repo | ariadne |
| issue file | workshop/issues/000211-sectionbody-truncates-at-a-fenced-heading.md |
| boundary | milestone M1 |
| milestone | M1 |
| window | 318689b02e0d9c347897cacc9b41be73d175f3b6..28057e24bfe4b9e80807ac457e72cfdf01f6ae40 |
| command | sdlc milestone-close --issue 211 --milestone M1 |
| reviewer | claude |
| timestamp | 2026-09-02T18:59:55-07:00 |
| verdict | FIX-THEN-SHIP |

## Review

```verdict
verdict: FIX-THEN-SHIP
confidence: high
```

The scanner itself is correct and the fix is real: I independently re-ran the deleted `PlanSectionRE` against `planWithFencedHeading` and confirmed all three close-gate regressions go red without it (0 unchecked vs 2, `[M1]` vs `[M1 M2]`, `(2,2)` vs `(4,2)`), and I diffed old-vs-new section extraction for `Spec`/`Plan`/`Done when` across all 406 `workshop/**/*.md` files — exactly one file changes, and its change is the intended fix (`workshop/history/plans/000145-…-plan.md` quotes `## Spec`/`## Plan`/`## Done when` inside a ```` ```go ```` block; the old regex wrongly found them). Zero gate-verdict flips on issue files, matching the Log's prediction. What blocks SHIP is not the scanner but the evidence around it: two of the five fence-form cases pass unchanged under the *old* buggy regex (measured), the "reproduces the input" property test asserts a tautology, the corpus test's oracle is `FenceSpans` itself, the close-gate regressions re-implement the guard instead of calling `findMilestonesMissingVerdict`/`computeClose` (which `testfix.Repo` already makes cheap), and `fence.go`/atlas document M2's end state as if it had landed. Separately, the heading-finding *class* was not swept — three production sites still locate `## <heading>` fence-blind, with a live corpus instance.

## 1. Strengths

- **The policy fork is the right architecture, and it's pinned.** `UnterminatedPolicy` as a typed parameter with the per-consumer rationale in-line (`cmd/sdlc/internal/issue/fence.go:22-50`) is the decision PQ-1 demanded, and `TestSectionBody_UnterminatedFenceIsProse` (`fence_test.go:80-104`) pins the *price* — the over-segmentation — so a later reader can't "fix" it back into a gate-disarming bug. Naming the price in a test is unusual discipline and exactly right here.
- **`FenceSpans`' two-pass shape matches the constraint.** The unterminated policy genuinely cannot be decided before EOF; expressing that as "record, then unwind from `openedAt`" (`fence.go:65-95`) is the minimal correct structure, not a workaround.
- **The move direction and the `project` seam.** Scanner moved DOWN into `issue` (correct — `project` imports `issue`), and `project` states its inherited policy as a named const with a comment rather than inheriting silently (`internal/project/doc.go:99-102`). I verified the two state machines are equivalent under `UnterminatedIsFenced`; the `project` suite passes with no other edit, as claimed.
- **The six-site sweep is complete as specified.** All of `close.go:563`, `close.go:1717`, `structural.go:160`, `sizing.go:63`, `plan.go:26`, `close_test.go:440` route through `PlanSectionBody`, and both stale comments were updated.
- **`PlanSectionBody` as a named shortcut** rather than five call sites spelling `SectionBody(body, "Plan")` — small, but it keeps the "Plan" literal in one place, which `structural_drift_test.go` now points at correctly.

## 2. Critical findings

None. No correctness bug, panic path, or silent error swallowing found in the shipped scanner. `fenceMarker` is index-safe (`len(trimmed) < 3` guards `trimmed[0]`), `SectionBody("")` returns `("", false)`, and `openedAt` is only read when `marker != 0`, which implies it was written.

## 3. Important findings

**I1 — Two of five fence-form cases don't fail without the fix (`fence_test.go:61`, `fence_test.go:130`).**
Measured, by running both extractors over the same fixtures:

| case | old regex passes | new passes | discriminating |
| --- | --- | --- | --- |
| plain fence | no | yes | ✅ |
| tilde fence | no | yes | ✅ |
| wide fence holding a narrow line | no | yes | ✅ |
| **indented fence** | **yes** | yes | ❌ |
| closed fence then a REAL heading | yes | yes | ❌ (control) |

The indented-fence case quotes `  ## Quoted` — indented — so `^## ` never matched it under the old regex either. The Done-when names "an indented fence" as one of the forms that must work; as written the case proves nothing. Fix: put an *unindented* `## Quoted` inside the indented fence (legal CommonMark, and the actual risk).
`TestFenceSpans_ReproducesInput` (`fence_test.go:133`) asserts only `len(FenceSpans(lines, policy)) == len(lines)`, which `make([]bool, len(lines))` on the first line of `FenceSpans` makes unfalsifiable. Its doc claims "classifying lines must not lose or invent any" and that "SplitFences' byte-exact reassembly guarantee rests on this" — neither is asserted. Fix: assert the partition — `visited ∪ skipped == [0,len)`, disjoint, and `strings.Join(lines, "\n") == body`.

**I2 — The corpus test's oracle is the function under test (`fence_test.go:156`).**
`inside := FenceSpans(lines, UnterminatedIsProse)` selects the "real" headings, and `SectionBody` uses the identical predicate on the identical spans — so `ok` can never be false and a `FenceSpans` bug moves both sides together. It does pin one real regression (SectionBody silently switching to `UnterminatedIsFenced` would red it), which is worth keeping, but the Plan priced this row as the primary corpus guard and it is not one. Fix: use an *independent* oracle — the pre-change `(?ms)^## <h>\s*\n(.*?)(?:^## |\z)` regex — and assert that every difference is a heading `FenceSpans` classifies as fenced. That form is what caught the one real corpus difference during this review.

**I3 — The close-gate regressions test the ingredients, not the gate (`cmd/sdlc/planfence_test.go:38,57`).**
Done-when says "`sdlc close` refuses … and `findMilestonesMissingVerdict` sees milestones after one. A test drives both." The tests instead call `issue.PlanSectionBody` and then re-apply `PlanUncheckedRE` / `milestonePlanRE` themselves — the same two lines `close.go:563` and `close.go:1723` contain. Re-point `close.go:563` at a different extractor and these stay green. `close_test.go:539-668` already builds real repos via `testfix.Repo`, so calling `findMilestonesMissingVerdict(planWithFencedHeading, "999", path)` and asserting `ordered == [M1 M2]` is a few lines. Fix at minimum the milestone one; the unchecked one can go through `computeClose`.

**I4 — The heading-finding class was not swept; three production sites remain fence-blind (`ARCH-PURPOSE`, family `consumer-enumeration-incomplete`).**
The Spec enumerated the `PlanSectionRE` consumers by grep, but not the class the bug belongs to — "code in `cmd/sdlc` that locates a `## <heading>` without fence awareness":

- `close.go:301` `insertLogLine` — mitigated by a *last-match* heuristic added in #66, which is a second, weaker workaround for the exact problem `FenceSpans` now solves, and which fails for a fenced `## Log` positioned *after* the real one. Measured: 1 of 406 corpus files already has that shape.
- `setstatus.go:302` `logHasEntryToday` — takes the **first** `## Log` (disagreeing with `insertLogLine` in the same binary) and terminates on `strings.Index(tail, "\n## ")`. **Live instance:** `workshop/history/issues/000066-close-log-line-under-day-header.md` has its first `## Log` at line 22 inside a fence, real one at line 68 — the reopen guard reads the wrong region on a real issue file.
- `changecode.go:741` `stripEstimateForHash` — line-wise `## ` scan; its own comment reasons about RE2's lack of lookahead, i.e. it is the same problem, and it is a natural `FenceSpans` consumer today.

(`issue.go:530`'s structure peek is cosmetic — list it, don't fix it.) This is legitimately M2-sized, but the *enumeration* belongs in the Plan this round, not the round after.

**I5 — `SectionBody` duplicates `project.SectionLineBounds` (`ARCH-DRY`).**
`section.go:20-38` and `internal/project/doc.go:106-121` are now the same algorithm — find the first non-fenced `## <name>`, run to the next non-fenced `## ` — differing only in policy and return type. The issue's whole thesis is that a fourth scanner was the wrong move; this diff removed one duplicate and introduced another. Fix: one exported `SectionLineBounds(lines []string, heading string, policy UnterminatedPolicy) (start, end int, ok bool)` in `issue`, with `SectionBody` joining its range and `project` calling it with `projectFencePolicy`.

**I6 — `fence.go` and the atlas document M2's end state as landed (family `docs-claim-unlanded-state`).**
`fence.go:1` says "the ONE fenced-code-block scanner for cmd/sdlc (#211)" — at HEAD, `fencedCodeRE` (`structural.go:227`) and `SplitFences` (`structural.go:246`) are both still there and still on their own logic. `atlas/workflow/issue-lifecycle.md:147-148` lists `stripCodeFences → UnterminatedIsProse` and `SplitFences → UnterminatedIsFenced` in a table headed "the unterminated-fence policy is a parameter", implying both are wired to `UnterminatedPolicy`; neither is. The atlas is the always-current map, so a reader will grep and find three scanners. Fix: mark those two rows "M2 — pending" (or "policy today, not yet a parameter") and soften `fence.go:1` to "the scanner section extraction is built on; `stripCodeFences`/`SplitFences` rebase in M2".

## 4. Minor findings

- `close_test.go:237` and `:254` keep verbatim inline copies of the deleted `PlanSectionRE`, so `TestPlanUncheckedDetection*` now pin a shadow of removed code and give false coverage for the very path this issue fixed. Route them through `issue.PlanSectionBody`.
- Fenced `- [ ] …` / `- [x] Mx …` lines inside a quoted block in `## Plan` are now counted by `PlanUncheckedRE`, `milestonePlanRE` and `CountPlanItems` — more likely than before, since fenced content now survives into `planBody`. Fails safe (spurious refusal / spurious verdict demand), but the counters could filter on `FenceSpans` for free.
- `structural.go:23` now reads "checkPlan encodes 'Plan' in `PlanSectionBody`, so a rename there needs a matching **regex** edit" — there is no regex any more.
- `workshopMarkdown` (`fence_test.go:187`) says "a missing workshop dir is a downstream repo, not a failure" and swallows the walk error, but the caller `t.Fatalf`s below 100 files. Pick one posture (`t.Skip` on empty is the consistent choice).
- **Pre-existing, not from this diff:** `go test ./cmd/sdlc/...` is red — `TestFleetPlanHasAuthoritativeCorrectedCoreConceptInventory` fails because `workshop/plans/000200-sdlc-fleet-thread-inventory-plan.md` was archived to history in `dfeba9c` (an ancestor of the base). Worth fixing opportunistically: it means the package `planfence_test.go` lives in cannot be used as green evidence.

## 5. Test coverage notes

- `TestFenceMarker` is a genuinely good boundary table (indent 3 vs 4, width 2 vs 3, inline backticks, tilde) — keep as is.
- Verified by revert simulation: all three `planfence_test.go` assertions fail against the old regex with the exact numbers the Log claims. The claimed fix is real.
- Not covered: `\r\n` line endings (they work — `TrimSpace` on the closer's rest absorbs `\r` — but nothing pins it); a tab-indented fence (`fenceMarker` only trims spaces, so a tab-indented ```` ``` ```` is treated as a fence at indent 0 — CommonMark would call it 4 columns); a heading as the final line with no trailing newline (now found, where the old regex returned not-found — a small improvement worth a line).
- The one behavioral difference I measured across the corpus is `000145-…-plan.md`, where `## Spec`/`## Plan`/`## Done when` correctly stop being found. Plan files aren't gated by `sdlc issue validate`, which is why the Log's byte-identical validate result held while the corpus did change — worth a sentence in the Log so the next reader knows the validate check and the corpus check answer different questions.

## 6. Architectural notes

- **ARCH-DRY — flag.** See I5 (duplicated section-bounds loop) and the Minor on `close_test.go`'s inline regex copies.
- **ARCH-PURE — pass, with a note.** `FenceSpans`, `fenceMarker`, `SectionBody` are pure and unit-tested directly; `project` injects its policy at the boundary. The corpus tests do read-only filesystem IO from an otherwise-pure package and hard-code `../../../../workshop`; acceptable as a fixture source, but it couples `internal/issue`'s test binary to repo layout.
- **ARCH-PURPOSE — flag.** See I4. The named instance (`SectionBody`) and its enumerable sibling (`PlanSectionRE`) are fixed; the class ("heading-shaped line inside a fence read as structure") has three more production members, one with a measured live instance.
- **ARCH-MOCK — pass.** No new external dependency; the `gitx` seam is untouched. Note that I3's fix uses the existing `testfix.Repo` fixture rather than adding a new double.
- **ARCH-CONSTRAINTS — pass.** `SectionBody` re-splits and re-scans the whole body per call and `CheckStructural` calls it 4–5× per issue, but issue files are ≤100 KB and this runs a handful of times per command; the corpus test over 2875 headings runs in 0.42s. No envelope needed.
- **ARCH-SECURE — pass.** Issue bodies are hand-edited, possibly written by older versions — untrusted by provenance. The chosen policy degrades in the visible direction (a malformed fence reveals *more* structure, never less), which is the correct failure mode for inputs that feed a refusal gate. No panic path on empty/truncated input, no credentials in scope.
- **For M2:** the `SplitFences` line-anchoring decision is now the load-bearing one. `SplitFences` uses substring `strings.Index("```")` while `fenceMarker` requires line-start with ≤3 spaces of indent, so rebasing flips 4-space-indented triple backticks from fenced to prose and `migrate` starts rewriting refs inside indented code blocks. `FenceSpans` returns per-line flags, so byte-exact reassembly has to be rebuilt from line offsets deliberately — build the offset table from the original string rather than re-joining with `"\n"`, or a file whose final line lacks a newline will round-trip wrong.

## 7. Plan revision recommendations

Add a `## Revisions` entry to `workshop/issues/000211-sectionbody-truncates-at-a-fenced-heading.md` covering:

1. **Done-when bullet 1 is contradicted by the chosen policy.** It lists "an unterminated fence" among the forms where the section "is extracted in full"; `UnterminatedIsProse` deliberately over-segments instead, as `TestSectionBody_UnterminatedFenceIsProse` documents. Reword to name the four forms extracted in full and state the unterminated case separately as "never hides a later `##`".
2. **Plan row 4 over-claims.** "visited + skipped lines reproduce the input" is not asserted anywhere (see I1); either deliver the partition assertion or drop the clause.
3. **Add an M2 row for the heading-finder enumeration** (I4), naming `close.go:301`, `setstatus.go:302`, `changecode.go:741` and the measured live instance in `000066`, so the class is written down rather than rediscovered.
4. **Note the validate-vs-corpus distinction** in `## Log`: `sdlc issue validate --all` was byte-identical *and* one corpus file's extraction changed (correctly) — the two checks cover different file sets, and the Done-when's "predicted: none" refers to issue verdicts only.

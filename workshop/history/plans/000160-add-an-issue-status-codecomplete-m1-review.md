# Boundary Review — ariadne#160 (milestone M1)

| field | value |
|-------|-------|
| issue | 160 — add an issue status: codecomplete |
| repo | ariadne |
| issue file | workshop/issues/000160-add-an-issue-status-codecomplete.md |
| boundary | milestone M1 |
| milestone | M1 |
| window | 2110aaab811f23641f57c55628b2bda362cce73c^..HEAD |
| command | sdlc milestone-close --issue 160 --milestone M1 |
| reviewer | claude |
| timestamp | 2026-07-02T16:41:14-07:00 |
| verdict | FIX-THEN-SHIP |

## Review

I have everything I need. Let me verify my key finding once more by confirming the derived help render would show `working → codecomplete` as legal (contradicting the new Guard 1b refusal).

`LegalTransitions("working")` iterates the lifecycle for `from=="working"` → `blocked, codecomplete, wontfix, punt`, so `RenderLifecycleHelp`'s LEGAL TRANSITIONS block renders `working → …, codecomplete, …` as legal — yet `set-status` Guard 1b refuses exactly that edge, and the FLAGS line says only `done` is excluded. Confirmed contradiction. Here is my review.

```verdict
verdict: FIX-THEN-SHIP
confidence: high
```

**Summary.** M1 is a clean, well-scoped vocabulary + enforcement change. The `codecomplete` status is added to the single CUE source, the committed `pkg/vocab/issue.json` is **byte-identical** to what `issue.cue` regenerates (verified — the single-source/ARCH-DRY property holds), the lifecycle edges and the extended `actual_hours!` guard are correct, and set-status's new `→ codecomplete` refusal (Guard 1b) plus the reworked `→ done` guard are implemented and tested. All M1 logic tests pass. The one thing blocking a clean SHIP is non-blocking at the gate: the `set-status` help text was not updated for the new `→ codecomplete` refusal, so `sdlc issue set-status --help` now actively contradicts both its own derived LEGAL TRANSITIONS block and the tool's real behavior. Everything M1 claimed to deliver is delivered; close/merge/push wiring is correctly deferred to M2/M3.

**Note on test execution:** I could not run the full `cmd/sdlc` suite — the live parent `sdlc milestone-close` (pid 38205) that spawned this review holds `.git/sdlc.lock`, so every mutating-command test that goes through `buildRoot()` (e.g. `TestSetStatusAlias_BothPathsMutate`) blocks until the review completes. That timeout is a harness artifact, **not** an M1 defect. I verified all pure/logic tests (`checkTransitionGuards`, `applyStatus`, `runSetStatus`, all `pkg/vocab`) pass, and regenerated the CUE→JSON to confirm sync.

### 1. Strengths
- **True single-source, verified.** `go run ./cmd/vocabulary export --noun issue` is byte-identical to the committed `pkg/vocab/issue.json` — the enum isn't hand-edited anywhere; consumers derive (`vocab.go` predicates, `RenderLifecycleHelp`). (ARCH-DRY / ARCH-PURPOSE pass.)
- **Guard ordering is correct.** Guard 0 (lifecycle graph) runs first, so `working → done` reads as "illegal transition" while `codecomplete → done` is a legal edge that Guard 1 then value-refuses — the two messages are distinct and correct (`setstatus.go:233-267`).
- **Test coverage tracks the model change precisely.** `RefusesDone` now exercises `codecomplete → done` + asserts `working → done` is illegal; new `RefusesCodecomplete`; `NormalTransitions` gained the three clean codecomplete-reopen/abandon/defer edges — importantly pinning that codecomplete-reopen does *not* trip Guard 2's done-reopen log requirement (`setstatus_test.go:129-134`).
- **Conformance laws auto-cover the new status** (reachable/escapable/documented-value) — no maintained list to drift (`conformance_test.go`, `issue.json` laws block).

### 2. Critical findings
None.

### 3. Important findings

**`cmd/sdlc/helptext/set-status.md:49-50` (and `:17-22`) — help not updated for the new `→ codecomplete` refusal; it now contradicts the tool's own output.** M1 introduced Guard 1b refusing `→ codecomplete`, but the help still says the `<status>` positional accepts "any status from STATUSES above except `done`" and has no `→ codecomplete` guard section. Meanwhile the derived `{{LIFECYCLE}}` block renders `working → …, codecomplete, …` as a *legal* transition (`RenderLifecycleHelp` via `LegalTransitions("working")`). So a user reading `--help` is told `codecomplete` is both a legal target and a settable status, but `sdlc issue set-status --issue N codecomplete` is refused. The established pattern for `done` (an explicit "→ done Always refused" section + the FLAGS carve-out) was not mirrored for `codecomplete`. This is the ARCH-PURPOSE shadow-sweep catch: a hand-maintained restatement of the refusal policy that did not derive the new rule.
  - *Fix:* in the FLAGS line, "except `done` **and `codecomplete`** (both are refused — use `sdlc close`)"; add a parallel "`→ codecomplete` Always refused. Use `sdlc close`." block alongside the `→ done` section (~`:17-22`). Non-blocking at the gate, but fix before merge since close is now the sole gate.

### 4. Minor findings
- `atlas/workflow/issue-lifecycle.md` / `vocabulary.md` and the Guard 1 message ("`sdlc merge` … codecomplete → done") describe close→codecomplete and merge→done, which are only *coded* in M2/M3 — at M1 `close.go:486` still writes `done` and merge/push don't flip. Acceptable because the branch merges atomically, but the docs/messages lead the code within the branch; ensure M2/M3 land before merge (they will).
- `cmd/vocabulary/validate_test.go:41,61,94` fixtures hardcode the 6-value status disjunction (no `codecomplete`). They're synthetic parser-input fixtures (still valid tests), but now unrealistic vs the real 7-value model — optional refresh.

### 5. Test coverage notes
- Logic coverage for M1 is solid (guards + predicates + conformance). Gap: nothing pins that the hand-written set-status help prose reflects the refusal set — which is exactly why finding #3 slipped; prose is hard to unit-test, so a `helptext` string assertion ("codecomplete" appears in a refusal context) would be the cheap guard if desired.
- I was unable to run the e2e/mutating-command tests (parent holds the repo lock). M1 doesn't touch `close.go`/`merge.go`/`push.go`, so those paths still see `done` behavior and should be unaffected; the implementor's Log claims the suite was green pre-lock. Low residual risk, flagged for transparency.

### 6. Architectural notes for upcoming work
- **ARCH-DRY: pass** — model single-sourced and verified in-sync; Guard 1 and Guard 1b are structurally similar but are two genuinely distinct value-specific policies with distinct messages, so not worth collapsing.
- **ARCH-PURE: pass** — `issue.cue` is pure data; `vocab` predicates and `checkTransitionGuards` are pure (string in, error out); all verified tests run without IO.
- **ARCH-PURPOSE: one flag (finding #3)** — the model is enforced (Guard 1b) and the JSON derives, but the set-status help is a hand-maintained consumer that didn't derive the new refusal. M2 will edit `set-status.md`/`close.md` anyway for the close→codecomplete reframe — fold the codecomplete-refusal doc there, or better, in M1's fix so the boundary isn't crossed with self-contradicting help.
- For M3: the `reviewed-head-unchanged` guard name is declared in the model but has no implementation yet (expected). When wiring it, note that `close.go` writes status *unconditionally* (no transition-table consult), so the model's legality of an edge and set-status's willingness to write it are independent axes — keep that in mind for the merge flip.

### 7. Plan revision recommendations
None required — the plan still matches the code. M1's own Log already records the set-status-enforcement pull-forward (M2 Step 3 → M1) and the shadow-sweep; the plan and code agree. The only follow-through is folding the `set-status.md` help fix (finding #3) into M2, which the plan's M2 "Help + atlas" step already scopes — just ensure it explicitly adds the `→ codecomplete` refusal note, not only the close→codecomplete reframe.

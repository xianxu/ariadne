# Boundary Review — ariadne#284 (whole-issue close)

| field | value |
|-------|-------|
| issue | 284 — Claims: multi-claim, publish, handoff, takeover |
| repo | ariadne |
| issue file | workshop/issues/000284-claims-handoff-takeover.md |
| boundary | whole-issue close |
| milestone | — |
| window | 01079ab94940e6d1df39eed60bfb48b148105aaa..8a53d19ce29869fbddbd2afa8a38d0e9f3adb837 |
| command | sdlc close --issue 284 |
| reviewer | claude |
| timestamp | 2026-10-07T21:58:59-07:00 |
| verdict | SHIP |

## Review

```verdict
verdict: SHIP
confidence: high
```

Round 12 of the #284 close gate: every finding still open is now fixed, and nothing new blocks the close. The last fix commit (`2e95e8f7`) does three things:
- **BR-42:** caps the claim-age history scan at `tracker.ClaimHistoryLimit` (5000 commits).
- **BR-39:** adds a regression test for the parse-error kill. It writes a 17 MB commit message, which makes the scanner fail with ErrTooLong. Without `finish(true)`, git blocks on a full pipe and the 30 s watchdog fails the test.
- **BR-41:** adds a plan revision recording that a takeover is published as an ordinary `claim` operation (`tracker.OpClaim`), with no separate token.

I checked the older Minors (BR-1..6, BR-19, BR-29..31) against the head code, and each is fixed there. The build passes, and `go test` passes for `internal/tracker` and `internal/issue`. One new Minor: the project file's checkboxes are behind its own milestone blocks.

1. **Strengths**
   - `claimtimes.go`: the history parser `ClaimTimes(io.Reader, …)` is pure and separate from the IO wrapper. The wrapper stops git before returning a read error (`claimtimes.go:113-116`), and a full read reports git's exit status.
   - `gitx.RemoteTrackingRef` (`trunkfile.go:197`) is now the only place that builds a remote-tracking ref name. `handoff.go`, `transferguard.go`, `landing.go` and `trackingRef()` all use it.
   - `handoffNoteCommitted` (`handoff.go:130`) looks only in the branch's unpushed commits for the note, so a rerun recognises its own earlier attempt rather than any matching text.
   - Existence checks now distinguish present, absent and error: `env.has` / `detailsAt` / `restoreToHead` in `republish.go`, and the `gitTest` checks in `handoff.go:272,329`.
   - `ChangeCards` removes duplicate ids itself and checks all the new card contents together through `validateReplacements` (`changecards.go:35-43,87`).

2. **Critical findings:** none.

3. **Important findings:** none.

4. **Minor findings**
   - `workshop/projects/claimant-ownership.md:50-52`: the M2, M3 and M4 checkboxes are still unticked, though each detail block records `closed:` and `actual:`. The M4 block also still says the scan "stops once every owned card is resolved" and doesn't mention the 5000-commit cap.

5. **Test coverage**
   - The BR-39 test pins the hang. The cap itself has no test (a card past the limit shows no age). That is cheap to leave untested, since the only outcome is an empty age string.

6. **Architecture**
   - **ARCH-DRY:** pass (ref building and id joining each have one helper).
   - **ARCH-PURE:** pass (`ClaimTimes`, `claimSetDecision`, `republishDecision` and `claimAge` are pure).
   - **ARCH-PURPOSE:** pass.
   - **ARCH-MOCK:** pass. Tests run against real git fixtures through the existing seams.
   - **ARCH-CONSTRAINTS:** pass. The history scan is now bounded.
   - **ARCH-SECURE:** pass. Oversized or malformed history records are skipped or returned as errors, never treated as evidence.
   - **ARCH-ORDER:** pass. Lost-response rerun and takeover compare-and-swap ordering are covered by the claim-set and handoff tests.
   - **ARCH-FUNERAL:** pass. Set-aside copies are removed, and the release record is visible in `state`.
   - One thing for later work: the same 5000-commit cap could serve as a pattern for other history readers.

7. **Plan revision recommendations:** none. The 2026-10-07 "as built" revision matches the code.

```findings
dispose:
  - id: BR-1
    disposition: addressed
    note: |
      fastForwardRest (claim.go:283-294) surfaces rev-parse and merge-base errors with their actual cause.
  - id: BR-2
    disposition: addressed
    note: |
      issue.JoinRefs backs claimRefs/claimArg/cardsMessage; ChangeCards calls snapshot.validateReplacements over the whole set.
  - id: BR-3
    disposition: addressed
    note: |
      ChangeCards dedupes ids itself (changecards.go:35-41).
  - id: BR-4
    disposition: addressed
    note: |
      --adopt folded into claim (M3); a handed-off member refuses with "claim it alone (sdlc claim --issue N)".
  - id: BR-5
    disposition: addressed
    note: |
      M lines nested under the 284 row; actuals are measured values in the detail blocks.
  - id: BR-6
    disposition: addressed
    note: |
      TestClaimSetLostResponseAndTakeover drives a lost set response and the rerun.
  - id: BR-19
    disposition: addressed
    note: |
      republish.go uses env.has with errors (detailsAt, restoreToHead); merge-base failure returns an error.
  - id: BR-29
    disposition: addressed
    note: |
      gitx.RemoteTrackingRef is used at all named sites (handoff, transferguard, landing, trackingRef).
  - id: BR-30
    disposition: addressed
    note: |
      handoffNoteCommitted searches only the branch's unpushed tail for the note commit subject.
  - id: BR-31
    disposition: addressed
    note: |
      handoff.go:272 and :329 use env.gitTest and stop on error.
  - id: BR-39
    disposition: addressed
    note: |
      finish(true) on readErr; TestRepositoryClaimTimesStopsGitOnAParseError hangs without it (17 MB record past the 16 MB scanner cap).
  - id: BR-41
    disposition: addressed
    note: |
      Plan revision 2026-10-07 "as built" records that a takeover is minted as claim.
  - id: BR-42
    disposition: addressed
    note: |
      HistoryStream takes --max-count; ClaimTimes bounded at ClaimHistoryLimit=5000.
findings:
  - id: new
    severity: Minor
    family: docs-sweep-missing
    title: |
      Project file leaves M2-M4 checkboxes unticked while their blocks record closed/actual
    detail: |
      workshop/projects/claimant-ownership.md:50-52 are unticked though each detail block has closed: 2026-10-07; the M4 block also omits the 5000-commit claim-age bound. Rule: a milestone-close updates both the checkbox and its detail block together.
```

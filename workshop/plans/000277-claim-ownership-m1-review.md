# Boundary Review — ariadne#277 (milestone M1)

| field | value |
|-------|-------|
| issue | 277 — Record claimant ownership atomically with issue reservation |
| repo | ariadne |
| issue file | workshop/issues/000277-claim-ownership.md |
| boundary | milestone M1 |
| milestone | M1 |
| window | bb6221d84ecd5ba27a1336cb78515ecdb2e5d602..61af891b21f6ca0f6027527fbe0de64054050697 |
| command | sdlc milestone-close --issue 277 --milestone M1 |
| reviewer | claude |
| timestamp | 2026-10-01T14:01:43-07:00 |
| verdict | FIX-THEN-SHIP |

## Review

```verdict
verdict: FIX-THEN-SHIP
confidence: medium
```

M1 does what it set out to do. Claim now writes a `claimant` record in the same blob-OID compare-and-swap that flips open → working. The record is a strict, fail-closed vocabulary kind, it is mirrored into the details like every other card field, and a pure `MatchClaimant` decides ownership on repository + machine + worktree. The pure core is well separated from the identity IO seam, and the tests cover parsing (old and hand-edited input), matching, relocation, the mirror, and the two-clone race. I found nothing Critical.

Two things should be fixed cheaply before the boundary:
- **The race test now expects the wrong message from the loser.** If the losing clone reads the card after the winner has published, it now gets a "foreign owner" refusal. That message contains neither "not open" nor "changed while claiming", which are the only two strings the test accepts. The help text promises the same two messages, so it has the same drift.
- **The plan and the code disagree on `MachineFingerprint`.** The plan's Core concepts table and test list say `MachineFingerprint(raw, key)`, with a test showing that a different key gives a different value. The code takes `MachineFingerprint(raw)` with a fixed domain prefix.

### 1. Strengths
- `internal/issue/claimant.go:57-89`: `parseClaimant` fails closed on every hand-edit case:
  - exact key set, with only `workspace` optional;
  - `!!str` scalars only, one line, not empty;
  - no flow style;
  - `machine` must be a 32-hex fingerprint, so a pasted raw UUID is rejected.
  
  The table test at `claimant_test.go:46-72` covers each case (ARCH-SECURE pass).
- `SetCardClaimant` replaces only its own span and re-validates through `ParseCard`. The byte-preservation test pins this.
- `sameCardValue` (`card.go:330`) compares mappings by their decoded content. That one change was all the mirror needed, and `TestClaimantMirrorsIntoDetails` covers adding the claimant, replacing it, removing it, and refusing a hand edit.
- `slotLabel` (`claimant.go:51`) correctly avoids labelling a plain clone `repo:0`. That bug would have been easy to ship, and the commit that added it records why.
- The race test's "loser's worktree is absent from tracker history" check (`claimremote_test.go:196`) is a good check that the losing claim published nothing.

### 2. Critical findings
None.

### 3. Important findings
- **`cmd/sdlc/claimremote_test.go:177`: the race test rejects one legitimate loser outcome.**
  - A loser whose `Snapshot()` runs after the winner has published reaches `ownedBy`, finds a foreign owner, and returns `"remote issue #9 is working, claimed by …"` (`claimdecision.go:76`).
  - Before this change, that path returned "not open". Now it hits `t.Errorf("loser was not a status/CAS refusal")`.
  - So the test is green only when both clones read the card before either publishes. It has sampled one ordering, not proven the race (ARCH-ORDER).
  - **Fix:** accept `"claimed by"` as a valid refusal. Better, make it deterministic: add an in-process test that runs `runClaim` from a second identity after the first claim has landed. That ordering is partly covered at `claimremote_test.go:90`, but not as a race loser.
  - `helptext/claim.md` should also say that a loser can see the foreign-owner refusal.
- **Plan drift: `MachineFingerprint(raw, key)` against `MachineFingerprint(raw)`.**
  - The plan's Core concepts table and M1 step promise a key parameter and a "different key gives a different value" test. The code has a constant `"ariadne-claimant\x00"` prefix, and the test checks a different *raw* value instead.
  - The behaviour is fine. The prefix gives domain separation, and the 128-bit raw ID can't be brute-forced.
  - **Fix:** append a `## Revisions` entry to the plan saying the key is a fixed domain constant, not a parameter, so the plan stops claiming an API the code doesn't have. The atlas and help text already say "keyed SHA-256", which is fine with a fixed key.

### 4. Minor findings
- `claim.go:136-143`: with `--dry-run`, an owner's repeat claim still calls `refreshLocalMirror` and writes the details file. That path should skip the refresh and print "would …".
- `claimdecision.go:78`: the unknown-owner refusal points to `sdlc claim --adopt`, which won't exist until M2. Acceptable inside the milestone, but don't use an M1-only binary on the live tracker.
- No rollout note yet:
  - M1 is already the step that writes claimant cards to the shared `issue-tracker` branch.
  - Any claim made with this branch's binary starts the flag day for every stale binary in the fleet, and the atlas has no rollout note yet (the plan puts it in M2).
  - Suggestion: add the flag-day paragraph to the atlas section now, or avoid claiming with this binary until the work lands.
- Plan says the race test is "skipped with a reason where the host's machine ID is unreadable", but no skip exists. On such a host every claim test fails instead. GitHub's ubuntu runners have `/etc/machine-id`, and CI doesn't run `go test`, so it's low impact.
- `slotLabel` (`claimant.go:59`) re-derives the slot layout (`worktree/<repo>-slot*/<repo>`) in `cmd/sdlc`. That layout belongs to `pkg/workspace` (ARCH-DRY). A `workspace.HasSlots(id)` helper there would keep one source of truth.
- `claimremote_test.go:177`: stray double space in `loser = res.dir;  !strings…`, and `gofmt` would rewrite that line.
- `MachineFingerprint` uses `fmt.Sprintf("%x", …)`; `hex.EncodeToString` would be more idiomatic.

### 5. Test coverage notes
- Pure entities are tested without IO. The parsers for `ioreg` and `machine-id` take fixture text, which is good.
- The identity seam is a package variable (`withClaimant`). That is safe only because the `cmd/sdlc` tests run serially, which `merge_e2e_test.go:125` documents. Any future `t.Parallel()` here would race on it.
- Missing: a `claimDecision` table test with a non-nil `me` across (status × claimant present × match). The Mine, Foreign and Unknown branches are exercised only through `runClaim` integration tests, and Unknown (a legacy working card) isn't exercised at all in M1. M2's adopt table is planned to cover it; consider adding the Unknown refusal case now, since M1 shipped that branch.

### 6. Architecture notes (ARCH-*)
- **ARCH-DRY:** passes, apart from the `slotLabel` layout duplication (Minor).
- **ARCH-PURE:** passes. Decisions sit in `internal/issue`; IO is confined to `cmd/sdlc/claimant.go`.
- **ARCH-PURPOSE:** passes for M1. Record and publish are delivered. Enforcement at the gates is honestly scoped to M2, not deferred as a "follow-up".
- **ARCH-MOCK:** passes, with a note.
  - `ioreg` and `scutil` are called directly, but parsing is pure and tested, and there is a live host test.
  - The seam is the `claimantIdentity` variable, which is acceptable for a read-only, stateless dependency.
- **ARCH-CONSTRAINTS:** passes. Claim now adds a few subprocess calls (`git config`, `ioreg`, `scutil`, workspace resolution), which is negligible for a one-off command.
- **ARCH-SECURE:** passes.
  - The raw machine ID never leaves the process.
  - The published exposure (user.name, machine name, path) is documented.
  - Input read from the card is parsed into a typed value and fails closed.
- **ARCH-ORDER:** flagged. Ownership and status share one CAS, which is right. But the race test's oracle can see only one interleaving and rejects the other legitimate one (Important above).
- **ARCH-FUNERAL:** passes. The claimant is replaced, never appended, and its size is bounded per card.
- **For M2:**
  - Put `requireOwnership` behind one helper that reuses `ownedBy`/`describeClaimant`, so the wording of the four gates' refusals isn't restated (ARCH-DRY).
  - The `issue set-status working` bypass is still open until M2 lands.

### 7. Plan revision recommendations
- Revisions: "`MachineFingerprint(raw)` takes no key. The key is a fixed domain-separation constant (`ariadne-claimant\x00`), and the test asserts determinism, 32-hex output, and separation between different raw IDs."
- Revisions: "The claim race loser may also see the foreign-owner refusal ('claimed by …'). Both the race-test oracle and the help text accept it."
- Optional: "The race-test skip on an unreadable machine ID was dropped (or implemented)", whichever matches the fix.

```findings
findings:
  - id: new
    severity: Important
    family: test-oracle-single-interleaving
    title: |
      Claim race test rejects the foreign-owner refusal a late-reading loser now gets
    detail: |
      A loser whose snapshot follows the winner's publish hits ownedBy, which returns "is working, claimed by ...". That message contains neither "not open" nor "changed while claiming", so the test is green only when both clones read before either publishes. Accept "claimed by" (or add a deterministic in-process loser test), and update the claim.md loser wording to match.
  - id: new
    severity: Important
    family: plan-code-contract-drift
    title: |
      Plan promises MachineFingerprint(raw, key) plus a key test; code has MachineFingerprint(raw)
    detail: |
      The Core concepts table and the M1 step describe a key parameter and a "different key gives a different value" test. The code uses a fixed domain prefix and tests separation between different raw IDs. The behaviour is fine; add a plan Revisions entry so the plan matches the code.
  - id: new
    severity: Minor
    family: dry-run-writes
    title: |
      An owner's repeat claim with --dry-run still refreshes and writes the local details mirror
  - id: new
    severity: Minor
    family: rollout-doc-timing
    title: |
      M1 already writes claimant cards to the shared tracker, but the flag-day rollout note is deferred to M2
    detail: |
      Any claim made with this branch's binary breaks stale fleet binaries. Add the atlas rollout paragraph now, or avoid claiming with this binary until the work lands.
  - id: new
    severity: Minor
    family: plan-code-contract-drift
    title: |
      The race-test skip on an unreadable host machine ID promised by the plan is not implemented
  - id: new
    severity: Minor
    family: layout-knowledge-outside-owner
    title: |
      slotLabel re-derives the slot directory layout in cmd/sdlc instead of asking pkg/workspace
  - id: new
    severity: Minor
    family: missing-branch-unit-test
    title: |
      No claimDecision table test with a non-nil claimant; the Unknown (legacy working card) refusal is untested in M1
```

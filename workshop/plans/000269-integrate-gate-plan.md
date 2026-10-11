# #269 — integrate gate: previous-noon freshness at review boundaries

**Goal:** before a boundary review (`sdlc close`, `sdlc milestone-close`, including
`--fixed-then-ship`), refuse when the branch is missing a code change that landed
on the configured main before the most recent noon, so the review and the suite
describe a tree that can ship. One freshness definition, shared with #303.

**Why now:** #304/#183/#320 made integrating main cheap. A merge or rebase no longer
widens a review window or forces a full re-close. So the rule can now be strict
without multiplying review cost (project requirement 6).

## Decisions

- **The rule (project req. 6):** the branch must contain every first-parent commit
  of fresh main whose committer date is at or before the most recent noon in the
  repository's timezone. Equivalently, `required` = `git rev-list -n1 --first-parent
  --before=<cutoff> <main>`, and the gate passes iff `required` is an ancestor of
  HEAD. First parent is the landing order: a PR merge commit's committer date is
  when it landed. For a fast-forwarded push it is the commit's own date. That makes
  the rule slightly stricter for those landings, never looser, and it's documented.
- **Docs-only staleness passes (the issue spec):** when `required` is not in HEAD,
  classify `diff(merge-base(HEAD, required), required)`'s paths with
  `publishGateHasCodeSurface`, the publish gate's code predicate (cmd/ markdown is
  embedded, so it is code; ARCH-DRY). No code surface → pass with an info line.
- **Tri-state probe (ARCH-ORDER, issue spec):** `Fresh | DocsOnly | Stale |
  Inconclusive`. A fetch failure, a missing main ref, or a git error is
  `Inconclusive`. Close prints a loud "could not check" warning and continues. It
  never prints the up-to-date line. It does not refuse, because an offline close is
  legitimate and the publish path re-fetches anyway. The close log line does not
  claim freshness.
- **Which main:** the same fresh main the review window uses. In a tracker repo
  that's `reviewMainRef` (one cached fetch per process, so the gate and the window
  agree on main). A stale note from it means Inconclusive. In a legacy repo, the
  gate runs `git fetch --quiet <remote> <branch>` for `gitx.TrunkRef()`'s remote,
  and a failure is Inconclusive.
- **Timezone (repo-configured, for multi-operator repos):** `.sdlc/freshness.json`
  `{"version":1,"timezone":"America/Los_Angeles"}`, tracked beside `fleet.json`.
  Absent → the machine's local zone, and the refusal names the zone it used. An
  unreadable file or an unknown zone refuses, naming the file; a wrong zone would
  silently move the cutoff. Ariadne ships the file with `America/Los_Angeles`.
- **Next action by push state:**
  - The branch has a remote-tracking counterpart (`refs/remotes/<remote>/<branch>`
    or `@{upstream}`) → `git merge <main>`. A rebase would need a force-push, which
    fights merge's local/remote/PR head check.
  - Otherwise → `git rebase <main>`.
  - Both: commit local edits first, re-run the suite, re-run the same close command.
  - A test runs the named remedy from the refused state and then passes (lessons:
    every remedy runs from the refused state).
- **Both boundaries.** The project rule says "at a review boundary". After #304 a
  milestone window is the interdiff replayed onto main, so integrating costs
  nothing there either. `milestone-close` gets the same gate and flag. The issue
  floated a warning-only milestone check; #304 removed its reason (stale
  `Review-Window` SHAs are A2, absorbed by #304).
- **Bypass:** `--no-integrate` on `close` and `milestone-close`, wired through
  `closeFlags.skip("integrate")`, so `--force` covers it. Registered in gatesig's
  catalog (AckPat/RefusalPat), and the refusal names the flag.
- **Placement:** in `computeClose`, after the re-close guard and before the window
  is planned, so a refusal costs no review. Dry-run runs it too: it's read-only
  apart from the fetch.

## Shape (ARCH-PURE)

New package `cmd/sdlc/internal/freshness` (#303 calls the same API at claim/start-plan):

- `Cutoff(now time.Time, loc *time.Location) time.Time`: the most recent noon at or
  before `now`. Pure. Tested across midnight, exactly noon, and DST days.
- `LoadLocation(root string) (*time.Location, string, error)`: reads
  `.sdlc/freshness.json` and returns the zone plus its source label.
- `Check(git Git, mainRef, head string, cutoff time.Time, isCode func([]string) bool) Result`:
  `Git` is a minimal `func(args ...string) (string, error)` seam. Production passes
  a gitx-backed runner; tests pass real git in a fixture repo with a bare remote,
  so there's no function-call mock (ARCH-MOCK). `Result{State, Required, Cutoff,
  Code []string, Docs int, Why string}`.

Shell: `cmd/sdlc/integrategate.go` holds `integrateGate(ctx, stderr, f)`. It resolves
main and fetches, loads the zone, takes `now` from a package var seam, calls `Check`,
and renders pass, info, warn or refusal. Rendering is a pure `formatIntegrate*`.

## ARCH notes

- ARCH-FUNERAL: creates nothing durable. The fetch updates the remote-tracking ref,
  which the publish path already does. `.sdlc/freshness.json` is static
  configuration, one file per repo.
- ARCH-CONSTRAINTS: one `git fetch` (cached with the window's in tracker repos), plus
  `rev-list -n1`, `merge-base`, `diff --name-only`. That's seconds at most, on a
  command that then runs a multi-minute review.
- ARCH-SECURE: trusts the configured remote and the git committer dates, the same
  trust the review window already places in them. The timezone file is repo
  content, reviewed like code.
- ARCH-ORDER: the gate reads, then the review dispatches. A main that moves between
  the gate and the review is caught at publish (#304's publish gate), not here.
  That's acceptable, because the cutoff is a past instant: a later push can't move
  `required` backwards.

## Steps

- [ ] `internal/freshness`: Cutoff + LoadLocation + Check, with unit tests (Cutoff
  table incl. DST; LoadLocation absent, valid, bad) and real-git Check tests:
  fresh; main moved after the cutoff only (fresh); docs-only before the cutoff;
  code before the cutoff (stale); a missing main ref (inconclusive).
- [ ] `integrategate.go` + wiring into `computeClose`, `--no-integrate` on close and
  milestone-close, and `skip("integrate")`. Tests drive `computeClose` against a
  fixture with a bare remote: refused with the rebase hint when unpushed and the
  merge hint when pushed; after running the named remedy it passes; a fetch
  failure prints "could not check" and proceeds; `--no-integrate` prints the ack.
- [ ] gatesig catalog entry; help text (`close.md` and `milestone-close.md` gate
  lists); recovery catalog preconditions for close and milestone-close; atlas
  `workflow/sdlc-binary.md`; `.sdlc/freshness.json` for ariadne.
- [ ] Targeted tests, then a full `make test` before close.

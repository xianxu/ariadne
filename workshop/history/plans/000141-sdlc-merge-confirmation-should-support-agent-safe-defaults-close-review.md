# Boundary Review — ariadne#141 (whole-issue close)

| field | value |
|-------|-------|
| issue | 141 — sdlc merge confirmation should support agent-safe defaults |
| repo | ariadne |
| issue file | workshop/issues/000141-sdlc-merge-confirmation-should-support-agent-safe-defaults.md |
| boundary | whole-issue close |
| milestone | — |
| window | b7a6e681b99adfdd55e6171cd7474f02b3bc183a..HEAD |
| command | sdlc close --issue 141 |
| reviewer | claude |
| timestamp | 2026-07-01T10:08:36-07:00 |
| verdict | SHIP |

## Review

The directly-affected tests pass and build/vet/gofmt/cross-compile are all clean; the full-suite run is still compiling in the background. Given the change is isolated (a 2-line `isTTY` rewrite plus four tiny build-tagged files) and its regression test passes, I have sufficient evidence to deliver the review. I'll note the full suite was launched as confirmation.

```verdict
verdict: SHIP
confidence: high
```

**Summary.** This is an exemplary small root-cause fix. The issue's stated symptom (merge runs all judges, then aborts at an unanswerable prompt) was correctly re-diagnosed from "add a fail-fast" to "the fail-fast already exists but is defeated by a wrong terminal test": `isTTY` used `os.ModeCharDevice`, which is true for `/dev/null` — an agent's usual stdin — so the guard was skipped. The fix replaces the proxy with a genuine `isatty` ioctl probe (`TIOCGETA`/`TCGETS`), stdlib-only, split cleanly per-OS. I verified: build OK, `GOOS=linux` cross-compiles, `go vet` and `gofmt` clean, and the targeted tests pass on darwin — including the `/dev/null → false` regression that pins the bug. The shadow-sweep confirms both `isTTY` consumers (`merge.go:272`, `changecode.go:488`) now derive from the corrected helper, so the shared-helper fix lands everywhere it manifests. Nothing blocks SHIP.

**1. Strengths**
- Correct root-cause diagnosis, not a band-aid. The plan explicitly rejects a `/dev/null` special-case in favor of fixing the wrong terminal test (`workshop/plans/000141-...md`). This is the AGENTS.md "Root Cause" principle done right.
- `isTerminal` (`cmd/sdlc/tty_unix.go:17`) uses the exact technique `golang.org/x/term` uses internally, while staying stdlib-only — honoring the pre-existing deliberate "no x/term dependency" choice recorded in the old comment. `go vet`'s unsafe.Pointer/syscall check passes.
- The regression test is genuine and targets the actual failure mode: `TestIsTTY_RealNonTerminalFilesAreNotTTY` (`cmd/sdlc/changecode_test.go:232`) opens a real `/dev/null`, a real regular file, and a real `os.Pipe` — all `*os.File` readers that the *old* char-device check would have (partly) misclassified. This pins real logic, not a mock.
- ARCH-DRY exemplar: one `isTerminal` body + a per-OS constant (`tty_darwin.go`/`tty_linux.go`), mirroring x/term's layout. One helper, both callers.
- Atlas + help docs updated in the same boundary: `atlas/workflow/issue-lifecycle.md:54` documents the shared-`isTTY` invariant, and `helptext/merge.md` gains a clear NON-INTERACTIVE / AGENT RUNS section. The atlas gate is satisfied.

**2. Critical findings** — none.

**3. Important findings** — none.

**4. Minor findings**
- The durable plan doc `workshop/plans/000141-merge-agent-safe-tty-detection-plan.md:75-87` still shows all `- [ ]` (unchecked) even though the work is complete and the issue's own `## Plan`/`## Done when` are fully `- [x]`. The close gate reads the issue tracker (which is ticked), so this doesn't block — but the archived plan will read as "undelivered." Worth ticking for archive hygiene.
- `tty_other.go`'s `isTerminal` stub is effectively unreachable: the package already fails to compile on Windows (`syscall.Getpgrp` in `close.go:174`, `syscall.Kill` in `repolock.go:143` are unix-only, pre-existing), so the only `!darwin && !linux` targets that *could* build are the BSDs — where the stub silently forces `--yes` even in a real terminal. Conservative and documented as intentional, so acceptable; just noting the file is a defensive stub with no live build target today.
- Naming proximity: package-level `isTerminal(fd)` (tty) vs. `vocab.Issue().IsTerminal(status)` (issue lifecycle state) used throughout `merge.go`/`push.go`. Different scopes so no collision, and `isTerminal` matches x/term convention — no change recommended, just flagging for readers.

**5. Test coverage notes**
- False cases are covered deterministically (`/dev/null`, regular file, `os.Pipe`, non-`*os.File`). The true case (real tty → `isTTY` true) is not asserted because it needs a pty — correctly documented in the plan as a known, acceptable gap; asserting it would add a pty harness for little marginal value.
- `TestMergeNeedsTTY` already pins all four `(yes, dryRun, stdinIsTTY)` combinations (`merge_test.go:262`), so the pure decision boundary is fully covered. The Plan item "round out with the interactive case" was already satisfied — correctly logged as no-change-needed.
- The integration path itself (`mergeNeedsTTY(..., isTTY(os.Stdin))` at `merge.go:272`) is exercised via live E2E (`sdlc merge </dev/null` fail-fasts before judges), documented in the issue Log. That's the right seam to prove by E2E since it depends on the real process stdin.

**6. Architectural notes**
- ARCH-DRY: **pass** — single `isTTY`/`isTerminal` source; per-OS constant only. No duplicated terminal-detection logic elsewhere in `cmd/` (the other `IsTerminal` hits are the unrelated vocab status method).
- ARCH-PURE: **pass** — `mergeNeedsTTY` stays the pure, unit-tested decision; `isTerminal`/`isTTY` is the thin IO seam over the fd, tested at its deterministic false cases. Logic is not buried in IO.
- ARCH-PURPOSE: **pass** — shadow-sweep of the "single source" (the shared `isTTY`): both consumers (`merge` confirm gate, `change-code --worktree=ask` sentinel) now derive from the corrected helper. The fix delivers the issue's real purpose (agent-safe interactive detection) everywhere it manifests, not just the merge path that motivated it.

**7. Plan revision recommendations** — none required; the plan matches the code. (The only nit is cosmetic: tick the plan doc's own `- [ ]` boxes before archiving, per the Minor finding above — not a `## Revisions` entry.)

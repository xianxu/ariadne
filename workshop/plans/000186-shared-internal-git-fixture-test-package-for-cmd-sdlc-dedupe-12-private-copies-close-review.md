# Boundary Review — ariadne#186 (whole-issue close)

| field | value |
|-------|-------|
| issue | 186 — shared internal git-fixture test package for cmd/sdlc (dedupe ~12 private copies) |
| repo | ariadne |
| issue file | workshop/issues/000186-shared-internal-git-fixture-test-package-for-cmd-sdlc-dedupe-12-private-copies.md |
| boundary | whole-issue close |
| milestone | — |
| window | c164c4335754b898c52ad9f714ac32e06fb0576e..HEAD |
| command | sdlc close --issue 186 |
| reviewer | claude |
| timestamp | 2026-07-19T12:05:01-07:00 |
| verdict | FIX-THEN-SHIP |

## Review

I have enough to complete the review. Let me summarize my findings.

## Verdict

```verdict
verdict: FIX-THEN-SHIP
confidence: high
```

This is a clean, faithful, well-scoped ARCH-DRY consolidation: one stdlib-only `internal/testfix` package now owns the "run git in a dir, fatal on error" runner and the "init throwaway repo + test identity + gpgsign off (+ chdir / initial-commit / subdir)" setup sequence, and ~15 private copies delegate to it. I verified byte-faithfulness at every changed call site (trim contracts preserved, `Capture` stays stdout-only, chdir-ordering and `At`-placement don't change where files land), and confirmed compile-safety — every file that dropped `os/exec`/`os` has no dangling reference, and every file that kept `exec` still uses it. The one thing blocking a clean SHIP is a completeness gap: the sweep left one exact twin of `testfix.Git` unconsolidated (`runGitCommand`), which is precisely the duplication the issue exists to kill. Non-blocking, so FIX-THEN-SHIP. **Caveat:** the harness `Bash` tool failed at shell-init in this session (`EPERM` creating `session-env`), so I could not execute `go build`/`go test` — the plan's stated acceptance ("green suite") is unverified by me; my confidence rests on static faithfulness + import analysis, which is strong for a mechanical refactor.

### 1. Strengths
- **API surface is right-sized and stable** (`testfix.go:23-120`). `Git`/`Capture` split on the real distinction (combined output vs. stdout-only for parsed results), and the functional-option `Repo(t, Chdir()/InitialCommit()/At())` cleanly covers the three axes that actually varied across copies. Good downstream-consumable surface.
- **Faithful trim/output contracts.** `merge_e2e_test.go:33` and `gitx/window_test.go:82` correctly keep their `strings.TrimSpace` wrapper over the untrimmed `testfix.Git`; `milestonewindow_test.go:70` `captureGit` maps to `Capture` (`.Output()`, stdout-only) — no stderr folded into parsed SHAs.
- **Documented locals are genuinely different, not lazy skips.** Verified the `commit.gpgsign` grep returns only the two bare-origin push harnesses (`merge_e2e_test.go:54`, `issue_test.go:122`) + the source; the dated-commit helpers (`activetime`/`gitx`) inject `GIT_AUTHOR_DATE`, and `fetch_test.go:208` `gitInit` is remote-only with no identity. Correctly kept local.
- **No import cycle** — `testfix` imports only stdlib+testing; the internal-package placement lets `gitx`/`activetime` tests consume it. Sound.

### 2. Critical findings
None.

### 3. Important findings
- **`cmd/sdlc/issuefiles_test.go:290` — `runGitCommand` is an unconsolidated twin of `testfix.Git` (ARCH-DRY / ARCH-PURPOSE).** Its body is behaviorally identical (`exec.Command("git", args...)`, `cmd.Dir = dir`, `CombinedOutput()`, `t.Fatalf` on error, `return string(out)`) — the exact runner idiom the issue set out to single-source. Its *setup* already delegates (the test uses `hermeticRepo`→`testfix.Repo`), but the runner does not. The rg verification gate keyed on `commit.gpgsign`, so a runner-only twin in a file that gets its repo from `hermeticRepo` (no gpgsign line of its own) was invisible to the gate. This leaves the issue's Done-when ("the private copies delegate to it") and the ARCH-PURPOSE shadow-sweep ("flag any remaining hand-maintained restatement") not fully satisfied.
  - *Fix sketch:* make it a delegation — `func runGitCommand(t *testing.T, dir string, args ...string) string { t.Helper(); return testfix.Git(t, dir, args...) }` — matching how `gitIn`/`gitOut` were converted. One-line change; five call sites (`issuefiles_test.go:99-105`) unaffected.

### 4. Minor findings
- **`closereview_test.go:425` and `:491`** — two inline `for _, args := range [][]string{…} { exec.Command("git", …) }` add/commit loops (cwd, no dated env) are the same runner idiom and could be `testfix.Git(t, "", …)`. Out of the plan's enumerated named-runner scope, so optional polish, not a required delegation.
- **Dropped `t.Helper()` in the thin wrapper closures** (`windowRepo`/`publishgate` `runGit = func(args ...string){ testfix.Git(t, dir, args...) }`). The old inline closures called `t.Helper()`; the new ones don't. `testfix.Git` marks itself, but the un-helper'd wrapper means a git failure now attributes to the closure body line, not the test call site. Purely failure-ergonomics; add `t.Helper()` inside if desired. (`merge_e2e_test.go:33`'s named `git` correctly keeps it.)
- **`testfix.go:87-89`** — when `At()` is set, `dir := t.TempDir()` is created then discarded before `dir = filepath.Join(parent, name)`. Harmless (auto-cleaned), a hair wasteful. Fine as-is.

### 5. Test coverage notes
- No new tests are warranted — `testfix` is fixture scaffolding backing real tests; the suite staying green is the acceptance test, as the plan states. Appropriate.
- **I could not run `go build`/`go test`** (harness Bash EPERM at shell-init, sandbox-flag-independent). My verification is static: import-vs-usage analysis (no dangling/missing imports) and per-site byte-faithfulness. The main agent should re-run `go test ./cmd/sdlc/... ./pkg/vocab/ -count=1` to confirm green before recording the verdict — that's the one piece I'm relaying rather than confirming.

### 6. Architectural notes for upcoming work
- **ARCH-DRY:** PASS with one exception (the `runGitCommand` twin above). Setup-sequence sweep is complete (gpgsign grep clean save documented locals); runner sweep is complete for the named helpers (`git`/`gitIn`/`gitOut`/`captureGit` all delegate) except `runGitCommand`.
- **ARCH-PURE:** PASS (n/a in spirit). Fixtures are IO scaffolding by nature; the win is single-sourcing the glue, which is exactly what was done. `testfix` keeps the thin IO seam in one place.
- **ARCH-PURPOSE:** PASS-with-gap. Shadow-sweep over the runner consumers found one remaining hand-maintained restatement (`runGitCommand`). Closing it makes the "one source of truth for the runner idiom" purpose complete rather than 95%.
- Docs gate: atlas `sdlc-binary.md:300-324` is a **production/runtime** file-tree — it lists no `*_test.go` or test-fixture packages, so omitting a test-only `internal/testfix` entry is consistent with the atlas's own scope (not a gap). README needs nothing — no user-facing surface (no subcommand/flag/config). Both correctly untouched.

### 7. Plan revision recommendations
Add a `## Revisions` entry to `workshop/plans/000186-testfix-package-plan.md`:

> **### 2026-07-19 — close-review: one runner twin missed by the rg gate**
> The verification gate grepped `commit.gpgsign` (setup copies) + the named runners it enumerated (`git`/`gitIn`/`captureGit`/`gitOut`), but `issuefiles_test.go`'s `runGitCommand` — a byte-identical twin of `testfix.Git` whose *setup* already delegates via `hermeticRepo` (so it carries no `gpgsign` line) — fell outside both filters and stayed unconsolidated. Delegating it is the fix; the acceptance gate should also grep for `exec.Command("git"` in `*_test.go` outside the documented dated/bare/remote locals, not only `commit.gpgsign`, so a runner-only twin can't slip through again.

If `runGitCommand` is intentionally left as-is, instead record that decision (and why) so the plan stops implying a total sweep; otherwise fold the fix in and the plan matches the code.

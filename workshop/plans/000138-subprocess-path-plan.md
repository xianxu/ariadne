# SDLC Subprocess PATH Resolution Implementation Plan

> **For agentic workers:** Consult AGENTS.md Section 3 (Subagent Strategy) to determine the appropriate execution approach: use superpowers-subagent-driven-development (if subagents are suitable per AGENTS.md) or superpowers-executing-plans to implement this plan. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** A fresh review/judge subprocess that `sdlc` spawns can resolve the `sdlc` binary (and its sibling owner-`bin/` tools) on `PATH` — even when the spawning shell's startup files never put `ariadne/bin` on `PATH` — without requiring any user `~/.zshenv`/`~/.bash_profile` change.

**Architecture:** The running `sdlc` knows its own absolute path via `os.Executable()`; its directory *is* the owner `bin/`. The fix injects that directory onto the subprocess's `PATH` at the single launch seam (`judge.Run`), so the agent inherits a `PATH` that can find `sdlc`. The PATH-augmentation is a pure function (`binAugmentedEnv`) unit-tested without spawning; `Run` is the thin IO that calls `os.Executable()` + execs (ARCH-PURE). One injection site covers both close and milestone-close (every boundary review + every `sdlc judge` go through `Run`) — no per-caller duplication (ARCH-DRY). Launch-failure errors gain the attempted binary + the effective `PATH` for diagnosability.

**Tech Stack:** Go; pure unit test for the env builder; a process-level fixture (real `sh -c 'command -v …'` spawn with a minimal base env) proving resolution; `dispatch.go` doc comment for the contract.

---

## Core concepts

### Pure entities (the conceptual core)

| Name | Lives in | Status |
|------|----------|--------|
| `binAugmentedEnv` | `cmd/sdlc/internal/judge/dispatch.go` | new |

- **binAugmentedEnv** — `binAugmentedEnv(execPath string, env []string) []string`: returns `env` with `filepath.Dir(execPath)` prepended to the `PATH` entry (synthesizing a `PATH=` entry if none exists), using `os.PathListSeparator`. No-op when `execPath`'s dir is empty/`.`. `execPath` is a *parameter* (not an `os.Executable()` call) precisely so it's testable with a controlled temp dir.
  - **Relationships:** consumed once by `Run`; 1 augmented env per subprocess launch.
  - **DRY rationale:** the only place PATH is composed for spawned agents; both boundary reviews and `sdlc judge` inherit it.
  - **Future extensions:** could prepend multiple dirs (e.g. a peer's `bin/`) if a future agent needs sibling tools beyond `sdlc`.

### Integration points (where pure meets the world)

| Name | Lives in | Status | Wraps |
|------|----------|--------|-------|
| `Run` | `cmd/sdlc/internal/judge/dispatch.go:36` | modified | `exec.CommandContext` |
| `Dispatch` (error) | `cmd/sdlc/internal/judge/dispatch.go:105` | modified | launch-failure surface |

- **Run** *(modified)* — resolves `os.Executable()`; on success sets `cmd.Env = binAugmentedEnv(exe, os.Environ())` before `CombinedOutput()`. If `os.Executable()` errors (rare), leaves `cmd.Env` nil (inherits parent env, today's behavior) — best-effort, never blocks the review. The single launch seam tests already stub for other assertions; the env logic is exercised via the pure helper + the process fixture, not the stub.
- **Dispatch (error)** *(modified)* — the real-launch-failure branch (`dispatch.go:105-108`) gains the resolved owner-`bin/` dir + the effective `PATH` in the error string, so a "binary missing" failure names what was attempted and where it looked (Done-when: diagnosable failures).

**Test surface.** `binAugmentedEnv` gets a pure table test (PATH present → prepended; PATH absent → synthesized; empty dir → no-op; ordering/separator correct). A process-level fixture (`TestRun_InjectsBinDirOnPath` or similar) drops a dummy executable named `sdlc` into a temp dir, builds an env via `binAugmentedEnv(tmp/sdlc, minimalEnv)` where `minimalEnv` is a deliberately narrow `PATH` (e.g. `/usr/bin:/bin`), and runs a real `sh -c 'command -v sdlc'` with that env — asserting it resolves to the temp dummy. This proves the end-to-end "narrow PATH still finds sdlc" claim without invoking a real agent.

---

## Design decisions

- **D1 — inject the owner bin/, don't rewrite the prompt.** The boundary-review prompt does not currently tell the reviewer to run `sdlc <verb>` (the architecture block is *embedded*, not fetched), and the reviewer may run `sdlc` on its own initiative. Augmenting the subprocess `PATH` (the spec's "an environment containing the SDLC owner bin/") covers every `sdlc` call the agent makes, not just prompt-scripted ones — strictly more robust than substituting an absolute path into prompt text.
- **D2 — `os.Executable()` is the source of the owner bin/.** Its directory is where the running `sdlc` lives (`ariadne/bin`). This works unchanged from a downstream repo (`pair`): the binary is `…/ariadne/bin/sdlc` regardless of cwd, so its dir is the right thing to add. No repo-walking, no remote, no hardcoded path.
- **D3 — single seam (ARCH-DRY).** Every agent launch — boundary reviews (close + milestone-close) and `sdlc judge` — flows through `Run`. Injecting there covers all of them; no DispatchOptions field to thread through callers.
- **D4 — best-effort, never blocking.** If `os.Executable()` fails, fall back to the inherited env (today's behavior). The review is too important to abort over a PATH nicety; the diagnostics (D5) make any residual failure legible.
- **D5 — diagnosable launch failures.** The launch-failure error names the attempted agent + the effective PATH (incl. the injected bin/), so the next environment issue is debuggable from the error alone (Done-when).
- **Non-goal — touching the user's shell config.** Explicitly out of scope (issue spec): the fix must not depend on `~/.zshenv`/`~/.bash_profile`.

---

## Chunk 1: PATH injection + diagnostics

### Task 1: `binAugmentedEnv` (pure)

**Files:** Modify `cmd/sdlc/internal/judge/dispatch.go`; Test `cmd/sdlc/internal/judge/dispatch_test.go` (new or existing)

- [ ] **Step 1: Write the failing test** — table cases:

```go
func TestBinAugmentedEnv(t *testing.T) {
    sep := string(os.PathListSeparator)
    // PATH present → bin dir prepended, rest preserved.
    got := binAugmentedEnv("/w/ariadne/bin/sdlc", []string{"HOME=/h", "PATH=/usr/bin" + sep + "/bin"})
    wantPath := "PATH=/w/ariadne/bin" + sep + "/usr/bin" + sep + "/bin"
    if !contains(got, wantPath) || !contains(got, "HOME=/h") { t.Errorf("got %v", got) }
    // PATH absent → synthesized.
    if !contains(binAugmentedEnv("/w/ariadne/bin/sdlc", []string{"HOME=/h"}), "PATH=/w/ariadne/bin") { ... }
    // Empty dir → no-op.
    in := []string{"PATH=/usr/bin"}
    if !reflect.DeepEqual(binAugmentedEnv("sdlc", in), in) { t.Errorf("bare name should be a no-op") }
}
```

- [ ] **Step 2: Run → fails** (undefined).
- [ ] **Step 3: Implement** `binAugmentedEnv` (prepend `filepath.Dir(execPath)` to the `PATH=` entry via `os.PathListSeparator`; synthesize if absent; no-op when dir is `""`/`.`).
- [ ] **Step 4: Run → PASS.**
- [ ] **Step 5: Commit** — `#138: binAugmentedEnv — prepend owner bin/ to subprocess PATH (pure)`

### Task 2: wire into `Run` + diagnostics

**Files:** Modify `cmd/sdlc/internal/judge/dispatch.go` (`Run`, `Dispatch` error)

- [ ] **Step 1:** In `Run`, after building `cmd`, `if exe, err := os.Executable(); err == nil { cmd.Env = binAugmentedEnv(exe, os.Environ()) }`.
- [ ] **Step 2:** In `Dispatch`'s real-launch-failure branch, extend the error: `fmt.Errorf("dispatch %s (PATH includes %s): %w", name, <bin dir or "?">, runErr)` — name the attempted agent + the injected owner bin/ so a missing-binary failure is diagnosable.
- [ ] **Step 3: Build + judge suite** — `go test ./cmd/sdlc/internal/judge/...`. Existing dispatch tests stub `Run`, so they're unaffected.
- [ ] **Step 4: Commit** — `#138: inject owner bin/ into agent subprocess env + diagnosable launch errors`

### Task 3: process-level fixture (minimal PATH resolves sdlc)

**Files:** Modify `cmd/sdlc/internal/judge/dispatch_test.go`

- [ ] **Step 1: Write the test** — `t.TempDir()`; write an executable file named `sdlc` (a `#!/bin/sh` stub, `chmod 0755`); build `env := binAugmentedEnv(filepath.Join(tmp,"sdlc"), []string{"PATH=/usr/bin:/bin"})`; run `cmd := exec.Command("sh", "-c", "command -v sdlc"); cmd.Env = env`; assert the output resolves to `tmp/sdlc`. This is the "subprocess with a minimal PATH" coverage the Done-when requires — a real spawn, no agent.
- [ ] **Step 2: Run → PASS** (proves the narrow-PATH agent would find `sdlc`).
- [ ] **Step 3: Commit** — `#138: process-level test — minimal-PATH subprocess resolves sdlc`

## Chunk 2: Docs

### Task 4: contract comment + atlas

**Files:** `dispatch.go` (doc comment on `Run`/`binAugmentedEnv`); `atlas/workflow/sdlc-binary.md` (boundary-review section)

- [ ] **Step 1:** Document the contract: agent subprocesses get the owner `bin/` on `PATH` (via `os.Executable()`), so a fresh reviewer resolves `sdlc` without the user's shell startup files; works from downstream repos. One sentence in the atlas boundary-review block.
- [ ] **Step 2: Commit** — `#138: doc — subprocess owner-bin PATH contract`

---

## Done-when mapping

| Issue Done-when | Delivered by |
|---|---|
| fresh review subprocess runs `sdlc --help` from a downstream repo w/o user shell startup | Tasks 1–2 (D1, D2) |
| tests / process fixture cover a minimal-PATH subprocess | Task 3 |
| command path uses the resolved SDLC binary instead of bare `sdlc` on a global PATH | Tasks 1–2 (PATH carries the resolved bin dir) |
| error output includes attempted path + PATH on resolution failure | Task 2 (D5) |

## Non-goals

- No change to the user's shell config; no dependence on `~/.zshenv`/`~/.bash_profile` (issue spec).
- Not rewriting prompt text to embed an absolute `sdlc` path (D1 — PATH injection covers all `sdlc` calls, not just scripted ones).
- Not changing agent selection, verdict format, trailers, or gates.

## Revisions

- **2026-06-29 — owner-bin single-source (change-code plan-quality gate, ARCH-DRY).**
  Per the gate's Finding 1, the owner-bin resolution was single-sourced into a new
  `ownerBinDir() (string, error)` = `filepath.Dir(os.Executable())`, consumed by both
  `Run` (build the subprocess env) and `Dispatch` (format the launch-failure
  diagnostic). Consequently `binAugmentedEnv`'s signature changed from
  `(execPath string, env []string)` (deriving `filepath.Dir` internally) to
  `(binDir string, env []string)` — it now takes the already-resolved dir. Add to
  the Core-concepts *Integration points* table: `ownerBinDir` (`dispatch.go`, new,
  wraps `os.Executable`). Line refs drifted: `Run` `:36`→`:77`, `Dispatch` error
  `:105`→`:149`.
- **2026-06-29 — close-review FIX-THEN-SHIP follow-up.** The boundary review flagged
  (Important #2) that the new launch-failure diagnostic had no test pinning its
  content; added assertions to `TestDispatch_LaunchError_Surfaces` (error contains
  the agent name + `owner bin` + `PATH=`). Also widened the diagnostic to include the
  effective `PATH` (literal Done-when parity) with a `"?"` dir fallback (Minor).

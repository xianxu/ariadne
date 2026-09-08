# Queue Datatype + `sdlc queue` Implementation Plan

> **For agentic workers:** Consult AGENTS.md Section 3 (Subagent Strategy) to determine the appropriate execution approach: use superpowers-subagent-driven-development (if subagents are suitable per AGENTS.md) or superpowers-executing-plans to implement this plan. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** A per-repo advisory work queue whose single copy lives on `origin/main`, edited through intent-carrying operations that replay across concurrent writes, plus the `queue` datatype prototype that documents it.

**Architecture:** Three pure entities (`Line`, `Doc`, `Intent`) with no IO, behind one generic integration point (`gitx.TrunkFile`) that reads and compare-and-swap-writes a single path on a remote branch with no working tree. `TrunkFile` uses `gitx`'s **own** package-level `run` shim (`window.go:32`) plus a sibling `runEnv` for the three calls needing `GIT_INDEX_FILE` — NOT the `gitRunner` interface from package `main`, which would be an import cycle (`main` already imports `gitx`) and whose ~20 callers need nothing from this. `TrunkFile.Update` takes a **transform function** and retries by re-reading and re-calling it — so the retry loop is generic and the *transform* decides mergeability. The queue supplies a replaying transform (`Intent.Apply`); `ariadne#207` will later supply a content-setting one for issue files. That split is the whole design: one CAS/retry primitive, per-caller semantics.

**Tech Stack:** Go, cobra, git plumbing (`ls-tree`, `cat-file`, `read-tree`, `hash-object`, `update-index`, `write-tree`, `commit-tree`, `push`), the existing `gitRunner` seam, `internal/testfix` for real-git fixtures.

---

## Core concepts

### Pure entities (the conceptual core)

| Name | Lives in | Status |
|------|----------|--------|
| `Line` | `cmd/sdlc/internal/queue/line.go` | new |
| `Doc` | `cmd/sdlc/internal/queue/doc.go` | new |
| `Intent` | `cmd/sdlc/internal/queue/intent.go` | new |

- **Line** — one queue entry: `Ref` (e.g. `ariadne#207`), `WhyNow`, optional `Tag`, and `Kind` (issue | project).
  - **Relationships:** N:1 with `Doc` (a Doc owns an ordered slice of Lines). No back-reference.
  - **DRY rationale:** Parsing and rendering a line exist once. Without it, the verb, the future close-gate sweep (deferred), and any reader each re-derive the `- ref — why [tag]` shape by regex.
  - **Future extensions:** A rank key, if concurrent reordering ever justifies making order derived rather than positional (considered and rejected in the Spec — it puts a machine field into a human-read list).

- **Doc** — the parsed queue file: leading prose preserved verbatim, then the ordered `[]Line`.
  - **Relationships:** 1:N with Line. 1:1 with the file at `workshop/queue.md` on the trunk.
  - **DRY rationale:** Round-tripping (`Parse` then `Render` is byte-identical for an unmodified doc) lives in one place, so an edit can never reformat the parts it did not touch.
  - **Future extensions:** Sections (e.g. `## Now` / `## Soon`) if flat ordering proves too coarse.

- **Intent** — a tagged union of the three operations plus `Apply(Doc) (Doc, error)`. **This is the entity the whole design rests on.** Because an Intent is an operation rather than file content, applying it to a base that moved preserves a peer's concurrent edit; a content-carrying write cannot.
  - **Relationships:** 1:1 with a single `sdlc queue` invocation. Holds no reference to a Doc — it is applied *to* one.
  - **DRY rationale:** First occurrence of a pattern likely to recur: any future trunk-backed artifact wanting merge-safe edits supplies an Intent-shaped transform.
  - **Future extensions:** New operations are new variants; `Apply` stays exhaustive over the union so the compiler names every site that must handle one (`ARCH-ORDER`: legal states enumerated, not emergent).

**Test surface:** all three get colocated `_test.go` files that run with **no git and no IO**. `Intent.Apply` is table-tested over the interleaving cells in the Spec — that table is the test list.

### Seam capability audit — the rule, and this plan's enumeration

**Rule:** before naming a shim as the seam, enumerate every `exec.Cmd` parameter
the work needs and check each against the shim's signature *at file:line*. Two
rounds of this gate found a capability missing one at a time (`Env`, then `Dir`
and stderr) because the check was done per-instance. The enumeration below is the
whole surface, decided in one pass.

`gitx.run` (`window.go:32`) is `exec.Command(name, args...).Output()`.

| `exec.Cmd` parameter | Needed? | Why | In `gitx.run`? |
|---|---|---|---|
| argv | yes | the plumbing commands | **yes** |
| `Dir` | **yes** | must operate on a named repo. Without it `Read` fetches from, and `Update` pushes `<commit>:main` to, whatever repo the process cwd sits in — in a test, the real ariadne repo and its real origin | **no** |
| `Env` | **yes** | `GIT_INDEX_FILE` for `update-index`/`write-tree`; only `read-tree` has `--index-output` | **no** |
| stderr | **yes** | three clauses render git's own text — "surface the last rejection" (Task 4), "the warning names the risk" (Task 5), "print the trunk state and the intent" (Task 11). `.Output()` drops it except inside `ExitError.Stderr` | **no** |
| stdout **separable from** stderr | **yes** | five call sites parse stdout as a value (`hash-object`, `write-tree`, `commit-tree`, `rev-parse`, `config`) and one returns file CONTENT. Combining them — the first implementation — folded git's "CRLF will be replaced by LF" warning into a parsed blob hash, and would have corrupted every read. **Added after implementation proved it; the audit's first pass asked for stderr but not for this.** | **n/a** |
| `Stdin` | no | blobs are written by `hash-object -w --path <tmpfile>`, from a file, never piped | n/a |
| context / timeout | no | `fetch` and `push` are network calls that can hang, but no git call in this binary takes a context today (`issueids.go`'s fetch included). Matching the existing convention rather than introducing a second one here; a hung fetch is the named residue, not an oversight | n/a |

**Resolution:** add
`gitx.runGitIn(dir string, env []string, args ...string) (stdout, stderr []byte, err error)`
— `cmd.Dir` and `cmd.Env` set, and the two streams returned **separately**. A **new** shim, not a widened
`run`: `run`'s existing callers were written against `.Output()` semantics, and
switching them to combined output would silently fold stderr into strings they
parse.

Two consequences that must be written into the code, not just remembered:

- `GIT_INDEX_FILE` is passed as an **absolute** path. A relative one resolves
  against the child's working directory, which is now `cmd.Dir` — unambiguous only
  if absolute.
- `NewTrunkFile` **refuses an empty dir**. The dangerous failure here is silent
  operation on the process cwd, and a guard makes it impossible rather than merely
  avoided by convention.

### Integration points (where pure meets the world)

| Name | Lives in | Status | Wraps |
|------|----------|--------|-------|
| `TrunkFile` | `cmd/sdlc/internal/gitx/trunkfile.go` | new | git plumbing on a remote ref |
| `queueCmd` | `cmd/sdlc/queue.go` | new | cobra + stdout/stderr |

- **TrunkFile** — reads and CAS-writes one path on a remote branch with **no working tree**. `Read(path) ([]byte, error)` (requires a reachable remote);
`ReadDegraded(path) ([]byte, string, error)` (returns a warning and the stale ref's content when offline);
`Update(path, msg string, transform func([]byte) ([]byte, error)) error`. `Update` loops: fetch → read blob → `transform` → build tree in a temp index → `commit-tree` → `push <commit>:main`; on non-fast-forward, loop again (max 3), re-reading and **re-calling transform** each time.
  - **Seam:** `gitx.runGitIn(dir, env, args...)` — a new package-level var beside `run` (`window.go:32`), carrying `Dir`, `Env`, and stdout/stderr returned **separately** per the audit above. `run` is left alone; `gitRunner` is deliberately NOT widened, since nothing outside this file needs any of it.
  - **Injected into:** `queueCmd` consumes it through an interface **declared in package `main`** (consumer-side, Go idiom) so the verb is testable with a fake; `gitx` exports the concrete type. `Intent.Apply` is passed in as the transform and never sees git.
  - **Offline:** a failed fetch degrades a READ to the stale tracking ref with a loud warning and REFUSES a write — the policy `issueids.go:40-49,126-145` already settled (`ARCH-DRY`).
  - **Future extensions:** `ariadne#207` consumes this for issue files with a content-setting transform. Its retry semantics then follow from its transform, not from a second retry loop — which is why #207's own spec defect (a content-preserving retry that re-lands a colliding issue id) cannot be built on top of this.

- **queueCmd** — the cobra verb: parse args into an `Intent`, call `TrunkFile`, render output and errors.
  - **Injected into:** nothing; it is the outermost shell.
  - **Future extensions:** `--repo` for a cross-repo queue (deferred in the Spec).

**Test surface for integration points.** `TrunkFile` is tested against a **real throwaway repo with a local bare origin** via `internal/testfix` — the established pattern in `issueids_test.go:29` and `merge_e2e_test.go:48`. `ARCH-MOCK`: git is the external binary and a temp repo is its portable stateful fake. A function-call mock cannot produce a genuine non-fast-forward rejection, which is the single most important behavior here.

---

## Chunk 1: M1 — the trunk-file CAS primitive

### Task 1: `TrunkFile.Read`

**Files:**
- Create: `cmd/sdlc/internal/gitx/trunkfile.go`
- Test: `cmd/sdlc/internal/gitx/trunkfile_test.go`

- [x] **Step 1: Write the containment guard test FIRST** — `NewTrunkFile("", ...)` returns an error. This is the test that makes "operates on the real repo by accident" impossible rather than merely unlikely, so it is written before the feature it guards.

```go
func TestNewTrunkFile_RefusesEmptyDir(t *testing.T) {
	if _, err := NewTrunkFile("", "origin", "main"); err == nil {
		t.Fatal("empty dir must be refused — otherwise git runs against the process cwd")
	}
}
```

- [x] **Step 2: Write the failing read test** — a repo with a bare origin, a file committed and pushed to main, then `Read` returns its bytes from `origin/main` while the working tree is on a *different* branch with *different* content. Note it passes `repo` explicitly and does **not** `testfix.Chdir()`: the point is that `Dir` carries the scoping, so a cwd-dependent implementation fails this test rather than silently passing in CI and destroying a developer's checkout.

```go
func TestTrunkFile_ReadsFromTrunkNotWorktree(t *testing.T) {
	repo, _ := trunkFixture(t) // helper: repo + bare origin, "queue.md" = "from main\n" on main
	testfix.Git(t, repo, "checkout", "-b", "feature")
	os.WriteFile(filepath.Join(repo, "queue.md"), []byte("from branch\n"), 0o644)

	tf := NewTrunkFile(repo, "origin", "main") // uses gitx's own run shim
	got, err := tf.Read("queue.md")
	if err != nil { t.Fatal(err) }
	if string(got) != "from main\n" {
		t.Errorf("Read = %q, want the TRUNK copy %q", got, "from main\n")
	}
}
```

- [x] **Step 3: Run both, verify they fail** — `go test ./cmd/sdlc/internal/gitx/ -run TestTrunkFile -v`. Expected: FAIL, `undefined: NewTrunkFile`.

- [x] **Step 4: Add `runGitIn`** per the seam audit — `cmd.Dir`, `cmd.Env`, and stdout/stderr returned separately, a new package-level var beside `run`. One test asserting it passes `Dir` and `Env` through, returns stderr on failure, and keeps stdout clean.

- [x] **Step 5: Implement `Read`** — refuse an empty dir; `fetch --quiet origin +refs/heads/main:refs/remotes/origin/main`, then `cat-file blob origin/main:<path>`, both through `runGitIn(dir, nil, ...)`. A missing path is not an error: return `(nil, nil)` so a first-ever write works. Reuse `issueids.go`'s explicit refspec form — `git fetch origin main` leaves `FETCH_HEAD` but does not always move the tracking ref.

- [x] **Step 6: Run, verify PASS.**

- [x] **Step 7: Commit** — `#209 M1: read one path from the trunk without a checkout`

### Task 2: `TrunkFile.Update` — happy path

**Files:** Modify `trunkfile.go`; Test: `trunkfile_test.go`

- [x] **Step 1: Write the failing test** — `Update` on a repo with **no worktree on main anywhere** (checkout is on `feature`) appends a line; assert `git show origin/main:queue.md` in the *bare origin* contains it, and that the working tree is untouched (`git status --porcelain` empty).

- [x] **Step 2: Run, verify it fails.**

- [x] **Step 3: Implement the plumbing sequence** through `runGitIn` (added in Task 1). In order, with a temp index:

```go
idx, _ := filepath.Abs(filepath.Join(tmp, "index")) // ABSOLUTE: resolves against cmd.Dir otherwise; NEVER $GIT_DIR/index
defer os.Remove(idx)                          // every exit path
// GIT_INDEX_FILE=idx git read-tree origin/main
// blob := git hash-object -w --path <relpath> <tmpfile>   // --path: apply .gitattributes
// GIT_INDEX_FILE=idx git update-index --add --cacheinfo 100644,<blob>,<relpath>
// tree := GIT_INDEX_FILE=idx git write-tree
// commit := git commit-tree <tree> -p origin/main -m <msg>   // + -S if commit.gpgsign
// git push origin <commit>:main
```

- [x] **Step 4: Run, verify PASS.**

- [x] **Step 5: Commit** — `#209 M1: CAS-write one path to the trunk with no working tree`

### Task 3: Round-trip fidelity

- [x] **Step 1: Write the failing test** — repo with `.gitattributes` marking `*.md` as `text eol=lf`; write CRLF content; assert the pushed blob checks out byte-identical to the intended content. Second case: a repo with `commit.gpgsign=true` produces a signed commit (skip via `t.Skip` if no signing key in the environment).
- [x] **Step 2: Run, verify it fails.**
- [x] **Step 3: Fix** — this is what `--path` buys; add `-S` when `git config --get commit.gpgsign` is true.
- [x] **Step 4: Run, verify PASS.**
- [x] **Step 5: Commit** — `#209 M1: apply gitattributes and signing to trunk writes`

### Task 4: Bounded retry, and the transform re-runs

**This is the milestone's load-bearing test.**

- [x] **Step 1: Write the failing test** — two publishers against one bare origin. Publisher A reads, then B reads-and-pushes, then A pushes → A is rejected, retries, and **A's transform is called a second time with B's content**. Assert both A's and B's lines are present, and record the transform call count.

```go
func TestTrunkFile_RetryReRunsTransformOnMovedBase(t *testing.T) {
	repo, origin := trunkFixture(t)
	tf := NewTrunkFile(repo, "origin", "main")

	calls := 0
	transform := func(old []byte) ([]byte, error) {
		calls++
		if calls == 1 { pushPeerLine(t, origin, "- peer\n") } // B wins the race
		return append(append([]byte{}, old...), []byte("- mine\n")...), nil
	}
	if err := tf.Update("queue.md", "test", transform); err != nil { t.Fatal(err) }

	got := showTrunk(t, origin, "queue.md")
	for _, want := range []string{"- peer\n", "- mine\n"} {
		if !strings.Contains(got, want) { t.Errorf("trunk missing %q:\n%s", want, got) }
	}
	if calls != 2 { t.Errorf("transform called %d times, want 2 (re-run on the moved base)", calls) }
}
```

- [x] **Step 2: Run, verify it fails** — without the retry, the push errors out.
- [x] **Step 3: Implement the bounded loop** — max 3 attempts; on exhaustion return an error carrying the **last** rejection text, not a generic message.
- [x] **Step 4: Run, verify PASS.**
- [x] **Step 5: Add the exhaustion test** — a transform whose peer pushes on *every* call; assert exactly 3 attempts were made and that the error carries **git's own rejection text** (e.g. `non-fast-forward`), which is only reachable because `runGitIn` returns stderr separately. An assertion on a generic wrapper message would pass against a shim that dropped stderr and prove nothing.
- [x] **Step 6: Commit** — `#209 M1: bounded CAS retry that re-runs the transform on the new base`

### Task 5: Offline and no-origin policy

**Files:** Modify `trunkfile.go`, `trunkfile_test.go`

- [x] **Step 1: Write the failing tests** — three cases against a fixture whose origin URL points at a nonexistent path: (a) `Read` succeeds from the stale tracking ref and the warning text names the risk; (b) `Update` refuses with a distinct error (a CAS push has nothing to compare against); (c) a repo with **no `origin` remote at all** reads the local ref and says so rather than erroring obscurely.
- [x] **Step 2: Run, verify they fail.**
- [x] **Step 3: Implement**, mirroring `issueids.go:40-49,126-145` in shape and in warning tone. Do not invent a second offline policy for the same binary (`ARCH-DRY`).
- [x] **Step 4: Run, verify PASS.**
- [x] **Step 5: Commit** — `#209 M1: degrade reads offline, refuse writes`

### Task 6: Temp-index hygiene

- [x] **Step 1: Write the failing test** — a transform that returns an error mid-`Update`; assert the temp index file no longer exists and `$GIT_DIR/index` is unmodified (compare mtime + the working tree stays clean).
- [x] **Step 2: Run, verify it fails** (if cleanup is only on the success path).
- [x] **Step 3: Fix with `defer`.**
- [x] **Step 4: Run, verify PASS.**
- [x] **Step 5: Commit** — `#209 M1: remove the temp index on every exit path`

### Task 6b: Atlas — the new primitive

- [x] **Step 1: Update `atlas/workflow/sdlc-binary.md`** with `gitx.TrunkFile`: what it owns (one path on the trunk, CAS-written, no working tree), the transform-function seam, and that `ariadne#207` is its second consumer. Add the row to `atlas/index.md` if a new file is created.
- [x] **Step 2: Commit** — `#209 M1: atlas — the trunk-file primitive`

- [x] **M1 — `sdlc milestone-close --issue 209 --milestone M1`**

---

## Chunk 2: M2 — the queue itself

### Task 7: `Line` parse/render

**Files:** Create `cmd/sdlc/internal/queue/line.go`, `line_test.go`

- [ ] **Step 1: Write failing table test** — round-trips `- ariadne#207 — after #206 merges, same dispatch [sdlc]` into `{Ref, WhyNow, Tag}` and back byte-identically. Cases: no tag; em-dash vs hyphen separator; a project ref; a malformed line (preserved verbatim, not dropped).
- [ ] **Step 2: Run, verify it fails.**
- [ ] **Step 3: Implement.** A line that does not parse is **kept as-is**, never discarded — the file is human-editable and losing a hand-written line would be the worst failure this feature can have.
- [ ] **Step 4: Run, verify PASS.**
- [ ] **Step 5: Commit** — `#209 M2: parse and render one queue line`

### Task 8: `Doc` round-trip

**Files:** Create `doc.go`, `doc_test.go`

- [ ] **Step 1: Write the failing test** — `Render(Parse(x)) == x` for a doc with leading prose, blank lines, and a trailing comment block.
- [ ] **Step 2: Run, verify it fails.**
- [ ] **Step 3: Implement** — preserve everything that is not a queue line verbatim, in place.
- [ ] **Step 4: Run, verify PASS.**
- [ ] **Step 5: Add `FuzzDocRoundTrip`** asserting `Render(Parse(b)) == b` over arbitrary bytes, seeded with the adversarial forms: no trailing newline, CRLF, a near-miss line, unicode vs ASCII dashes, a bare `- ` at EOF, an empty file. `Doc` parses a human-editable file arriving from the trunk — input this process did not produce (`ARCH-SECURE`) — so the guard must be mechanical, not four chosen inputs. This is what makes "a malformed line is kept as-is, never discarded" true across the whole input space.
- [ ] **Step 6: Run** — `go test ./cmd/sdlc/internal/queue/ -run FuzzDocRoundTrip -fuzz FuzzDocRoundTrip -fuzztime 30s`. Expected: no failing corpus entry.
- [ ] **Step 7: Commit** — `#209 M2: round-trip the queue document, fuzz-guarded`

### Task 9: `Intent.Apply` — the interleaving table

**Files:** Create `intent.go`, `intent_test.go`

- [ ] **Step 1: Write the failing table test — one case per Spec row:**

| Case | Assert |
|---|---|
| `Add` onto a base that gained unrelated lines | both present, peer's line untouched |
| `Remove X`, X already gone | no-op, no error |
| `Add X`, X already present | one line, newer why-now, `Applied.Note` says which was kept |
| `Move X --before Y`, Y deleted | returns `ErrAnchorMissing` naming X and Y |
| `Move X`, X deleted | returns `ErrSubjectMissing` |
| `Add` of a malformed ref | rejected before any git call |
| `Add` with a newline or control char in why-now | rejected before any git call — a newline would become a second entry and break `Doc`'s round-trip |
| `Add` with ` — ` or `[...]` inside why-now | rejected, not escaped — escaping costs the human-readability the format exists for |

- [ ] **Step 2: Run, verify it fails.**
- [ ] **Step 3: Implement `Apply`** as an exhaustive switch over the union. Errors are typed so the verb can render the current queue alongside them.
- [ ] **Step 4: Run, verify PASS** — `go test ./cmd/sdlc/internal/queue/ -v`. **No git, no mocks** in this package's tests; if one is needed, the purity boundary is wrong.
- [ ] **Step 5: Commit** — `#209 M2: apply queue intents, with every interleaving cell tested`

### Task 10: The verb

**Files:** Create `cmd/sdlc/queue.go`, `queue_test.go`, `cmd/sdlc/helptext/queue.md`; Modify `cmd/sdlc/main.go` (one `add(...)` line, in workflow order)

- [ ] **Step 1: Write the failing test** — `sdlc queue add ariadne#210 "why"` against a real fixture with no main worktree lands the line on the bare origin; bare `sdlc queue` lists from the trunk, not the working tree.
- [ ] **Step 2: Run, verify it fails.**
- [ ] **Step 3: Implement** — build the `Intent`, hand `Intent.Apply` to `TrunkFile.Update` as the transform. The verb itself holds no retry logic.
- [ ] **Step 4: Run, verify PASS.**
- [ ] **Step 5: Write `helptext/queue.md`** — the four operations, the trunk-is-the-base rule, and the interleaving table.
- [ ] **Step 6: Commit** — `#209 M2: sdlc queue, backed by the trunk`

### Task 11: Failure is a handoff

- [ ] **Step 1: Write the failing test** — a `move` whose anchor was concurrently deleted; assert stderr contains **both** the current remote queue and the unapplied intent, and the exit is non-zero.
- [ ] **Step 2: Run, verify it fails.**
- [ ] **Step 3: Implement the error rendering** — per `sdlc --help`, errors are next-action specs; an operator or agent must be able to re-derive the edit from the message alone.
- [ ] **Step 4: Run, verify PASS.**
- [ ] **Step 5: Commit** — `#209 M2: a refused queue edit prints the trunk state and the intent`

### Task 12: The datatype prototype

**Files:** Create `construct/datatype/queue.md`

- [ ] **Step 1: Write it** following `construct/datatype/target.md`'s shape — `type: type`, `name: queue`, a discovery `description` triggering on "queue", "what's next", "pick the next thing", "plan the sequence", then narrative prose covering every Done-when clause: advisory-not-binding vs `deps:`; division of labour with project `status`/`roadmap` and a project's `## Breakdown`; the line format; issue-line vs project-line with its failure mode; **the file is read through `sdlc queue` because a working-tree copy can be stale**; and the rot risk pointing at the deferred close-gate removal.
- [ ] **Step 2: Regenerate** — `make weave`.
- [ ] **Step 3: Verify no hand-edit was needed** — `grep -n queue construct/generated/datatype/SKILL.md` finds it, AND `git diff --stat cmd/datatype/ construct/local/datatype/` is empty. Do **not** diff `construct/generated/`: it is gitignored (`.gitignore:30`, per `atlas/workflow/data-artifacts.md:20` it is per-repo and never committed), so that check passes whether or not a hand-edit happened. The file a wrongful hand-edit would touch is `cmd/datatype/SKILL.md.tmpl`.
- [ ] **Step 4: Commit** — `#209 M2: queue datatype prototype`

### Task 13: Seed, through the verb

- [ ] **Step 1: Seed** ariadne's real current ordering using `sdlc queue add` — **not** by hand-writing the file. This is the cheapest test of whether the format is right, and hand-writing it would skip the thing under test.
- [ ] **Step 2: Verify** `sdlc queue` from a feature branch renders what was added.
- [ ] **Step 3: Note on #207** that `gitx.TrunkFile` now exists and its content-setting transform is the seam to consume — plus the retry-semantics defect flagged in this issue's Spec.
- [ ] **Step 4: Commit** — `#209 M2: seed ariadne's queue through the verb`

### Task 13b: Atlas — the verb and the noun

- [ ] **Step 1: Update `atlas/workflow/sdlc-binary.md`** — add `queue` to the verb table with what it defends.
- [ ] **Step 2: Update the terminology/datatype map** so `queue` as a noun is findable, and confirm `atlas/index.md` links every file it should.
- [ ] **Step 3: Commit** — `#209 M2: atlas — the queue verb and noun`

- [ ] **M2 — `sdlc close --issue 209`**

---

## Verification

- `go test ./cmd/sdlc/...` green (note: `TestFleetPlanHasAuthoritativeCorrectedCoreConceptInventory` fails pre-existing on a hardcoded path to an archived plan — ariadne#210).
- `sdlc queue` correct from: a feature branch, a detached HEAD, and a checkout with no worktree on main anywhere.
- Every row of the Spec's interleaving table has a named test.
- `construct/generated/datatype/SKILL.md` lists `queue` with no hand-edit.

---

## Revisions

### 2026-09-07 — M1 as built

Two changes the plan did not anticipate, both forced by the code rather than
chosen, and both now reflected above rather than left as a changelog to replay:

1. **`runGitIn` returns stdout and stderr separately.** The seam audit asked for
   stderr and stopped there. Implementation proved that insufficient: with
   `CombinedOutput`, a repo carrying `*.md text eol=lf` makes git print "CRLF will
   be replaced by LF" on `hash-object`, and that warning landed inside the parsed
   blob hash — and would have landed inside file content on every read. The
   gitattributes round-trip test (Task 3) is what caught it. The audit table now
   carries the row.

2. **Retryability is observed, not parsed.** The plan implied matching git's
   rejection text. The M1 review showed a bare `"rejected"` match also catches
   refusals retrying cannot fix (a declined pre-receive hook, no push permission),
   which burn the whole attempt budget and then report contention that never
   happened. `Update` now re-resolves the tracking ref after a failed push and
   retries only if the trunk actually moved. Same reasoning replaced
   `isMissingPath`'s phrase matching with `cat-file -e`'s exit code, and a raw
   `== "true"` config compare with `--type=bool` (git stores `commit.gpgsign`
   verbatim, so a repo configured `yes` was silently getting unsigned commits).
   One rule: **observe git's state, do not parse its English.**

**Commit granularity diverged from the plan's one-task-one-commit shape:** Tasks
2+4 landed together (the CAS write is not meaningfully testable without the retry
that defines its contract) and Tasks 3+5 together. Recorded rather than tidied
away.

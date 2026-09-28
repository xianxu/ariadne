# sdlc move Implementation Plan

> **For agentic workers:** Consult AGENTS.md Section 3 (Subagent Strategy) to determine the appropriate execution approach: use superpowers-subagent-driven-development (if subagents are suitable per AGENTS.md) or superpowers-executing-plans to implement this plan. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** `sdlc move [:N]` moves the current slot's issue branch into slot `:N` (default `:0`), replacing the #248 manual procedure.

**Architecture:** Observe both slots through git into a plain `moveFacts` value, decide with a pure `checkMove`, then perform two `git switch` calls. Observation runs twice (preflight and again just before switching) and the move refuses if the two differ. Git remains the final collision guard (`--no-overwrite-ignore`).

**Tech Stack:** Go, cobra, real-git fixtures (`cmd/sdlc/internal/testfix`, `procedureFixture`).

Issue: `workshop/issues/000260-slot-move.md` (Spec and Done when are the contract).

---

## Core concepts

### Pure entities

| Name | Lives in | Status |
|------|----------|--------|
| `moveSide` | `cmd/sdlc/moveplan.go` | new |
| `moveFacts` | `cmd/sdlc/moveplan.go` | new |
| `checkMove` | `cmd/sdlc/moveplan.go` | new |
| `untrackedCollisions` | `cmd/sdlc/moveplan.go` | new |

- **moveSide** — one slot as observed: root, address, branch, resting branch,
  HEAD, tracked/staged/submodule changes, untracked paths, in-progress Git
  operation.
  - **Relationships:** `moveFacts` holds two (From, To).
- **moveFacts** — everything the decision needs: both sides, whether they share
  `repo_identity`, the paths branch A checks out, and the destination rest's
  commits missing from A (`Parked`) and from its upstream (`Unpublished`).
  Comparable with `reflect.DeepEqual`, which is how the second observation is
  checked against the first.
- **checkMove(facts, acceptParked) error** — all refusal rules in one place,
  returning the first failure as an actionable message.
  - **DRY rationale:** the preflight and the dry run use the same rules; the
    #248 atlas prose shrinks to a pointer instead of restating them.
- **untrackedCollisions(untracked, incoming) []string** — the untracked paths
  that the incoming tree would overwrite: the same path, an untracked file where
  A has a directory, or an untracked path under a directory that A has as a file.
  - **Future extensions:** if `sdlc merge`'s return to rest ever needs the same
    check, it reuses this function.

Unit tests: `cmd/sdlc/moveplan_test.go`, no IO.

### Integration points

| Name | Lives in | Status | Wraps |
|------|----------|--------|-------|
| `observeMove` | `cmd/sdlc/move.go` | new | `git` via `execGitRunner`, `workspace.Resolve` |
| `runMove` | `cmd/sdlc/move.go` | new | the two `git switch` calls, output |
| `NewMoveCmd` | `cmd/sdlc/move.go` | new | cobra; `markMutatingCommand` |
| `gitOperationInProgress` | `cmd/sdlc/issuemovedetail.go` | modified | takes a git func and a root instead of `*trackerEnv` |

- **observeMove(dir, address)** — resolves both identities, runs
  `status --porcelain=v1 -z --untracked-files=all --ignore-submodules=none` in
  each root, checks operation markers, lists A's tree, and runs the two
  destination `log` ranges.
  - **Injected into:** nothing; it produces `moveFacts` for `checkMove`. Tested
    through `runMove` against real git (the procedure fixture), no mocks.
- **gitOperationInProgress** — widened from `*trackerEnv` to
  `(git func(...string) (string, error), root string)` so `move` reuses it;
  `issue move-detail` passes `env.git, env.root`.
- The repository lock (`markMutatingCommand`) lives in the Git common
  directory, which all slots share, so one lock covers both.

### Decisions

- **Post-move build is printed, not run.** #248's post-move build is prose in the
  destination repository's `AGENTS.local.md` (pair: "run `make build` there").
  The command prints it as the next step and notes that running sessions keep
  their old binaries. Making it executable needs a machine-readable
  declaration, which is out of scope (Simplicity First). Recorded as a Spec
  revision.
- **`--accept-parked`** is the explicit acceptance of destination resting
  commits that A lacks (#248 step 3's "temporary smoke test" choice).
- **Replaced tests:** `TestWorkspaceProcedureMoveBranchToPrimaryPreservesParkedMainAndScratch`
  and `TestWorkspaceProcedureMoveRejectsIncomingUntrackedCollision` exercised
  the manual procedure and parsed its atlas shell block. `move_test.go`
  covers the same behavior through the command, so both are deleted together
  with the atlas steps they read.

## Chunk 1: pure decision

### Task 1: collisions and move rules

**Files:**
- Create: `cmd/sdlc/moveplan.go`
- Test: `cmd/sdlc/moveplan_test.go`

- [ ] **Step 1: failing tests.** `TestUntrackedCollisions` (table): same path;
  `a` untracked vs incoming `a/b`; `a/b` untracked vs incoming `a`; `ab` vs
  `a/b` (no false prefix match); none. `TestCheckMove` (table built from a
  valid `moveFacts` base, one mutation per row, asserting an error substring or
  nil): valid → nil; different repo; same root; source detached; source on its
  resting branch; source changes; source untracked; source operation;
  destination not on resting; destination changes; destination operation;
  destination untracked colliding (names the path); destination untracked not
  colliding → nil; parked non-empty → refuses naming `--accept-parked`; parked
  non-empty with accept → nil.
- [ ] **Step 2:** `go test ./cmd/sdlc -run 'TestUntrackedCollisions|TestCheckMove'` → FAIL (undefined).
- [ ] **Step 3: implement.**

```go
package main

import (
	"fmt"
	"sort"
	"strings"

	"github.com/xianxu/ariadne/cmd/sdlc/internal/gitx"
)

// moveSide is one slot as `sdlc move` observed it (#260).
type moveSide struct {
	Root, Address, Branch, Resting, Head string
	Changes                              []gitx.StatusEntry // tracked, staged, submodule
	Untracked                            []string
	Operation                            string
}

// moveFacts is everything checkMove decides on. It is compared with
// reflect.DeepEqual to prove nothing changed between preflight and switch.
type moveFacts struct {
	From, To      moveSide
	SameRepo      bool
	Incoming      []string // paths branch From.Branch checks out
	Parked        []string // To.Resting commits absent from the branch
	Unpublished   []string // To.Resting commits absent from its upstream
}

// checkMove holds every refusal rule of `sdlc move`; nil means the two
// switches may run.
func checkMove(f moveFacts, acceptParked bool) error {
	from, to := f.From, f.To
	switch {
	case !f.SameRepo:
		return fmt.Errorf("%s and %s are different repositories", from.Root, to.Root)
	case from.Root == to.Root:
		return fmt.Errorf("%s is already the current slot", to.Address)
	case from.Branch == "":
		return fmt.Errorf("%s is detached; switch to the issue branch to move", from.Root)
	case from.Branch == from.Resting:
		return fmt.Errorf("%s is on its resting branch %s; there is no issue branch to move", from.Address, from.Resting)
	case to.Branch != to.Resting:
		return fmt.Errorf("%s is on %s, not its resting branch %s; move that branch out first", to.Address, orDetached(to.Branch), to.Resting)
	}
	for _, s := range []moveSide{from, to} {
		if s.Operation != "" {
			return fmt.Errorf("%s has a Git operation in progress (%s); finish or abort it first", s.Address, s.Operation)
		}
		if len(s.Changes) > 0 {
			return fmt.Errorf("%s has uncommitted changes (%s); commit or remove them first", s.Address, statusLine(s.Changes[0]))
		}
	}
	if len(from.Untracked) > 0 {
		return fmt.Errorf("%s has untracked files (%s); commit or remove them first", from.Address, from.Untracked[0])
	}
	if c := untrackedCollisions(to.Untracked, f.Incoming); len(c) > 0 {
		return fmt.Errorf("%s has untracked files that %s would overwrite: %s; move or remove them first", to.Address, from.Branch, strings.Join(c, ", "))
	}
	if len(f.Parked) > 0 && !acceptParked {
		return fmt.Errorf("%s has %d commit(s) that %s lacks, which the move would leave out of the test:\n  %s\npublish them and rebase %s first, or rerun with --accept-parked for a temporary smoke test",
			to.Resting, len(f.Parked), from.Branch, strings.Join(f.Parked, "\n  "), from.Branch)
	}
	return nil
}

func orDetached(branch string) string {
	if branch == "" {
		return "a detached HEAD"
	}
	return branch
}

// untrackedCollisions returns the untracked paths that checking out incoming
// would overwrite: the same path, or one being a directory prefix of the other.
func untrackedCollisions(untracked, incoming []string) []string {
	var out []string
	for _, u := range untracked {
		for _, p := range incoming {
			if u == p || strings.HasPrefix(p, u+"/") || strings.HasPrefix(u, p+"/") {
				out = append(out, u)
				break
			}
		}
	}
	sort.Strings(out)
	return out
}
```

- [ ] **Step 4:** rerun → PASS.
- [ ] **Step 5:** commit `#260: move: pure refusal rules and collision check`.

## Chunk 2: command

### Task 2: widen gitOperationInProgress

**Files:** Modify `cmd/sdlc/issuemovedetail.go:68-86,106`.

- [ ] Change the signature to `gitOperationInProgress(git func(...string) (string, error), root string)`,
  replacing `env.git`/`env.root` inside; the caller passes `env.git, env.root`.
- [ ] `go test ./cmd/sdlc -run MoveDetail` → PASS. Commit with Task 3.

### Task 3: observe, switch, verify

**Files:**
- Create: `cmd/sdlc/move.go`, `cmd/sdlc/helptext/move.md`
- Modify: `cmd/sdlc/main.go` (register after `workspace`)
- Test: `cmd/sdlc/move_test.go`

- [ ] **Step 1: failing tests** (all on `procedureFixture`; slot 1 is the
  source, on `000001-procedure` with one committed file `feature`):
  - `TestMoveToPrimary`: scratch file in `:0`; `runMove(slot1, "", …)` →
    `:0` on the branch at the recorded HEAD, slot 1 on `main-slot1` at its
    old HEAD, scratch file unchanged, output names the post-move step.
  - `TestMoveToSlot`: target `:2` → same checks.
  - `TestMoveRefusals` (table; each row asserts an error substring and that a
    snapshot of both slots — `status --porcelain=v1 --untracked-files=all`,
    `show-ref`, `symbolic-ref HEAD` — is unchanged): missing slot `:5`; `:0`
    tracked change; `:0` on another branch; `:0` untracked `feature`
    (collision); slot 1 tracked change; slot 1 untracked file; slot 1 on its
    resting branch; `:0` resting commit absent from the branch.
  - `TestMoveAcceptParked`: that last case with `acceptParked` → moves; `main`
    still at the parked commit.
  - `TestMoveDryRun`: prints the plan, snapshot unchanged.
  - `TestMoveSecondSwitchFails`: `:0` has ignored `build/output`; the branch
    force-adds `build/output` → error names the retry
    (`git -C <:0 root> switch 000001-procedure`), the branch ref is still at
    the recorded HEAD, slot 1 on `main-slot1`, `:0` on `main`, the ignored file
    unchanged.
- [ ] **Step 2:** `go test ./cmd/sdlc -run TestMove` → FAIL.
- [ ] **Step 3: implement `move.go`.**

```go
// move.go — `sdlc move [:N]` moves this slot's issue branch into slot :N
// (default :0) and returns this slot to its resting branch (#260).
package main

// NewMoveCmd: Use "move [address]", MaximumNArgs(1), flags --dry-run and
// --accept-parked, wrapped in markMutatingCommand; RunE calls
// runMove(".", address, accept, dry, stdout, stderr).

// runMove:
//  1. facts := observeMove(dir, address); checkMove(facts, accept).
//  2. Print: "Move <branch> from <from.Address> to <to.Address>", then
//     Unpublished commits (if any) as information and Parked commits when
//     accepted. --dry-run stops here.
//  3. again := observeMove(...); !reflect.DeepEqual(facts, again) → refuse
//     "slots changed since the preflight; rerun".
//  4. git -C from -c submodule.recurse=false switch --no-overwrite-ignore <from.Resting>
//  5. git -C to   -c submodule.recurse=false switch --no-overwrite-ignore <branch>
//     on failure: return an error with git's output, stating the branch ref is
//     intact at <head>, <from.Address> is on <from.Resting>, and the retry is
//     `git -C <to.Root> switch <branch>` after reconciling <to.Address>.
//  6. Verify with observeMove-style reads: to on branch at facts.From.Head,
//     from on its resting branch at its pre-move resting HEAD.
//  7. Print the next step: "Run <to.Root>'s post-move build from its
//     AGENTS.local.md, if it declares one. Running sessions keep the old
//     binaries."

// observeMove(dir, address):
//  - from := workspace.Resolve(r, dir, ""); to := workspace.Resolve(r, dir, address or ":0").
//  - side(identity): Root=WorktreeRoot; Address/Branch/Resting/Head via
//    workspaceText(..., ""); status -z parsed by gitx.ParseStatusZ with
//    "??" entries → Untracked, the rest → Changes;
//    Operation = gitOperationInProgress(trimmed GitInDir(root), root).
//  - SameRepo = from.RepoIdentity == to.RepoIdentity.
//  - When from.Branch != "": Incoming = ls-tree -r -z --name-only <branch>.
//  - When to.Resting != "": upstream = rev-parse --abbrev-ref
//    --symbolic-full-name <rest>@{upstream} (error → refuse: "<rest> has no
//    configured upstream"); Parked = log --format=%h %s <branch>..<rest>;
//    Unpublished = log --format=%h %s <upstream>..<rest>.
```

  Help text `helptext/move.md`: usage lines (`sdlc move`, `sdlc move :2`,
  `--dry-run`, `--accept-parked`), what it checks, that it never stashes,
  resets, deletes or pushes, and moving back (`sdlc move :1` from the other
  slot). Register: `add(NewMoveCmd(), "move", "Move this slot's issue branch into another slot (default :0)")`.
- [ ] **Step 4:** `go test ./cmd/sdlc -run 'TestMove|MoveDetail|Help'` → PASS.
- [ ] **Step 5:** commit `#260: move: sdlc move switches the issue branch between slots`.

### Task 4: docs and replaced procedure tests

**Files:**
- Modify: `atlas/workflow/workspace-branching.md` ("Move this branch to :N"),
  `README.md` (near the `sdlc workspace` section), `atlas/index.md` if a new
  file is added (none planned).
- Modify: `cmd/sdlc/workspace_procedure_test.go` (delete the two move tests).

- [ ] Rewrite the atlas section to: run `sdlc move [:N]` (`--dry-run` first);
  what it refuses and why; the operator decision `--accept-parked` stands for;
  the post-move build and relaunch; moving back. Keep only what a person
  decides.
- [ ] Delete the two replaced tests; `go test ./cmd/sdlc/...` → PASS.
- [ ] Commit `#260: atlas: point the branch move at sdlc move`.

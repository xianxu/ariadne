# Lessons

Compact guidance for Ariadne's SDLC, generated contracts, and base-layer work.
Detailed incidents remain in their owning issue or review artifact.

## Contracts and single sources

- A prose policy becomes an integration contract when code reads the repository.
  Put the contract in one referenced source, give it a machine-readable token,
  and make consumers gate on that token rather than prose presence.
- Generate projections from the source model. A hand-maintained copy of generated
  data, an enum registry, a plan table, or an atlas list drifts silently.
- A declaration nothing reads is a comment with a type. Trace every field,
  capability, and status through construction, persistence, and consumers.
- A target can lie by aspiration. Mark only mechanisms actually implemented and
  tested; do not generalize a high-clarity proof to unbuilt siblings.
- When merging or removing concepts, hunt every consumer before deleting a
  symbol. A follow-up issue must not offload the current issue's purpose.

## Guards and proof

- A guard must fail closed on malformed, unknown, unsupported, and ambiguous input.
  A catalogued flag may waive only the refusal named by that catalog entry.
- A syntactic guard cannot support an absolute semantic claim. Bound the claim to
  the language actually inspected and state what remains outside it.
- Distinguish what a guard computes from what it takes on faith. Negative scans
  and blind reads cannot prove absence; preserve query failure as failure.
- A guard test must have teeth: mutation-check the guard and make the test invoke
  the production entry point. A helper test or a happy end state can be vacuous.
- Enumerate the real population from the tree, not from the finding's examples or
  a comment. When a helper is extracted, prove its caller still traverses it.
- A gate the agent can skip is not a gate. Put the critical transition in the
  binary and make its refusal actionable without a bypass.
- Optional evidence stays optional only when unavailable is distinguishable from
  unsupported and from a negative result.

## Paths, processes, and integration

- A best-effort projection must not add a refusal to the verb that hosts it.
  If a cosmetic refresh's input cannot be read, the output should degrade to
  the unrefreshed bytes and stay deterministic for any proof; it should not
  fail closed and wedge a landing (#275).

- A changed path or residency includes search recipes, generated references,
  comments, and execution records. Sweep the old location, not only feature prose.
- A peer migration starts by checking branch, cleanliness, and the real fleet;
  never clean another repo or assume “nothing to migrate” from one local view.
- Command wrappers that call `os.Exit` cannot be made best-effort by their caller.
  Return errors from libraries; let the command boundary choose severity and
  perform deferred cleanup, including init races and hard exits.
- Agent-invoked CLI verbs must run headless and gate on durable state, not local
  convenience or an interactive editor.
- A read-only check must not resolve writer-side configuration (a publication
  target, an upstream) just to learn where to read. CI checkouts are detached with
  no local branch config; take the source explicitly from the caller (#266).
- A live integration test must exercise filesystem, Git, process, or network
  behavior with real data. A fake keyed to the same value shape masks IO bugs.
- A silent `0` or empty result is a footgun when it is indistinguishable from a
  real answer. Return status and provenance that explain what was observed.
- `cd` into a temp workspace must hard-guard the path; an empty variable must not
  fall through to the host repository. Glob paths directly instead of using
  `ls` in command substitution.

## Review, plans, and artifacts

- When a fix round changes a contract, grep every restatement of it (atlas,
  README, the help page of every verb whose behavior changed, plan steps and
  tables, Revisions) and update them in the same commit. Registered flags are
  checked against their FLAGS section by `TestEveryFlagAppearsInItsHelp`. A fixed rule documented the old way reads as two rules (#277).
- Absence is not evidence of a transfer. Authorize a state change from
  positive evidence written by the actor entitled to make it, never from
  "the other party no longer holds X" (#277 BR-9).

- Acquisition and later discovery must accept the same topology; test re-entry
  from every provisioned repository shape, including linked feature worktrees.
- Independent clones need clone-specific feature-worktree destinations; a fleet
  path keyed only by repository and branch collides across environments.

- New CLI surfaces need discoverable README usage as well as detailed atlas and
  embedded help documentation; check all three before closing.
- User-runnable procedures need the same README discovery link, even when they
  introduce no command. Preservation tests must snapshot both source and destination;
  snapshot after intentional fixture movement rather than skipping that ref.
- Before a boundary, read the actual verdict and reconcile the plan against the
  committed tree. “Inert at runtime” is not the same as “removed”; an unticked
  row can disable the guard that proves it.
- Multi-milestone inventories distinguish delivered entities from future files
  and modifications. Separate completed implementation checklists from the
  acceptance gate whose successful execution is still pending.
- Check generated review artifacts against the complete boundary range before
  committing; an unstaged-only whitespace check misses already committed prose.
- Review snapshots enumerate every mutable artifact in their prompt and remain
  bounded. Do not truncate judge output with `tail` or `grep`.
- A close claim names evidence at the corrected mutation boundary. A plan revision
  edits every affected row, signature, citation, and status field.
- Atlas updates land in the milestone that introduces the surface. After a symbol
  moves, grep the old location and refresh downstream links and search recipes.
- Record actual hours from the measurement tool. Commit issue/plan syncs before
  long-running reviews so design survives compaction.
- Stage explicit paths; never let `git add -A` sweep unrelated untracked WIP.
  Verify `change-code` created the intended branch before committing.

## Base-layer and data sharp edges

- Preserve raw Git path output through adapters; trimming whitespace before the
  path parser loses valid path bytes. Detect lexical environment candidates
  before canonicalization so a symlink cannot turn a refusal into generic setup.
- Private dependency scans must accept every name acquisition accepts. Global
  fleet backup/dot-directory heuristics are not a private-repository contract.

- Validate external identifier grammar before interpreting sentinel values. An
  all-zero Git OID is unborn only at a valid hash length; malformed zeros must
  not bypass validation by becoming a null value.
- Editing `construct/adapted/` changes rendered downstream instructions. Inspect
  generated output and downstream consumers before landing a base-layer change.
- A span edit is only safe if its boundary is explicit and tested. A pathspec is
  repository-relative even when the configured root is absolute.
- `strings.TrimSpace` on a whole porcelain-status blob removes the first status
  column; parse fields instead of slicing columns.
- CUE schemas need whole-corpus validation, not only self-vetting examples. State
  machine gates must cover every legitimate flow before tightening a status enum.
- User strings, slugs, and symlinked paths need serialization and canonical-path
  tests. Resolve logical and physical paths before comparing them.
- Generated review sidecars and durable records need bounded sizes and explicit
  framing; raw transcripts do not belong in source artifacts.

## Working rule

When designing a gate, name the source of truth, the exact closed grammar, the
production transition it controls, and the mutation that would make its test red.
The simplest durable authority beats a clever scan of consequences.

- Git candidate reachability proves remote state, not which caller won: concurrent
  writers can construct identical commits. Reservation retries must distinguish
  actual ref updates, up-to-date responses, confirmed rejection, and uncertainty.
- Race conformance must exercise both client stale-lease rejection and server
  receive-pack ref-lock rejection; both occur in real concurrent Git publication.

- Review cancellation must be wired from actual CLI signals through command
  contexts to owned subprocesses. A background-context entry point can orphan
  a reviewer after its repository lock is released, despite unit cancellation tests.
- Before recording any review verdict, revalidate the complete prepared read set,
  including absent plans and ledger generations; rejected verdicts also write state.

- When reusing a publication guard, inspect its hidden destination and selection
  assumptions. Pass the pinned configured trunk and owned close records; an
  issue-body diff can omit independently published review evidence.
- Bind GitHub identity to Git's effective fetch and push URLs, including rewrites.
  Revalidate the selected branch and head as well as destination configuration.
- Archive retries need an intact complete generation: preserve compatible prose,
  refuse changed lifecycle metadata or missing owned artifacts, and verify the
  exact reachable archive commit before deleting local recovery refs.

- Scope cleanliness to effects: integration/switching need a clean issue checkout,
  while proven ref-only cleanup on rest can preserve new staged and unstaged work.
  Test both immediate post-switch edits and retries after each cleanup effect.

- A cancelled or output-limited Git command has no trustworthy predicate exit code;
  keep those errors distinct from completed `exit 1` absence/ancestry results.
- Match writer limits to reader limits, including framing overhead; a successful
  write must not make the next authoritative read refuse its own data.
- A bounded writer must not accidentally expose an embedded `ReadFrom` fast path
  that bypasses its `Write` limit. Test the real subprocess/`io.Copy` boundary.
- Authoritative Git blob identities require exact bytes: bypass clean filters on
  writes and archive export filters on reads; test caller and global attributes.
- Rebase overwrite checks must cover paths touched by intermediate replay commits,
  including files deleted again before HEAD, not only the target and final trees.
- When a procedure promises two history comparisons, show and test both Git
  commands. State which local tracking evidence may be stale before calling
  commits unpublished.
- Cross-repository requirements need a named issue and review boundary in the
  owning repository; a single repository's pinned diff cannot prove a peer edit.
- Tests for executable agent instructions must consume the authored instruction
  when practical; duplicating its commands in a fixture cannot catch a missing
  step in the guide.
- When a new operator phrase is added to an agent guide, update the existing
  README entry point for that workflow in the same change.
- Before submitting a milestone, enumerate the non-test files added in its window
  (`git diff --name-status BASE HEAD | grep '^A' | grep -v _test`) and reconcile
  them with the plan's Core concepts table and each Task "Files:" list; reviewing
  the inventory by memory missed six entities and two phantom files (#252 M2).
- When a verb's behaviour changes, `rg` the verb's name across
  `cmd/sdlc/helptext/` and `atlas/` and classify every hit (still true / now
  false / legacy until cutover) in the same change: per-verb help is the workflow
  contract agents read, and fixing only the pages you remember left four stale
  descriptions (#252 M2).
- Checks run before effects: a verb's dry-run and refusal paths must leave every
  file unchanged, so compute refreshes in memory and write only after the last
  check, including building any receipt or operation that validates its input
  (#252 M2: a stale-mirror refresh was written before refusals ran).
- An observation error is never evidence of absence: a Git probe distinguishes
  exit 1 (false) from any other failure (#252 M2).
- Stage explicit paths, never a directory (`git add cmd/sdlc` committed a 9 MB
  build artifact and an unrelated gofmt rewrite in #252 M3); build binaries to
  `-o` outside the source tree.
- A landing's identity is what the forge confirmed (the PR's merge commit), not
  ancestry of the branch's commits: squash and rebase rewrite them (#252 M3).
- A deferred effect replays inputs pinned when it was decided, never the live
  worktree a later commit may have changed (#252 M3) — and replays them
  three-way: only where the target still holds the version the decision saw.
  A later committed edit is newer than the pin; overwriting it is data loss
  (#252 M3 BR-30). Every pinned input the replay does not apply is reported
  (path + reason, durably where possible), never dropped silently.
- A cleanup path shared by landing and abandonment completes nothing: gate each
  completion on the confirmed landing itself, not on reaching the step after it
  (#252 M3 BR-29: removing an unmerged worktree marked its cards done).
- Every identifier in a ticked plan row greps to shipped code. When a name
  changes in implementation, rename it in the row (planned name in parentheses)
  as well as in Revisions (#252, 4th inventory-drift finding).
- Fixtures must not normalize what production callers don't: resolving a temp
  root's symlinks up front hid close comparing /tmp and /private/tmp paths
  until an e2e used the raw slot paths (#252 M4). Compare filesystem paths
  canonically on both sides (`canonRoot`).
- Before designing a transformation of existing records, run the strict
  parser over the real population read-only: a fleet probe showed a fifth of
  issue files failing it and duplicate IDs, which reshaped the migration
  (#252 M4). Then rehearse the whole command on a disposable copy.
- An artifact that pins another's identity (a blob OID, a digest) is built by
  one constructor from the source's final bytes, after its last mutation — and
  the pure plan asserts pin == identity(final). Deriving the pin alongside the
  first version and mutating the source afterwards (binding a close to a card)
  left main naming a blob the tracker never held (#252 M4 BR-34).
- An end-to-end test that skips a gate (`--no-judge`, `NoJudge: true`) also
  skips every bug behind it: the #252 slot-cycle e2e and the first smoke cycle
  both bypassed the publish gate, so the transfer guard refusing the owner's
  own edits on a direct push surfaced only in the second smoke cycle. Run the
  deterministic gates in e2e; stub only the LLM.
- "Both sides make the identical change, so the merge is clean" holds only
  until either side edits that region again; durable convergence needs the
  change as a shared ancestor (merge it), not a byte-identical copy. #252's
  first reconcile committed main's conversion bytes and conflicted at the
  first mirror refresh (#252, leftover-branch matrix).
- When a rollout changes shape (big-bang → incremental), re-audit every
  premise the old shape let you drop: #252 M2 made verbs tracker-only for a
  simultaneous cutover, the per-repo rollout silently needed them legacy-capable,
  and only a live soak found it. For "behaves like the old binary" claims, test
  differentially against the old binary (`testdata/legacy-equivalence.sh`).
- A fix to a helper shared by both modes needs checking in both: #252's
  tracker-mode `TrunkRef` fix (local main behind the trunk) silently broke
  legacy mode (local main ahead of it), caught only by re-running
  `legacy-equivalence.sh` before close. When a fix picks one ref over another,
  enumerate every ordering of the two (behind, ahead, diverged) and test each.
- A mode decision is itself a read, and must degrade like one: #252's
  legacy/tracked selector asked the remote with no offline fallback, so a
  verb that never needed the network before started needing it just to learn
  which code path to run. When adding a gate in front of an existing local
  verb, ask what it does offline, and test it with an unreachable remote.
- Never slice `git status` output at fixed columns when the runner trims:
  the first entry's leading status space (" M path") is data, and trimming
  shifts its path by one. Read status as `--porcelain=v1 -z` through
  `gitx.ParseStatusZ` from untrimmed output. Hit twice in one day (#255's
  migrate dirty list; #259's close evidence silently dropped the gate ledger).
- A generated block inside a tracked file (weave's `.gitignore` block) is
  shared across checkouts, but the inventory that justifies each line is
  per-checkout. Rebuilding the block from local state alone erased the lines
  another checkout's compile committed (#263). Remove only what local
  provenance proves you own. And before adding a migration exception, check
  in git history where the legacy data actually lived: #263's first draft
  exempted legacy entries "inside the block", which never existed there.
- A fleet-wide "all migrated" claim must come from a sweep, not a worklist:
  #241's log declared every ariadne-layer repo adopted while parley.nvim and
  the three brains had never been touched (the Spec's "brain repos retain
  their commit rhythm" was misread as an exclusion). Before claiming fleet
  completion, enumerate every repo under the workspace and check the end
  state directly (inventory present, managed block, no owned link still
  tracked, one workflow include). Same for "no atlas surface": search the
  atlas for the component's name, not just the change's vocabulary (#264).
- A multi-pass writer to one shared file must let only the pass that owns a
  migration perform it: weave's data pass stripped the legacy ignore list
  that only the later artifacts pass could replace, so a refused artifacts
  pass left generated files exposed (#264).

- A protocol compatibility exception needs behavioral probes for both admission
  and refusal: confirmed old mode, restored state, retargeted state, and unknown
  provenance. New-format happy paths cannot prove an old-format exception.

- Don't infer whose work a commit is from git topology. Neither ancestry (a
  parent and its child share the parent's commits) nor a branch name (the
  owner can commit elsewhere) says who owns it; read the `#N` subject tag.
  #272's first start-plan guard refused a parent once it advanced past its
  child's fork, and #274's handoff guard refuses owner edits made off the
  owner's branch.

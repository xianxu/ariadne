# Boundary Review — ariadne#152 (whole-issue close)

| field | value |
|-------|-------|
| issue | 152 — Switch git remotes from SSH to HTTPS so sdlc verbs run in-sandbox |
| repo | ariadne |
| issue file | workshop/issues/000152-switch-git-remotes-from-ssh-to-https-so-sdlc-verbs-run-in-sandbox.md |
| boundary | whole-issue close |
| milestone | — |
| window | a5f6873e096d01f6fdf5515faba434af07b103e7..HEAD |
| command | sdlc close --issue 152 |
| reviewer | claude |
| timestamp | 2026-07-01T10:49:42-07:00 |
| verdict | SHIP |

## Review

Shadow-sweep complete. `insteadOf` lives in exactly one automated place (the container overlay `setup.sh`); the host is a documented one-time step; `bootstrap-peers.sh` is a post-clone operator one-shot with no config hook, confirming the implementor's justification. Everything verified end-to-end.

```verdict
verdict: SHIP
confidence: high
```

This is a tight, correct, well-documented fix. I reproduced the original bug and validated the fix independently: without `--add`, the second `git config` line clobbers the first so **only** `ssh://git@github.com/` survives (leaving `git@github.com:` — the form real origins actually use — on SSH); the new `--unset-all` + two `--add` form yields **both** rewrites, is idempotent across re-runs (no duplicate accumulation), tolerates the empty-key case (`--unset-all` exits 5, absorbed by `|| true`), correctly rewrites both SSH forms to HTTPS, and leaves `gcrypt::ssh://` brain remotes untouched — exactly matching the documented caveat. The atlas doc accurately mirrors the code, records the durability decision, and flags the gcrypt follow-up. Nothing blocks SHIP; the notes below are all Minor/future.

**1. Strengths**
- The core fix is correct and I verified it against real `git config` semantics, not just the commit message: the `--add` multi-value bug is real and the fix resolves it (`setup.sh:17-19`).
- Idempotency was handled deliberately — `--unset-all … || true` before re-adding prevents duplicate-value accumulation on `make sandbox-clean` re-runs (`setup.sh:17`). Easy to have missed; it's right.
- The inline comment (`setup.sh:10-15`) and atlas section (`openshell-sandbox.md:41-73`) both explain *why* `--add` is required, so the next editor won't silently reintroduce the clobber.
- Scope discipline: the encrypted `gcrypt::ssh://` brain remotes are correctly identified as non-matching and left alone, with the sensitive transport switch explicitly deferred rather than half-done (`openshell-sandbox.md:69-73`).
- The Revisions section honestly records that implementation *narrowed* the problem (container path already existed; real gap was the host config) rather than papering over the pivot.

**2. Critical findings** — none.

**3. Important findings** — none.

**4. Minor findings**
- `atlas/workflow/openshell-sandbox.md` is not linked from `atlas/index.md` (which does list its sibling `workflow/sandbox.md:14`). Pre-existing since #44, not introduced here — but this boundary added a whole new `## Git transport` section, so it's the natural moment to add the index link (AGENTS.md §8: "Keep atlas/index.md linking every file").
- The **host** command block in the atlas (`openshell-sandbox.md:59-62`) omits the `--unset-all` line that the container `setup.sh` uses, so it isn't idempotent — an operator who runs it twice accumulates duplicate `insteadOf` values. Functionally harmless (both values rewrite to the same HTTPS URL, git raises no error), and it's labeled "one-time," so low stakes; consider matching the container's clear-then-add form for symmetry.
- The stated rationale for keeping the host step manual — "bootstrap.sh clones peers *before* any config step could run" (`openshell-sandbox.md:64-67`, issue Revisions) — is a bit imprecise: the host clone works over SSH regardless of ordering, so timing isn't the real blocker. The genuinely correct reason to keep it manual is **blast radius** — baking a global transport rewrite into a base-layer script like `bootstrap-peers.sh` would silently flip git transport for *every* downstream user (sandbox or not). The decision is right; the "before clone" justification just isn't the load-bearing one.

**5. Test coverage notes**
- `.openshell/overlay/setup.sh` is a provisioning overlay with no automated test harness, consistent with the rest of the file. The shipped bug (missing `--add`) is exactly the class a test could have caught, but standing up a test rig for the overlay is disproportionate here and not the repo's convention. I substituted manual verification in an isolated config file (bug reproduced, fix + idempotency + rewrite + gcrypt-exclusion all confirmed), which is adequate proof for this boundary.

**6. Architectural notes for upcoming work**
- ARCH-DRY — **pass.** `insteadOf` appears in exactly one automated location (the container overlay); the host restatement is documentation of a manual step, not duplicated logic. No shared helper is warranted (container vs. host are different execution contexts that can't share code).
- ARCH-PURE — **pass (N/A).** This is boundary/IO glue (a shell provisioning script); there's no business logic to separate.
- ARCH-PURPOSE — **pass, with a noted deferred consumer.** Shadow-sweep of the consumers: (a) OpenShell container — *enforced/automated* via `setup.sh` ✓; (b) host `~/.gitconfig` for the Claude Code sandbox — delivered as an *applied* runtime machine change + a documented one-time operator step, not automation; (c) `gcrypt::ssh://` brain remotes — explicitly deferred as a separable, sensitive follow-up ✓. The host side is the one non-enforced consumer. That's a legitimate durability decision here (auto-applying a global transport rewrite in a base-layer script would over-reach onto all downstream users), and the issue's Spec listed "one-time operator step vs. baked into setup" as an open question it resolved deliberately — so this is a bounded deferral of a genuinely separable concern, not an under-delivery of the point. **Future:** if a host post-clone setup hook ever materializes (e.g. a `construct/setup.sh` host phase), the host `insteadOf` should migrate into it to close this manual-step gap so fresh hosts derive it automatically.

**7. Plan revision recommendations** — none. The plan, Revisions, and Log accurately match the code and the verified behavior; no drift to reconcile.

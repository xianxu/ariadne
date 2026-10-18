# Boundary Review — ariadne#447 (whole-issue close)

| field | value |
|-------|-------|
| issue | 447 — review checkouts lack main ref and remote config, so repo tests needing main authority can't run in review |
| repo | ariadne |
| issue file | workshop/issues/000447-review-checkouts-lack-main-ref-and-remote-config-so-repo-tests-needing-main-authority-can-t-run-in-review.md |
| boundary | whole-issue close |
| milestone | — |
| window | 60ed08fd6e643c0bd38cdbe164c39497224c7766..126eb88b52d3a8b4b671895272ba6b2214d418e1 |
| command | sdlc close --issue 447 |
| reviewer | claude (anthropic; author anthropic) |
| isolation | fenced |
| timestamp | 2026-10-18T14:30:03-07:00 |
| verdict | SHIP |

## Review

```verdict
verdict: SHIP
confidence: high
```

This change does what the plan said it would. The review checkout now gets the slot's `branch.main.{remote,merge}` settings, the fetch refspec of each named remote, and the main refs (`refs/heads/main` and each `refs/remotes/<R>/main`). Each copied remote's URL and push URL is set to the inert `readonly://review-checkout`, so the checkout cannot push or fetch. `go test ./cmd/sdlc/internal/reviewbox/` passes. To check the regression test, I turned off the `copyMainAuthority()` call in `populate`. `TestCheckoutCarriesMainAuthorityReadOnly` then failed (`config --get-regexp ^branch\.main\.` exit 1), and passed again once the call was restored. The checkout was left clean. Nothing blocks shipping. I found two Minor edge cases.

1. **Strengths**
   - The pure/IO split is clean. `mainAuthority` (`authority.go:35`) is a pure parser tested in a table with no IO. `copyMainAuthority` (`checkout.go:139`) is a thin layer of git calls around it.
   - The allowlist is enforced twice. The `--get-regexp` pattern selects only main-authority keys, and the parser rejects any other key rather than skipping it. Tests cover a URL, an `insteadOf` and a push URL.
   - Read-only comes from structure, not convention. The inert URL has no remote helper, and the test checks that push and fetch both fail and that neither the slot nor its bare origin changes.
   - Placement is right. The copy runs after the detached checkout, so `update-ref refs/heads/main` never moves a checked-out branch. `Reset` reuses `populate`, so a retried attempt gets the same authority.
   - The regression test sets up the real pair#665 shape: an issue branch whose main has moved on. It checks the merge base and `main@{upstream}` against the slot's own answers.

2. **Critical:** none.

3. **Important:** none.

4. **Minor**
   - **Nested branch named `main` (`authority.go:49`):** a remote branch like `refs/remotes/origin/feature/main` passes the suffix check. The checkout then gets a stray remote `origin/feature` with an inert URL. I confirmed git accepts `remote.origin/feature.url`. It is harmless (inert URL, and the slot has the same refs) but adds a remote the slot doesn't have. Fix: accept only remote names without a `/`.
   - **Config read scope (`checkout.go:140`):** `git config --get-regexp` reads system, global and local config. A user-level `branch.main.*` or `remote.*.fetch` would be copied as if it were the slot's. A valueless key anywhere in scope makes checkout creation fail, which blocks the whole review. Adding `--local` would limit the copy to the repository and make it match the stated "slot's authority".

5. **Test coverage notes:** the pure parser has 9 table rows, including hostile keys and malformed rows. The integration tests use real git in temp repos, which is the package's existing pattern. Pair's actual repolint isn't run; the test checks the three facts it reads, which is enough evidence for this repo.

6. **Architecture**
   - **ARCH-DRY:** pass. It reuses the `git`/`gitOutput` helpers.
   - **ARCH-PURE:** pass.
   - **ARCH-PURPOSE:** pass. Both Done-when items are delivered and tested. The Spec says "main ref at the reviewed base", but the code copies the slot's current main. That matches the Spec's own aim ("same main authority a slot checkout has"), and the Plan states it explicitly.
   - **ARCH-MOCK:** pass. Real git in temp repos is this package's established seam, with no network.
   - **ARCH-CONSTRAINTS:** pass. It adds two reads and a few writes per populate.
   - **ARCH-SECURE:** pass. The copy goes through an allowlist with an inert URL, and malformed rows are refused. See the `--local` note above.
   - **ARCH-ORDER:** pass. Nothing is kept between events; the copy runs once per populate.
   - **ARCH-FUNERAL:** pass. Everything written lives inside the checkout, which `Close` and `Sweep` already remove.

7. **Docs and plan:** the atlas has been updated (`atlas/workflow/sdlc-binary.md`). There is no new user-facing surface, so README doesn't need a change. The plan needs no revisions.

```findings
findings:
  - id: new
    severity: Minor
    kind: defect
    family: allowlist-shape-too-loose
    title: |
      A remote branch named feature/main creates a stray remote origin/feature in the review checkout
    detail: |
      authority.go:49 accepts any refs/remotes/<X>/main, so refs/remotes/origin/feature/main makes addRemote("origin/feature"). copyMainAuthority then sets remote.origin/feature.url, which git accepts (verified). It is harmless but adds a remote the slot does not have; accept only remote names without a slash.
  - id: new
    severity: Minor
    kind: defect
    family: authority-source-scope
    title: |
      The main-authority config read covers global and system scope, not just the slot's repository
    detail: |
      checkout.go:140 runs git config --get-regexp without --local, so a user-level branch.main.* or remote.*.fetch is copied as if it were the slot's. A valueless key anywhere in scope makes checkout creation fail and blocks the review; add --local.
```

<!-- sdlc-review-usage [{"version":1,"intent":{"disposable_root":true,"id":"7e377110-ae60-4ef1-8559-65139d46009a","repository":"github.com/xianxu/ariadne","issue_id":"000447","recipe":"milestone-review","round":1,"attempt":1,"agent":"claude","root":"/Users/xianxu/Library/Caches/sdlc/review-checkouts/9207fed21adc0d72","reviewed_head":"126eb88b52d3a8b4b671895272ba6b2214d418e1","slot":"ariadne:2","requested_at":"2026-10-18T21:28:27.97644Z"},"phase":"finished","pid":78431,"started_at":"2026-10-18T21:28:28.183367Z","finished_at":"2026-10-18T21:30:02.907021Z","session_id":"16988d97-20d6-4b89-b7b9-dc28c42796a2","outcome":"returned"}] -->

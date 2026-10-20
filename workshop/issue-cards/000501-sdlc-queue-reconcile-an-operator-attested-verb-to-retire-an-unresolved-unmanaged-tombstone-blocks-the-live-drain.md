---
id: 000501
status: codecomplete
created: 2026-10-20
updated: 2026-10-20
estimate_hours:
github_issue:
started: 2026-10-20T10:56:04-07:00
claimant:
    operator: Xian Xu
    machine: 4716879978a7b90f6b583da1716fd0e9
    machine_name: Xian’s MacBook Pro
    workspace: ariadne:3
    worktree: /Users/xianxu/workspace/worktree/ariadne-slot3/ariadne
    repository: github.com/xianxu/ariadne
actual_hours: 0.48
tracker:
    version: 1
    completion:
        token: close-1bd33dda3891
        repository: github.com/xianxu/ariadne
        reviewed_head: fe03d734f03caabf9bd30cf238d24e9dcc8f8008
        evidence_commit: f98fb7152455c13c832d1055a38e1b4e81e896ee
---

# sdlc queue reconcile: an operator-attested verb to retire an unresolved unmanaged tombstone (blocks the live drain)

## Problem

ariadne#481's queue keeps an **unresolved tombstone** when an unmanaged direct `make test` owner ends without proven
descendant cleanup (by design: no ps-based absence inference). Only a verified OS reboot retires it today, and
`sdlc queue` has no verb for the operator to reconcile one on the same boot. Live on 10-20:
`~/.local/state/ariadne/queue/unresolved/2b6ca56a5f4147569665ee2f073ec211.json` (managed=false, phase=unknown,
original PID 91801 / start 1792501451.112364, now absent; raw sha256 50d51e03…) keeps `sdlc queue status` at
complete=false forever, so pair#722's live drain can never reach "drained" and its acceptance is blocked.

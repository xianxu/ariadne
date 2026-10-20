---
id: 000501
status: open
created: 2026-10-20
updated: 2026-10-20
estimate_hours:
github_issue:
---

# sdlc queue reconcile: an operator-attested verb to retire an unresolved unmanaged tombstone (blocks the live drain)

## Problem

ariadne#481's queue keeps an **unresolved tombstone** when an unmanaged direct `make test` owner ends without proven
descendant cleanup (by design: no ps-based absence inference). Only a verified OS reboot retires it today, and
`sdlc queue` has no verb for the operator to reconcile one on the same boot. Live on 10-20:
`~/.local/state/ariadne/queue/unresolved/2b6ca56a5f4147569665ee2f073ec211.json` (managed=false, phase=unknown,
original PID 91801 / start 1792501451.112364, now absent; raw sha256 50d51e03…) keeps `sdlc queue status` at
complete=false forever, so pair#722's live drain can never reach "drained" and its acceptance is blocked.

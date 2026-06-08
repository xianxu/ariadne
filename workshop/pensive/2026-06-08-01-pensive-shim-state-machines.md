---
type: pensive
date: 2026-06-08
topic: shim construction — the R / M / S state machines
mode: thoughts
description: Living findings doc for ariadne#71. The shim pattern is more than a port + stateful fake — the fake is an executable model of the provider's hidden state machine, and a consumer-POV state machine should be made explicit. Hidden provider state manifests as faults on our side.
references: [workshop/issues/000071-external-service-shims.md]
---

# Pensive: Shim construction — the R / M / S state machines

This is a work-in-progress note. As I build out shims (gh done, Google OAuth
in flight, more coming via nous#45) I want to keep the pattern's findings in one
place, so that when I decide to *improve* the pattern — e.g. "a shim needs not
just a port but an explicit user-side state machine" — I can retrofit every
existing shim against the improved form deliberately, not by memory. ariadne#71
is the promotion gate; this is the scratchpad feeding it. Today's fixed design
decisions in #71 are *port + stateful fake + dual-backend contract +
`New(Conf)`/`NewFake(Conf)`*. The finding below is a candidate extension to that
list: **the fake is an executable state-machine model, and the consumer-POV
state machine should be a first-class, explicit artifact.**

## The core finding: three machines, collapse them to one spec

When you shim an external service there are really three state machines, not
two, and the durable move is to collapse them:

- **R — the real provider machine** (GitHub, Google's OAuth issuer). Huge,
  authoritative, and *hidden*. It has states we will never see: refresh-token
  families with reuse-detection, consent records, tenant policy, admin/password
  revocation, the GitHub new-account visibility lag. Not exposed; only
  *inferable* from how it answers our operations.
- **M — the fake** (`shim'(X)`). Not "stubbed responses" — it's our **executable
  inferred model of R**, restricted to the slice we can observe. M is a
  *hypothesis* about R.
- **S — the consumer-POV state machine.** The states *our own code branches on*
  (OAuth: `Active`/`Expired`/`NeedsReauth`/…; gh: `NotInvited`/`InvitePending`/
  `Collaborator`/…). Smaller than M.

The recommendation I've landed on: **don't model R, and don't keep two machines
in sync. Hold exactly one explicit artifact — S — as the single source of truth,
and treat the fake and the real adapter as two *implementations* of S that the
contract test bisimulates.** R is never modeled directly; it's only ever
consulted as an *oracle* through the dual-backend contract. The fake is
S-made-executable. That turns "two state machines" into **one spec (S), two
implementations (fake / real), one conformance check (the contract)** — which is
exactly where the shim pattern earns its keep.

The formal backbone (kept light): there's an abstraction function α mapping real
provider states → consumer states. We *choose* the fake so its abstracted
dynamics equal S; grounding checks that the *real* provider's abstracted dynamics
also equal S on the paths we exercise. When nous#42 discovered "real GitHub
returns `200 {permission:none}`, not `404`," that was α being wrong — S had
mis-bucketed a distinction, and grounding is what corrected it. So: posit minimal
S → encode as the fake → ground against R → on divergence, refine S → document
the transitions R won't let us exercise. That loop is conformance testing /
hand-building a Mealy machine of the provider, and we deliberately never chase
R's internal states that no consumer branches on. **S is the coarsest machine
whose every distinction some consumer actually acts on** — adding a state must be
justified by a branch in our code.

## The eureka: hidden provider state manifests as *faults* on our side

The hard part of modeling R is the transitions **we don't drive and can't see** —
the provider revokes our grant (password change, admin action, RT-reuse
detection, policy TTL), and from our POV R moved underneath us with no operation
from us. We only *materialize* it lazily on the next call, which now fails.

The insight: **those provider-initiated transitions are exactly what "fault
injection" is.** The fake's fault knobs aren't faults — they are *R's autonomous
transitions, made triggerable by the test*. That reframes the fault API from an
ad-hoc bag of knobs into a principled enumeration: the faults *are* the edges of
S that only R (or the fake standing in for R) can fire. For OAuth that's
`RevokeGrant`, `ExpireGrant`, `DowngradeScope`, `DenyConsent`, `WrongAccount`,
`Transient`. For gh it's the new-account visibility lag and the
`PUT collaborators` no-op-against-existing-invitation peculiarity. Designing the
fake well = enumerating R's *observable autonomous transitions*, not guessing at
error cases. This is what makes the fake a **model** rather than a mock.

A corollary that keeps surfacing: a "current state" in S is always a *belief*,
lazily reconciled. `CheckHealth` is the explicit "go observe now" read (a probe,
no persistence). The honest stance — "Active means active as far as we last
observed; the provider can move between this check and the next call" — is not a
caveat, it's the semantics. S must represent "provider may have moved" as
first-class, which is precisely why the autonomous-transition faults belong *in*
the machine.

## A seam this sharpened: lifecycle (neutral) vs payload/wire (per-provider)

Designing OAuth for a second provider (Microsoft / GitLab / Okta — the YAGNI
objection was wrong here: generalization *is* the deliverable) split the core
cleanly along the state machine:

- **Neutral / shared = the lifecycle = the state machine S.** authorize → active
  → expired → refresh → dead → reauth → revoke. This is provider-independent —
  it's *why* the port generalizes. The refresh-rotation logic I wrote even
  already handles both Google (sometimes rotates the refresh token) and Microsoft
  (always rotates) from the same branch.
- **Per-provider = the payload + wire, which are *not* transitions.** Identity
  claim extraction (`email`+`email_verified` for Google; `preferred_username`/
  `upn` and *no* `email_verified` for Microsoft), endpoint/param dialect
  (`offline_access` scope vs `access_type=offline`), error-code → transition
  mapping, the revoke mechanism (RFC 7009 endpoint vs Graph `revokeSignInSessions`).

The litmus that proved the seam is real: a shared `credentialFromToken` that
rejects `email_verified==false` would reject *every* Microsoft token (missing
claim decodes to false). So the `email_verified` check belongs *below* the
per-provider seam — the state machine doesn't care which claim is the identity,
only that we reached `Active`. The abstraction fell along a real joint.

## Process finding (operator): build against ≥2 real providers + the fake at once

Don't construct a shim against one real provider and the fake, then bolt on a
second provider later. **Work two different real providers plus the fake
simultaneously** (gh: GitHub now, GitLab as the second target; oauth: Google +
Microsoft). Designing against two real backends from the start is what pressures
the port and the state machine into a genuinely durable shape — a single real
backend lets provider-specific quirks masquerade as the abstraction. The fake is
the third implementation that has to satisfy the same S as both. n≥2-real + fake
is the unit of a trustworthy shim.

## Where this points for ariadne#71

Candidate addition to #71's fixed design decisions, to fold in at promotion time
(after the pattern proves out on oauth ×2 providers, then is validated by
retrofitting gh):

> Every `shim(X)` ships an **explicit consumer-POV state machine S** (a `target`
> artifact — an invariant defended from drift): states, transitions, the
> operation/observation that fires each, the provider-autonomous transitions
> (the fault set), and per-transition grounding status. The fake executes S; the
> dual-backend contract is a **transition-coverage table** bisimulating S against
> the real service on the observable quotient. The port stays provider-neutral;
> the wire + payload (identity-claim extraction, endpoint dialect, error mapping,
> revoke mechanism) are the per-provider adapter.

This is the generalization of the pattern from "fake-with-seeded-state" to
"fake-as-contract-checked-model." It does *not* replace the port; it adds the
articulation layer that makes the fake's fidelity legible and the faults
principled.

**Second candidate addition — a conformance-grounding index with a freshness
ledger.** The dual-backend contract grounds the fake against reality, but
grounding is only as trustworthy as it is *fresh*: a stale cert is grounding of
unknown validity. So the pattern should require not just *that* each shim has a
conformance test, but a single legible **index of one grounding layer** — a
per-(shim, provider) table recording the conformance test, the grounding creds,
the boundary (what's grounded vs fake-only/manual), and crucially the
**last-certified date + result**. As the surface grows (every external dependency
gets a shim; each shim ≥2 real providers), this index is how we see at a glance
which fakes are still anchored to reality and which have drifted out of
certification. nous built the first instance:
`atlas/nous/shim-conformance-grounding.md` (gh certified 2026-06-06, Google
2026-06-08). Operator's framing: this is *one layer of grounding* of the whole
system, and belongs in #71's pattern, not just as a nous artifact. Open edge: the
date is hand-maintained today — does it stay a manually-updated ledger, or
eventually a checked artifact (a cert that writes its own timestamp)? Resist
automation until the manual ledger proves it's worth it.

## Open questions

- **How formal should S be?** A prose `target` with a transition table is enough
  to start. A machine-checkable DSL (so the contract is *generated* from S) is
  tempting but smells like premature framework-building — #71 explicitly forbids
  a shared cross-service framework. Resist until n is large.
- **gh retrofit timing.** gh already has an *implicit* S (invite/collaborator
  lifecycle; the nous#25 visibility lag is gh's analogue of oauth's clock-driven
  expiry; `FailListInvitations` is already an R-autonomous-transition fault). So
  the model fits gh — but by #71's own "don't generalize from n=1" rule, retrofit
  gh to *explicit* S only after explicit-S proves out on oauth, and use the gh
  retrofit as the **cross-domain validation** (control-plane CRUD vs credential
  lifecycle — does one S-formalism fit both?). Expectation: the retrofit is a
  test+doc articulation over the fake's existing state, likely *zero* production
  change. Don't do it preemptively.
- **Async sub-machines.** OAuth's `Auth` transition contains a nested
  consent/callback machine (`PendingConsent → CodeReceived | Denied`) with an
  async event. Is "transition containing a sub-machine the fake drives by
  delivering events" a recurring shape worth naming, or an oauth one-off? Watch
  the next few shims.
- **Grounding boundary as part of S.** Some S-transitions can't be exercised
  against R cheaply (OAuth consent leg = non-headless; revoke = destructive to the
  grounding token). These live as per-transition annotations on S, not a separate
  machine. Confirm that scales as the surface grows (nous#45).

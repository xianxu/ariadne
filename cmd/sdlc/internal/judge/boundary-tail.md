## Prior rounds — dispose of these BEFORE raising anything new

{{PRIOR_FINDINGS}}

Do NOT re-raise a finding listed as already disposed — not at the same severity,
and not at a lower one. If a disposed finding is genuinely still wrong, dispose it
`not-addressed` and say what remains, rather than raising it again under a new id.

{{FINDINGS_BLOCK}}

## Answerable boundary response

REWORK requires at least one new blocking finding with a nonempty title and detail,
or an explicit `not-addressed` disposition maintaining a known blocking finding
in this review's scope. Use `id: new` for new findings; the binary assigns BR-n.
An old open finding alone, a Minor-only finding, or an unknown ID cannot justify
REWORK. Every disposition must refer to a known finding in scope.

When the review environment prevents verification, emit `kind: unverifiable`
(the optional kind defaults to ordinary defect semantics). Its detail must name
the attempted command/check, the observed restriction or failure, and the host
evidence or environment change that would answer it. Do not emit empty REWORK.

The implementer can answer an unverifiable finding through `sdlc finding dispute`
with measured host evidence. In the next round, explicitly dispose every pending
unverifiable dispute with a nonempty note answering that evidence: withdraw it
when the evidence holds, or mark it `not-addressed` and explain what remains.
Do not silently carry a dispute forward or replace it with a new ID.

Invalid responses share the existing two-attempt budget with missing verdicts.
After the first invalid response, the checkout is reset, confinement recreated,
and any declared preparation rerun before review. A second invalid response ends
in a protocol error: neither attempt's findings are admitted, the reviewed head
does not advance, and the close does not finalize.

{{BOUNDARY_CONTRACT}}

{{REVIEW_NOTES}}Review window:
{{REVIEW_WINDOW}}

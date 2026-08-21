{{CODE_REVIEW_BODY}}

{{ARCH_BLOCK}}

## Prior rounds — dispose of these BEFORE raising anything new

{{PRIOR_FINDINGS}}

Do NOT re-raise a finding listed as already disposed — not at the same severity,
and not at a lower one. If a disposed finding is genuinely still wrong, dispose it
`not-addressed` and say what remains, rather than raising it again under a new id.

{{FINDINGS_BLOCK}}

{{BOUNDARY_CONTRACT}}

Diff:
{{DIFF}}

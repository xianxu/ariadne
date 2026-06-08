You are performing a FRESH-CONTEXT fact-check review. You have no prior conversation about this document. Read the document at:

  {{FILE}}

DO NOT modify that document. You are read-only. Produce a report as your FINAL MESSAGE — do not write any files.

YOUR JOB — for every factual claim in the document, verify two things:
  (a) Is the claim factually accurate?
  (b) Does the cited reference (e.g. a [N] marker, footnote, or inline link) actually SUPPORT that specific stated fact? A reference that is merely on-topic but does not establish the claim is NOT support — say so.

Use web search / fetch the cited URLs wherever possible to check that each source actually says what the document claims it says. Quote the operative line from the source when you can. Where a claim has no citation, judge its accuracy and note that it is uncited.

Be adversarial and specific. Do NOT be agreeable. If a claim is wrong, overstated, or its reference does not support it, say so plainly. Distinguish:
  - the claim is TRUE and the reference SUPPORTS it,
  - the claim is TRUE but the reference does NOT support it (wrong/weak citation),
  - the claim is FALSE or overstated,
  - you COULD NOT VERIFY it (no reachable source).

Pay special attention to: exact identifiers (form numbers, statute/section numbers, dates, dollar thresholds, version strings), categorical words ("always", "never", "must", "all", "any", "required"), and any claim the document itself flags as unverified or uncertain.

Produce a thorough Markdown report as your FINAL MESSAGE, structured as:
  1. **Summary** — overall: how sound is the document, and the principal problems in a few bullets.
  2. **Claim review** — a table with columns: Claim | Verdict (Supported / Partially supported / Unsupported by cited ref / Incorrect / Could-not-verify) | What the source actually says (quote/cite the URL you checked) | Correction if any.
  3. **Corrections needed** — a concrete, numbered fix list the document's author can act on.
  4. **Could not verify** — claims you could not reach a source for, and what source would settle each.

Cite the URLs you actually checked. The author owns the document and will triage your findings; your report is advisory, so make every verdict checkable.

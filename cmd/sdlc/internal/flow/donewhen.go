package flow

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strings"

	"github.com/xianxu/ariadne/cmd/sdlc/internal/issue"
)

// The contract sections. Spec and Revisions are the contract (the constitution
// records a mid-stream reframe by APPENDING Revisions, so a check that watched
// Spec alone would miss exactly the reframe it exists for — #231 PQ-2); Done
// when is the acceptance oracle the quick flow's single review judges against.
const (
	secSpec      = "Spec"
	secRevisions = "Revisions"
	secDoneWhen  = "Done when"
)

// ContractHashes hashes the contract and the acceptance criteria as they stand.
// change-code records both in the quick flow record on every run, so the anchor
// moves whenever the contract is re-fixed and the check needs no git history.
//
// Section bodies come from the fence-aware extractor, and whitespace is
// collapsed first so re-wrapping prose is not a reframe. Each hash is the 8-hex
// prefix of a sha256 — an integrity hint for the agent's own freshness check,
// not a security boundary.
func ContractHashes(body string) (spec, done string) {
	s, _ := issue.SectionBody(body, secSpec)
	r, _ := issue.SectionBody(body, secRevisions)
	d, _ := issue.SectionBody(body, secDoneWhen)
	return shortHash(squash(s) + "\x00" + squash(r)), shortHash(squash(d))
}

// WithContract returns f carrying the body's current contract hashes.
func WithContract(f Flow, body string) Flow {
	f.spec, f.done = ContractHashes(body)
	return f
}

// DoneWhenPresent requires at least one `## Done when` bullet. On the quick
// flow the close review is the only review and Done-when is its only oracle, so
// a populated `related:` frontmatter — which satisfies the structural gate —
// does not count here.
func DoneWhenPresent(body string) error {
	if issue.HasDoneWhenBullet(body) {
		return nil
	}
	return fmt.Errorf("quick flow: `## Done when` has no bullet — it is the close review's only " +
		"oracle, so write the acceptance criteria before closing")
}

// DoneWhenFresh refuses a contract that moved without its acceptance criteria:
// the recorded Spec+Revisions hash differs from the current one while the Done
// when hash does not. A record without hashes has no anchor and refuses with
// both fixes named.
func DoneWhenFresh(rec Flow, body string) error {
	if rec.spec == "" || rec.done == "" {
		return fmt.Errorf("quick flow: no contract anchor in the flow record — re-run " +
			"`sdlc change-code` to record one, or pass --no-done-when-fresh if Done-when is current")
	}
	spec, done := ContractHashes(body)
	if spec != rec.spec && done == rec.done {
		return fmt.Errorf("quick flow: `## Spec` or `## Revisions` changed since change-code but " +
			"`## Done when` did not — restate the acceptance criteria for what is being built now, " +
			"or pass --no-done-when-fresh with the reason in --verified")
	}
	return nil
}

func squash(s string) string { return strings.Join(strings.Fields(s), " ") }

func shortHash(s string) string {
	sum := sha256.Sum256([]byte(s))
	return hex.EncodeToString(sum[:])[:8]
}

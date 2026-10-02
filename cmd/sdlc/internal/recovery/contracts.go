// Package recovery is the single source of the workflow verbs' operation
// recovery contracts (#280): for each verb, what it does when repeated,
// interrupted or answered by a lost response, how to observe its outcome, and
// which tests prove each guarantee. The catalog is appended to each verb's help
// (attachRecoveryContracts in cmd/sdlc) and rendered in the `sdlc help
// recovery` topic; a contract test ties every proof to a test that exists and
// every workflow verb to an entry.
//
// Scope: repositories with an issue tracker (#252). A legacy repository's verbs
// publish details to main directly; there the contracts are unknown.
//
// The contracts describe mechanisms that already exist (card compare-and-swap,
// receipts, `issue recovery reconcile`, landing observations, #279
// observations). They never promise exactly-once execution of an agent's
// instructions: a claim reserves an issue, it does not make later edits
// idempotent.
package recovery

import (
	"fmt"
	"strings"
)

// Class is what repeating a verb does. Never "idempotent" for everything.
type Class string

const (
	// ReadOnly writes nothing; repeat freely.
	ReadOnly Class = "read-only"
	// DuplicateSafeRefusal: a repeat after success is refused with a message
	// and changes nothing.
	DuplicateSafeRefusal Class = "duplicate-safe-refusal"
	// ConvergentRetry: a repeat reaches the same end state — after success it
	// is a no-op, after a lost response it settles the outcome.
	ConvergentRetry Class = "convergent-retry"
	// NonRepeatable: a repeat performs a new effect (a new review, a new
	// generation); recover through the named recovery path, not by re-running.
	NonRepeatable Class = "non-repeatable"
)

// Classes is the closed set, in teaching order.
var Classes = []Class{ReadOnly, DuplicateSafeRefusal, ConvergentRetry, NonRepeatable}

// ClassMeaning explains each class once.
var ClassMeaning = map[Class]string{
	ReadOnly:             "writes nothing; repeat freely",
	DuplicateSafeRefusal: "a repeat after success is refused and changes nothing",
	ConvergentRetry:      "a repeat reaches the same end state: a no-op after success, the settled outcome after a lost response",
	NonRepeatable:        "a repeat performs a new effect; recover through the named path, never by re-running",
}

// Scope bounds every contract: they describe issue-tracker repositories.
const Scope = "repositories with an issue tracker (#252). In a legacy repository (no issue-tracker branch) these guarantees are unknown: its verbs publish details to main directly."

// Proof is one guarantee and the tests that demonstrate it. A claim with no
// tests is rendered as unproven — an unknown, never a promise.
type Proof struct {
	Claim string
	Tests []string
}

// Contract is one verb's (or one family's) recovery contract.
type Contract struct {
	Verbs         []string // command paths below `sdlc`, e.g. "claim", "issue set-status"
	Class         Class
	Effects       string // what it writes
	Evidence      string // how to observe the outcome (read-only)
	Preconditions string
	Repeat        string // what a repeat does
	LostResponse  string // what to do after a lost response or an interruption
	Ends          string // when its guarantees cease
	Proofs        []Proof
}

// Validate checks the catalog's own shape: a known class, every field
// present, at least one proof, every proof a claim.
func Validate(cs []Contract) error {
	seen := map[string]bool{}
	for _, c := range cs {
		if len(c.Verbs) == 0 {
			return fmt.Errorf("a contract names no verb")
		}
		name := strings.Join(c.Verbs, ", ")
		known := false
		for _, k := range Classes {
			known = known || c.Class == k
		}
		if !known {
			return fmt.Errorf("%s: unknown class %q", name, c.Class)
		}
		for field, v := range map[string]string{"effects": c.Effects, "evidence": c.Evidence, "preconditions": c.Preconditions,
			"repeat": c.Repeat, "lost-response": c.LostResponse, "ends": c.Ends} {
			if strings.TrimSpace(v) == "" {
				return fmt.Errorf("%s: no %s", name, field)
			}
		}
		if len(c.Proofs) == 0 {
			return fmt.Errorf("%s: no proofs (an unproven contract states its claims with no tests)", name)
		}
		for _, p := range c.Proofs {
			if strings.TrimSpace(p.Claim) == "" {
				return fmt.Errorf("%s: a proof without a claim", name)
			}
		}
		for _, v := range c.Verbs {
			if seen[v] {
				return fmt.Errorf("%s has two contracts", v)
			}
			seen[v] = true
		}
	}
	return nil
}

// For returns verb's contract.
func For(verb string) (Contract, bool) {
	for _, c := range Catalog {
		for _, v := range c.Verbs {
			if v == verb {
				return c, true
			}
		}
	}
	return Contract{}, false
}

// Section renders verb's contract for its --help page.
func Section(verb string) string {
	c, ok := For(verb)
	if !ok {
		return "RECOVERY (#280): no contract recorded for `sdlc " + verb + "` — see `sdlc help recovery`."
	}
	var b strings.Builder
	fmt.Fprintf(&b, "RECOVERY (#280) — %s:\n%s", c.Class, wrap(ClassMeaning[c.Class], "  ", "  "))
	b.WriteString(wrap(Scope, "  Scope:          ", strings.Repeat(" ", 18)))
	field := func(label, text string) {
		b.WriteString(wrap(text, fmt.Sprintf("  %-15s ", label), strings.Repeat(" ", 18)))
	}
	field("Effects:", c.Effects)
	field("Evidence:", c.Evidence)
	field("Preconditions:", c.Preconditions)
	field("On repeat:", c.Repeat)
	field("Lost response:", c.LostResponse)
	field("Ends when:", c.Ends)
	b.WriteString("  Proven by:\n")
	for _, p := range c.Proofs {
		proof := strings.Join(p.Tests, ", ")
		if len(p.Tests) == 0 {
			proof = "UNPROVEN (no test; treat as unknown)"
		}
		b.WriteString(wrap(p.Claim+" — "+proof, "    - ", "      "))
	}
	b.WriteString("  Full table, classes and agent guidance: `sdlc help recovery`.")
	return b.String()
}

// helpWidth is the column help text wraps at.
const helpWidth = 79

// wrap fills text into lines no wider than helpWidth (a word longer than the
// line stays whole), the first line prefixed by first, the rest by rest.
func wrap(text, first, rest string) string {
	var b strings.Builder
	line := first
	for _, word := range strings.Fields(text) {
		if line != first && line != rest && len(line)+1+len(word) > helpWidth {
			b.WriteString(line + "\n")
			line = rest
		}
		if line == first || line == rest {
			line += word
		} else {
			line += " " + word
		}
	}
	b.WriteString(line + "\n")
	return b.String()
}

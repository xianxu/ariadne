package recovery

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

// ClassesText lists the classes for the `sdlc help recovery` topic.
func ClassesText() string {
	var b strings.Builder
	for _, c := range Classes {
		b.WriteString(wrap(ClassMeaning[c], fmt.Sprintf("  %-24s ", c), strings.Repeat(" ", 27)))
	}
	return strings.TrimRight(b.String(), "\n")
}

// ScopeText is the wrapped scope line for the topic.
func ScopeText() string { return strings.TrimRight(wrap(Scope, "Scope: ", "  "), "\n") }

// Table is one entry per contract: its verbs and class, then its
// lost-response action.
func Table() string {
	var b strings.Builder
	for _, c := range Catalog {
		b.WriteString(wrap(fmt.Sprintf("%s [%s]", strings.Join(c.Verbs, ", "), c.Class), "  ", "  "))
		b.WriteString(wrap("after a lost response: "+c.LostResponse, "      ", "      "))
	}
	b.WriteString("  (each verb's --help carries its full contract and proofs)")
	return b.String()
}

// Actor is who runs an example step.
type Actor string

const (
	Coordinator Actor = "coordinator" // asked for the work; verifies effects
	Recipient   Actor = "recipient"   // the slot that does the work
)

// Expect is one fact a coordinator reads from the observation: a dot path into
// `sdlc issue show N --json` and its expected value.
type Expect struct{ Path, Equals string }

// Step is one step of the scheduling example. Rendered into the help topic
// and executed, step by step, by TestSchedulingExampleRuns — so the documented
// example and the tested one are the same data.
type Step struct {
	Actor     Actor
	Does      string   // what the step is, in words
	Command   string   // the command run ("" for a step that is not an sdlc command)
	Expect    []Expect // for an observation step
	Otherwise string   // what the coordinator does when the evidence differs
	// TrackerUnreachable: the step happens with the tracker remote unreachable.
	TrackerUnreachable bool
	// IfVerdict: the step happens only after a close with this verdict.
	IfVerdict string
}

// Example is the scheduling example: delegate an issue across slots and verify
// each effect through read-only observation, never through the messaging
// channel's acknowledgement.
var Example = []Step{
	{Actor: Coordinator, Does: "asks a free slot to take the issue (a Couch message; its receipt proves delivery to the terminal, nothing more)", Command: ""},
	{Actor: Recipient, Does: "claims it", Command: "sdlc claim --issue N"},
	{Actor: Coordinator, Does: "verifies the claim landed, and whose it is", Command: "sdlc issue show N --json",
		Expect:    []Expect{{"card.status", "open"}, {"assignment.relation", "other-workspace"}},
		Otherwise: "still open: the request may be lost or unread — resending is safe (claim is a convergent retry; one claim wins). Look again after about 30 s before concluding anything."},
	{Actor: Recipient, Does: "starts planning on the issue branch", Command: "sdlc start-plan --issue N"},
	{Actor: Recipient, Does: "enters implementation", Command: "sdlc change-code --issue N"},
	{Actor: Coordinator, Does: "sees the branch, the slot's activity and the recorded flow", Command: "sdlc issue show N --json",
		Expect:    []Expect{{"branch.state", "present"}, {"workspaces.state", "present"}, {"checkpoints.state", "present"}},
		Otherwise: "no branch yet: the recipient has not reached start-plan. Activity is not progress — wait for checkpoints (checkpoints.flow says quick or full once change-code records it)."},
	{Actor: Recipient, Does: "implements and commits the work", Command: ""},
	{Actor: Recipient, Does: "closes it (the boundary review runs)", Command: "sdlc close --issue N --verified '<evidence>'"},
	{Actor: Recipient, Does: "commits the review's fixes", IfVerdict: "FIX-THEN-SHIP"},
	{Actor: Recipient, Does: "lands the close's evidence after the fixes", Command: "sdlc issue recovery reconcile --issue N", IfVerdict: "FIX-THEN-SHIP"},
	{Actor: Coordinator, Does: "sees the close completed: its evidence, and no open blocking finding", Command: "sdlc issue show N --json",
		Expect:    []Expect{{"completion.state", "present"}, {"checkpoints.reviews[close].open_blocking", "0"}},
		Otherwise: "a close in progress or interrupted: never ask for it again (close is non-repeatable) — the recipient runs `sdlc issue recovery reconcile --issue N`."},
	{Actor: Coordinator, Does: "with the tracker unreachable, reads a stale answer", Command: "sdlc issue show N --json", TrackerUnreachable: true,
		Expect:    []Expect{{"tracker.state", "stale"}},
		Otherwise: "stale or unknown is not negative evidence: do not conclude the claim or close was lost; look again when the tracker is reachable."},
}

// ExampleText renders the example for the help topic.
func ExampleText() string {
	var b strings.Builder
	for i, s := range Example {
		does := s.Does
		if s.IfVerdict != "" {
			does = "(after a " + s.IfVerdict + " verdict only) " + does
		}
		b.WriteString(wrap(does, fmt.Sprintf("  %2d. %-12s ", i+1, s.Actor+":"), strings.Repeat(" ", 19)))
		if s.Command != "" {
			fmt.Fprintf(&b, "%s$ %s\n", strings.Repeat(" ", 19), s.Command)
		}
		for _, e := range s.Expect {
			fmt.Fprintf(&b, "%sexpect %s = %s\n", strings.Repeat(" ", 19), e.Path, e.Equals)
		}
		if s.Otherwise != "" {
			b.WriteString(wrap(s.Otherwise, strings.Repeat(" ", 19)+"otherwise: ", strings.Repeat(" ", 19)))
		}
	}
	return strings.TrimRight(b.String(), "\n")
}

var indexRE = regexp.MustCompile(`^(\w+)\[(\w+)\]$`)

// Lookup reads a dot path from a decoded observation. `list[key]` selects the
// element of list whose "boundary" equals key (the reviews list). The second
// result is false when the path does not resolve.
func Lookup(doc any, path string) (string, bool) {
	cur := doc
	for _, part := range strings.Split(path, ".") {
		field, key := part, ""
		if m := indexRE.FindStringSubmatch(part); m != nil {
			field, key = m[1], m[2]
		}
		obj, ok := cur.(map[string]any)
		if !ok {
			return "", false
		}
		if cur, ok = obj[field]; !ok {
			return "", false
		}
		if key != "" {
			list, ok := cur.([]any)
			if !ok {
				return "", false
			}
			found := false
			for _, el := range list {
				if m, ok := el.(map[string]any); ok && m["boundary"] == key {
					cur, found = el, true
					break
				}
			}
			if !found {
				return "", false
			}
		}
	}
	switch v := cur.(type) {
	case string:
		return v, true
	case float64:
		return strconv.FormatFloat(v, 'f', -1, 64), true
	case bool:
		return strconv.FormatBool(v), true
	}
	return "", false
}

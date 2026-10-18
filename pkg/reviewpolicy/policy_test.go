package reviewpolicy

import (
	"errors"
	"strings"
	"testing"

	"github.com/xianxu/ariadne/pkg/agentenv"
)

func TestFamilyOf(t *testing.T) {
	for agent, want := range map[string]Family{
		"claude": Anthropic, "codex": OpenAI, "gemini": Google,
		"human": "", "qoder": "", "": "",
	} {
		if got := FamilyOf(agent); got != want {
			t.Errorf("FamilyOf(%q) = %q, want %q", agent, got, want)
		}
	}
}
func agentsOf(plan []Candidate) string {
	var s []string
	for _, c := range plan {
		s = append(s, string(c.Agent))
	}
	return strings.Join(s, ",")
}
func TestPlan(t *testing.T) {
	other := Policy{Mode: Other}
	all := func(Agent) (string, bool) { return "", true }
	noCodex := func(a Agent) (string, bool) {
		if a == Codex {
			return "codex not on PATH", false
		}
		return "", true
	}
	none := func(a Agent) (string, bool) { return string(a) + " missing", false }
	anth, oai, goo := Anthropic, OpenAI, Google
	for _, tc := range []struct {
		name    string
		policy  Policy
		main    Agent
		authors []Family
		kind    ChangeKind
		avail   Availability
		want    string // candidate agents in order
		skips   string // skipped agents in order
	}{
		// #310's floor: same and explicit name one reviewer whatever the authors.
		{"same reviews with the main agent", Policy{Mode: Same}, Claude, []Family{anth}, Bounded, all, "claude", ""},
		{"the default mode is the constant", Policy{}, Codex, []Family{oai}, DesignHeavy, all, string(map[Mode]Agent{Same: Codex, Other: Claude}[DefaultMode]), ""},
		{"explicit keeps its pair", Policy{Mode: Explicit, Pairs: map[Agent]Agent{Claude: Gemini}}, Claude, []Family{anth}, Bounded, all, "gemini", ""},
		// The dynamic `other`: every other family, in the kind's order.
		{"claude-authored bounded change: codex first", other, Claude, []Family{anth}, Bounded, all, "codex,gemini", ""},
		{"claude-authored design: codex first (claude excluded)", other, Claude, []Family{anth}, DesignHeavy, all, "codex,gemini", ""},
		{"codex-authored design: claude first", other, Codex, []Family{oai}, DesignHeavy, all, "claude,gemini", ""},
		{"codex-authored bounded change: claude, then gemini", other, Codex, []Family{oai}, Bounded, all, "claude,gemini", ""},
		{"mixed authors leave only the third family", other, Claude, []Family{anth, oai}, Bounded, all, "gemini", ""},
		{"all three families authored: nobody left", other, Claude, []Family{anth, oai, goo}, Bounded, all, "", ""},
		{"unknown authors: every agent in order", other, Claude, nil, Bounded, all, "codex,claude,gemini", ""},
		// Availability is a pre-launch skip, recorded.
		{"an unavailable codex is skipped, not silently same-family", other, Claude, []Family{anth}, Bounded, noCodex, "gemini", "codex"},
		{"nobody available: empty plan, every skip recorded", other, Claude, []Family{anth}, Bounded, none, "", "codex,gemini"},
		{"a repo order overrides the default", Policy{Mode: Other, Orders: map[ChangeKind][]Agent{Bounded: {Gemini, Codex}}}, Claude, []Family{anth}, Bounded, all, "gemini,codex", ""},
	} {
		plan, skips := Plan(tc.policy, tc.main, tc.authors, tc.kind, tc.avail)
		var skipped []string
		for _, s := range skips {
			skipped = append(skipped, string(s.Agent))
			if s.Reason == "" {
				t.Errorf("%s: skip of %s has no reason", tc.name, s.Agent)
			}
		}
		if got := agentsOf(plan); got != tc.want {
			t.Errorf("%s: plan %q, want %q", tc.name, got, tc.want)
		}
		if got := strings.Join(skipped, ","); got != tc.skips {
			t.Errorf("%s: skips %q, want %q", tc.name, got, tc.skips)
		}
		for _, c := range plan {
			for _, f := range tc.authors {
				if tc.policy.EffectiveMode() == Other && c.Family == f {
					t.Errorf("%s: candidate %s shares the author family %s", tc.name, c.Agent, f)
				}
			}
		}
	}
}
func TestReviewPlanCarriesModels(t *testing.T) {
	p := Policy{Mode: Other, Models: map[Agent]string{Codex: "gpt-6-astra"}}
	plan, _ := Plan(p, Claude, []Family{Anthropic}, Bounded, nil)
	if len(plan) < 2 || plan[0].Model != "gpt-6-astra" || plan[1].Model != "" {
		t.Fatalf("models not carried: %+v", plan)
	}
	if got := plan[0].String(); got != "codex/gpt-6-astra (openai)" {
		t.Errorf("candidate string %q", got)
	}
}
func TestValidateOrdersAndModels(t *testing.T) {
	for _, tc := range []struct {
		name   string
		policy Policy
		want   string // "" valid
	}{
		{"orders and models", Policy{Mode: Other, Orders: map[ChangeKind][]Agent{Bounded: {Gemini, Codex}}, Models: map[Agent]string{Codex: "gpt-6-astra"}}, ""},
		{"unknown kind", Policy{Orders: map[ChangeKind][]Agent{"huge": {Codex}}}, "unknown change kind"},
		{"unknown agent in an order", Policy{Orders: map[ChangeKind][]Agent{Bounded: {"grok"}}}, "unknown agent"},
		{"unknown agent in models", Policy{Models: map[Agent]string{"grok": "x"}}, "unknown agent"},
		{"a model with spaces", Policy{Models: map[Agent]string{Codex: "gpt 6"}}, "no usable model"},
		{"a flag-like model", Policy{Models: map[Agent]string{Codex: "--yolo"}}, "no usable model"},
		{"an empty order", Policy{Orders: map[ChangeKind][]Agent{Bounded: {}}}, "two families"},
		{"a single-family order", Policy{Orders: map[ChangeKind][]Agent{Bounded: {Claude}}}, "two families"},
		{"a duplicate in an order", Policy{Orders: map[ChangeKind][]Agent{Bounded: {Codex, Claude, Codex}}}, "twice"},
	} {
		err := tc.policy.Validate()
		if (err == nil) != (tc.want == "") || err != nil && !strings.Contains(err.Error(), tc.want) {
			t.Errorf("%s: %v, want %q", tc.name, err, tc.want)
		}
	}
}

// Under `other`, Reviewer never hands the main agent its own family.
func TestOtherNeverSameFamily(t *testing.T) {
	for _, main := range []Agent{Claude, Codex, Gemini} {
		p := Policy{Mode: Other, Orders: map[ChangeKind][]Agent{Bounded: {main, Claude, Codex}}}
		if got := p.Reviewer(main); FamilyOf(string(got)) == FamilyOf(string(main)) {
			t.Errorf("%s reviewed by %s, its own family", main, got)
		}
	}
}

// ReviewerEntry is one parsed Review-Reviewer trailer.
type ReviewerEntry struct {
	Agent, Model string
	Family       Family
	Authors      []Family
	Degraded     bool
}

func firstOf(c Choice) Candidate {
	f, _ := c.First()
	return f
}
func TestResolve(t *testing.T) {
	other := Policy{Mode: Other, Models: map[Agent]string{Claude: "opus"}}
	claudeEnv := agentenv.Env{ClaudeCode: "1"}
	c := Resolve("", false, claudeEnv, other, []Family{Anthropic}, Bounded, nil)
	if c.Forced || firstOf(c).Agent != Codex {
		t.Errorf("planned choice: %+v", c)
	}
	c = Resolve("claude", true, claudeEnv, other, []Family{Anthropic}, Bounded, nil)
	if !c.Forced || len(c.Candidates) != 1 || firstOf(c).Agent != Claude || firstOf(c).Model != "opus" {
		t.Errorf("--agent choice: %+v", c)
	}
	c = Resolve("", false, agentenv.Env{AgentCmd: "gemini"}, other, nil, Bounded, nil)
	if !c.Forced || firstOf(c).Agent != Gemini {
		t.Errorf("AGENT_CMD choice: %+v", c)
	}
	// An empty choice reports no first candidate, and its refusal names the skips.
	none := Resolve("", false, claudeEnv, other, []Family{Anthropic}, Bounded, func(a Agent) (string, bool) { return string(a) + " missing", false })
	if _, ok := none.First(); ok {
		t.Error("an empty choice offered a first candidate")
	}
	if err := error(&NoReviewerError{Skips: none.Skips}); !errors.Is(err, ErrNoReviewer) || !strings.Contains(err.Error(), "codex: codex missing") {
		t.Errorf("refusal %v", err)
	}
}
func TestRecordRoundTrip(t *testing.T) {
	codex := Candidate{Agent: Codex, Model: "gpt-6-astra", Family: OpenAI}
	gemini := Candidate{Agent: Gemini, Family: Google}
	claude := Candidate{Agent: Claude, Family: Anthropic}
	for _, tc := range []struct {
		name   string
		used   Candidate
		choice Choice
		skips  []Skip
		want   string
	}{
		{"plain", codex, Choice{Authors: []Family{Anthropic}}, nil, "codex/gpt-6-astra (openai; author anthropic)"},
		{"fallback", gemini, Choice{Authors: []Family{OpenAI, Anthropic}}, []Skip{{Agent: Codex, Reason: "usage limit\nmore"}}, "gemini (google; author anthropic,openai); degraded: codex unavailable: usage limit"},
		{"forced same family", claude, Choice{Authors: []Family{Anthropic}, Forced: true}, nil, "claude (anthropic; author anthropic); degraded: same-family by choice"},
		{"unknown authors", codex, Choice{}, nil, "codex/gpt-6-astra (openai; author unknown)"},
	} {
		got := Record(tc.used, tc.choice, tc.skips)
		if got != tc.want {
			t.Errorf("%s: %q, want %q", tc.name, got, tc.want)
			continue
		}
		e, ok := ParseRecord(got)
		if !ok || e.Agent != string(tc.used.Agent) || e.Family != tc.used.Family || e.Degraded != strings.Contains(tc.want, "degraded") {
			t.Errorf("%s: parsed %+v %v", tc.name, e, ok)
		}
	}
	tally := TallyReviewers([]string{
		"codex/gpt-6-astra (openai; author anthropic)",
		"codex (openai; author anthropic)",
		"claude (anthropic; author openai)",
		"gemini (google; author anthropic); degraded: codex unavailable: quota",
		"not a record",
	})
	want := map[string]int{"anthropic → openai": 2, "openai → anthropic": 1, "anthropic → google (degraded)": 1, "unparsed": 1}
	for k, n := range want {
		if tally[k] != n {
			t.Errorf("tally[%q] = %d, want %d (all %v)", k, tally[k], n, tally)
		}
	}
}

// ParseRecord reads a Review-Reviewer value back; ok is false for a
// value not in Record's form.
func ParseRecord(v string) (ReviewerEntry, bool) {
	head, _, degraded := strings.Cut(v, "; degraded:")
	open := strings.Index(head, " (")
	if open < 0 || !strings.HasSuffix(head, ")") {
		return ReviewerEntry{}, false
	}
	who, inner := head[:open], head[open+2:len(head)-1]
	fam, authors, ok := strings.Cut(inner, "; author ")
	if !ok {
		return ReviewerEntry{}, false
	}
	e := ReviewerEntry{Family: Family(fam), Degraded: degraded}
	e.Agent, e.Model, _ = strings.Cut(who, "/")
	if authors != "unknown" {
		for _, a := range strings.Split(authors, ",") {
			e.Authors = append(e.Authors, Family(a))
		}
	}
	return e, e.Agent != ""
}

// TallyReviewers counts Review-Reviewer values by "author → reviewer" family
// pair, for a week's retro; values not in the form are counted as "unparsed".
func TallyReviewers(values []string) map[string]int {
	out := map[string]int{}
	for _, v := range values {
		e, ok := ParseRecord(strings.TrimSpace(v))
		if !ok {
			out["unparsed"]++
			continue
		}
		key := orUnknown(familyList(e.Authors), "unknown") + " → " + orUnknown(string(e.Family), "unknown")
		if e.Degraded {
			key += " (degraded)"
		}
		out[key]++
	}
	return out
}

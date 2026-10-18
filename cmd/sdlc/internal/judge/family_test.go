package judge

import (
	"errors"
	"strings"
	"testing"
)

func TestFamilyOf(t *testing.T) {
	for agent, want := range map[string]Family{
		"claude": FamilyAnthropic, "codex": FamilyOpenAI, "gemini": FamilyGoogle,
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

func TestReviewPlan(t *testing.T) {
	other := ReviewerPolicy{Mode: ReviewerOther}
	all := func(AgentCLI) (string, bool) { return "", true }
	noCodex := func(a AgentCLI) (string, bool) {
		if a == AgentCodex {
			return "codex not on PATH", false
		}
		return "", true
	}
	none := func(a AgentCLI) (string, bool) { return string(a) + " missing", false }
	anth, oai, goo := FamilyAnthropic, FamilyOpenAI, FamilyGoogle
	for _, tc := range []struct {
		name    string
		policy  ReviewerPolicy
		main    AgentCLI
		authors []Family
		kind    ChangeKind
		avail   Availability
		want    string // candidate agents in order
		skips   string // skipped agents in order
	}{
		// #310's floor: same and explicit name one reviewer whatever the authors.
		{"same reviews with the main agent", ReviewerPolicy{Mode: ReviewerSame}, AgentClaude, []Family{anth}, KindBounded, all, "claude", ""},
		{"the default mode is the constant", ReviewerPolicy{}, AgentCodex, []Family{oai}, KindDesignHeavy, all, string(map[ReviewerMode]AgentCLI{ReviewerSame: AgentCodex, ReviewerOther: AgentClaude}[DefaultReviewerMode]), ""},
		{"explicit keeps its pair", ReviewerPolicy{Mode: ReviewerExplicit, Pairs: map[AgentCLI]AgentCLI{AgentClaude: AgentGemini}}, AgentClaude, []Family{anth}, KindBounded, all, "gemini", ""},
		// The dynamic `other`: every other family, in the kind's order.
		{"claude-authored bounded change: codex first", other, AgentClaude, []Family{anth}, KindBounded, all, "codex,gemini", ""},
		{"claude-authored design: codex first (claude excluded)", other, AgentClaude, []Family{anth}, KindDesignHeavy, all, "codex,gemini", ""},
		{"codex-authored design: claude first", other, AgentCodex, []Family{oai}, KindDesignHeavy, all, "claude,gemini", ""},
		{"codex-authored bounded change: claude, then gemini", other, AgentCodex, []Family{oai}, KindBounded, all, "claude,gemini", ""},
		{"mixed authors leave only the third family", other, AgentClaude, []Family{anth, oai}, KindBounded, all, "gemini", ""},
		{"all three families authored: nobody left", other, AgentClaude, []Family{anth, oai, goo}, KindBounded, all, "", ""},
		{"unknown authors: every agent in order", other, AgentClaude, nil, KindBounded, all, "codex,claude,gemini", ""},
		// Availability is a pre-launch skip, recorded.
		{"an unavailable codex is skipped, not silently same-family", other, AgentClaude, []Family{anth}, KindBounded, noCodex, "gemini", "codex"},
		{"nobody available: empty plan, every skip recorded", other, AgentClaude, []Family{anth}, KindBounded, none, "", "codex,gemini"},
		{"a repo order overrides the default", ReviewerPolicy{Mode: ReviewerOther, Orders: map[ChangeKind][]AgentCLI{KindBounded: {AgentGemini, AgentCodex}}}, AgentClaude, []Family{anth}, KindBounded, all, "gemini,codex", ""},
	} {
		plan, skips := ReviewPlan(tc.policy, tc.main, tc.authors, tc.kind, tc.avail)
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
				if tc.policy.mode() == ReviewerOther && c.Family == f {
					t.Errorf("%s: candidate %s shares the author family %s", tc.name, c.Agent, f)
				}
			}
		}
	}
}

func TestReviewPlanCarriesModels(t *testing.T) {
	p := ReviewerPolicy{Mode: ReviewerOther, Models: map[AgentCLI]string{AgentCodex: "gpt-6-astra"}}
	plan, _ := ReviewPlan(p, AgentClaude, []Family{FamilyAnthropic}, KindBounded, nil)
	if len(plan) < 2 || plan[0].Model != "gpt-6-astra" || plan[1].Model != "" {
		t.Fatalf("models not carried: %+v", plan)
	}
	if got := plan[0].String(); got != "codex/gpt-6-astra (openai)" {
		t.Errorf("candidate string %q", got)
	}
}

func TestUnavailable(t *testing.T) {
	quota := `{"type":"result","subtype":"error","is_error":true,"result":"API Error: 429 rate limit exceeded"}`
	began := `{"type":"assistant","message":{"content":[{"type":"text","text":"Reading the diff."}]}}` + "\n" + quota
	authErr := `{"type":"result","subtype":"error","is_error":true,"result":"Invalid API key · Please run /login"}`
	other := `{"type":"result","subtype":"error","is_error":true,"result":"the model refused the prompt"}`
	// codex and gemini, measured 2026-10-18: codex echoes the prompt to stderr
	// before its ERROR: lines; gemini on an ineligible account.
	codexHeader := "OpenAI Codex\n--------\nmodel: gpt\n--------\nuser\nreview this; watch the quota and rate limit handling\n"
	codexQuota := codexHeader + `ERROR: {"type":"error","status":429,"error":{"type":"usage_limit_reached","message":"You've hit your usage limit."}}` + "\n"
	codexBadModel := codexHeader + `ERROR: {"type":"error","status":400,"error":{"type":"invalid_request_error","message":"The 'x' model is not supported"}}` + "\n"
	codexBegan := codexHeader + "codex\nI'll read the diff.\nexec\ngit diff\n" + `ERROR: {"type":"error","status":429}` + "\n"
	geminiIneligible := "Error: IneligibleTierError: This client is no longer supported for Gemini Code Assist for individuals.\n    at throwIneligibleOrProjectIdError (file:///x.js:1:1)\n"
	for _, tc := range []struct {
		name      string
		agent     AgentCLI
		out       string
		stderr    string
		nonZero   bool
		launchErr error
		want      bool
	}{
		{"launch failure", AgentCodex, "", "", false, errors.New(`exec: "codex": executable file not found in $PATH`), true},
		{"claude: quota before any event", AgentClaude, quota, "", true, nil, true},
		{"claude: sign-in failure before any event", AgentClaude, authErr, "", true, nil, true},
		{"claude: quota after the review began never switches", AgentClaude, began, "", true, nil, false},
		{"claude: an unrelated failure is not unavailability", AgentClaude, other, "", true, nil, false},
		{"claude: stderr carries the cause", AgentClaude, `{"type":"system","subtype":"init"}`, "Error: insufficient_quota", true, nil, true},
		{"codex: usage limit before any output", AgentCodex, "", codexQuota, true, nil, true},
		{"codex: a bad model is a misconfiguration, not unavailability", AgentCodex, "", codexBadModel, true, nil, false},
		{"codex: the echoed prompt's words are not a cause", AgentCodex, "", codexHeader, true, nil, false},
		{"codex: printed a review, so it began", AgentCodex, "partial review text", codexQuota, true, nil, false},
		{"codex: stderr shows it began reviewing, so it never switches", AgentCodex, "", codexBegan, true, nil, false},
		{"gemini: an ineligible account", AgentGemini, "", geminiIneligible, true, nil, true},
		{"a zero exit is not unavailability", AgentGemini, "", geminiIneligible, false, nil, false},
		{"a run with a verdict is never unavailable", AgentCodex, "VERDICT: SHIP", codexQuota, true, nil, false},
	} {
		run := ReadStream([]byte(tc.out))
		reason, got := Unavailable(tc.agent, run, []byte(tc.stderr), tc.nonZero, tc.launchErr)
		if got != tc.want {
			t.Errorf("%s: unavailable=%v (%q), want %v", tc.name, got, reason, tc.want)
		}
		if got && reason == "" {
			t.Errorf("%s: unavailable without a reason", tc.name)
		}
	}
}

func TestValidateOrdersAndModels(t *testing.T) {
	for _, tc := range []struct {
		name   string
		policy ReviewerPolicy
		want   string // "" valid
	}{
		{"orders and models", ReviewerPolicy{Mode: ReviewerOther, Orders: map[ChangeKind][]AgentCLI{KindBounded: {AgentGemini, AgentCodex}}, Models: map[AgentCLI]string{AgentCodex: "gpt-6-astra"}}, ""},
		{"unknown kind", ReviewerPolicy{Orders: map[ChangeKind][]AgentCLI{"huge": {AgentCodex}}}, "unknown change kind"},
		{"unknown agent in an order", ReviewerPolicy{Orders: map[ChangeKind][]AgentCLI{KindBounded: {"grok"}}}, "unknown agent"},
		{"unknown agent in models", ReviewerPolicy{Models: map[AgentCLI]string{"grok": "x"}}, "unknown agent"},
		{"a model with spaces", ReviewerPolicy{Models: map[AgentCLI]string{AgentCodex: "gpt 6"}}, "no usable model"},
		{"a flag-like model", ReviewerPolicy{Models: map[AgentCLI]string{AgentCodex: "--yolo"}}, "no usable model"},
		{"an empty order", ReviewerPolicy{Orders: map[ChangeKind][]AgentCLI{KindBounded: {}}}, "two families"},
		{"a single-family order", ReviewerPolicy{Orders: map[ChangeKind][]AgentCLI{KindBounded: {AgentClaude}}}, "two families"},
		{"a duplicate in an order", ReviewerPolicy{Orders: map[ChangeKind][]AgentCLI{KindBounded: {AgentCodex, AgentClaude, AgentCodex}}}, "twice"},
	} {
		err := tc.policy.Validate()
		if (err == nil) != (tc.want == "") || err != nil && !strings.Contains(err.Error(), tc.want) {
			t.Errorf("%s: %v, want %q", tc.name, err, tc.want)
		}
	}
}

func TestLocalAvailability(t *testing.T) {
	onPath := func(name string) (string, error) {
		if name == "gemini" {
			return "", errors.New("not found")
		}
		return "/bin/" + name, nil
	}
	for _, tc := range []struct {
		agent       AgentCLI
		containment Containment
		goos        string
		want        bool
	}{
		{AgentClaude, ContainCheckout, "darwin", true},
		{AgentCodex, ContainCheckout, "darwin", false},
		{AgentCodex, ContainCheckout, "linux", true},
		{AgentCodex, ContainFenced, "darwin", true},
		{AgentGemini, ContainNone, "linux", false},
	} {
		reason, ok := LocalAvailability(tc.containment, tc.goos, onPath)(tc.agent)
		if ok != tc.want || !ok && reason == "" {
			t.Errorf("%s under %v on %s: ok=%v %q, want %v", tc.agent, tc.containment, tc.goos, ok, reason, tc.want)
		}
	}
}

// Under `other`, Reviewer never hands the main agent its own family.
func TestReviewerOtherNeverSameFamily(t *testing.T) {
	for _, main := range []AgentCLI{AgentClaude, AgentCodex, AgentGemini} {
		p := ReviewerPolicy{Mode: ReviewerOther, Orders: map[ChangeKind][]AgentCLI{KindBounded: {main, AgentClaude, AgentCodex}}}
		if got := p.Reviewer(main); FamilyOf(string(got)) == FamilyOf(string(main)) {
			t.Errorf("%s reviewed by %s, its own family", main, got)
		}
	}
}

// The reviewer's model reaches each CLI's own flag.
func TestAgentArgsPassTheModel(t *testing.T) {
	for agent, flag := range map[AgentCLI]string{AgentClaude: "--model", AgentCodex: "-m", AgentGemini: "-m"} {
		_, args, err := agentArgs(DispatchOptions{Agent: agent, Model: "m-1", Prompt: "p"})
		if err != nil {
			t.Fatal(err)
		}
		joined := strings.Join(args, " ")
		if !strings.Contains(joined, flag+" m-1") || args[len(args)-1] != "p" {
			t.Errorf("%s args %v lack %s m-1 before the prompt", agent, args, flag)
		}
		_, plain, _ := agentArgs(DispatchOptions{Agent: agent, Prompt: "p"})
		if strings.Contains(strings.Join(plain, " "), flag+" ") {
			t.Errorf("%s passes a model flag with no model: %v", agent, plain)
		}
	}
}

package judge

import (
	"errors"
	"strings"
	"testing"
)





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

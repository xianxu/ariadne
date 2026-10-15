package agentid

import "testing"

func TestIdentityRoundTrip(t *testing.T) {
	for _, tc := range []struct {
		id   Identity
		want string
	}{
		{Identity{"claude", "claude-opus-5-5", "medium"}, "claude/claude-opus-5-5/medium"},
		{Identity{"codex", "gpt-6-astra", ""}, "codex/gpt-6-astra/-"},
		{Identity{"claude", "", ""}, "claude/-/-"},
		{Identity{}, Human},
	} {
		if got := tc.id.String(); got != tc.want {
			t.Errorf("%+v.String() = %q, want %q", tc.id, got, tc.want)
		}
		back, ok := Parse(tc.want)
		if !ok || back != tc.id {
			t.Errorf("Parse(%q) = %+v, %v; want %+v", tc.want, back, ok, tc.id)
		}
	}
	// #387: history from #367 says "sdlc" for the same thing.
	if id, ok := Parse(Legacy); !ok || id != (Identity{}) {
		t.Errorf("Parse(%q) = %+v, %v; want the human identity", Legacy, id, ok)
	}
	for _, bad := range []string{"", "claude", "claude/x", "-/m/e", "a/b/c/d"} {
		if _, ok := Parse(bad); ok {
			t.Errorf("Parse(%q) accepted a malformed value", bad)
		}
	}
}

func TestStamp(t *testing.T) {
	id := Identity{"claude", "claude-opus-5-5", "high"}
	const tr = "Agent: claude/claude-opus-5-5/high"
	for _, tc := range []struct{ name, in, want string }{
		{"subject only", "#9: close", "#9: close\n\n" + tr + "\n"},
		{"subject with newline", "#9: close\n", "#9: close\n\n" + tr + "\n"},
		{"body, no trailers", "#9: x\n\nwhy it changed", "#9: x\n\nwhy it changed\n\n" + tr + "\n"},
		{"joins a trailer block", "#9: close\n\nReview-Verdict: SHIP\nReview-Window: a..b\n",
			"#9: close\n\nReview-Verdict: SHIP\nReview-Window: a..b\n" + tr + "\n"},
		{"already stamped", "#9: x\n\nAgent: codex/gpt-6-astra/medium\n", "#9: x\n\nAgent: codex/gpt-6-astra/medium\n"},
	} {
		if got := Stamp(tc.in, id); got != tc.want {
			t.Errorf("%s: Stamp = %q, want %q", tc.name, got, tc.want)
		}
	}
}

func TestMissingAgent(t *testing.T) {
	got := MissingAgent([]Commit{
		{SHA: "a", Agents: []string{"claude/m/e"}},
		{SHA: "b"},
		{SHA: "c", Merge: true},
	})
	if len(got) != 1 || got[0].SHA != "b" {
		t.Fatalf("MissingAgent = %+v, want only b", got)
	}
}

func TestParseLog(t *testing.T) {
	out := "aaa\x1fp1\x1f#9: x\x1fclaude/m/e\x1d\n" +
		"bbb\x1fp1\x1f#9: bare\x1f\x1d\n" +
		"ccc\x1fp1 p2\x1fMerge main\x1f\x1d\n" +
		"ddd\x1fp1\x1ftwo\x1fa/b/c\x1esdlc\x1d\n"
	got := ParseLog(out)
	if len(got) != 4 {
		t.Fatalf("ParseLog = %+v", got)
	}
	if got[0].Agents[0] != "claude/m/e" || len(got[1].Agents) != 0 || !got[2].Merge || len(got[3].Agents) != 2 {
		t.Fatalf("ParseLog = %+v", got)
	}
	if miss := MissingAgent(got); len(miss) != 1 || miss[0].SHA != "bbb" {
		t.Fatalf("MissingAgent = %+v", miss)
	}
}

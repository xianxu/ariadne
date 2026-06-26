package transcripts

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

// codexCWDFromBytes is the pure parse: the robustness matrix (malformed first
// line, no session_meta, empty input) is a parse-level concern, so it's asserted
// directly on bytes — no temp files (ARCH-PURE, plan-quality finding #2).
func TestCodexCWDFromBytes(t *testing.T) {
	cases := []struct {
		name string
		in   string
		want string
	}{
		{"clean", `{"type":"session_meta","payload":{"cwd":"/w/repo"}}`, "/w/repo"},
		{"malformed first line then valid",
			"{not json\n" + `{"type":"session_meta","payload":{"cwd":"/w/repo"}}`, "/w/repo"},
		{"no session_meta record",
			`{"type":"response_item","payload":{}}`, ""},
		{"blank lines around meta",
			"\n\n" + `{"type":"session_meta","payload":{"cwd":"/w/x"}}` + "\n", "/w/x"},
		{"empty input", "", ""},
		{"only whitespace", "   \n\t\n", ""},
		{"meta without cwd", `{"type":"session_meta","payload":{}}`, ""},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if got := codexCWDFromBytes([]byte(c.in)); got != c.want {
				t.Errorf("codexCWDFromBytes(%q) = %q, want %q", c.in, got, c.want)
			}
		})
	}
}

// The Codex harness walks the date-sharded YYYY/MM/DD/*.jsonl store, selects
// files by session_meta.cwd, and survives malformed / no-meta / empty sessions
// without aborting the walk. Returned Files are sorted, cwd-matched only.
func TestCodexHarnessSelectsByCWDAndSurvivesMalformed(t *testing.T) {
	root := t.TempDir()
	day := filepath.Join(root, "2026", "06", "26")
	if err := os.MkdirAll(day, 0o755); err != nil {
		t.Fatal(err)
	}
	write := func(name, content string) string {
		p := filepath.Join(day, name)
		if err := os.WriteFile(p, []byte(content), 0o644); err != nil {
			t.Fatal(err)
		}
		return p
	}
	match := write("match.jsonl",
		"{not json\n"+`{"timestamp":"2026-06-26T16:00:00Z","type":"session_meta","payload":{"cwd":"/w/repo"}}`)
	write("nometa.jsonl", `{"timestamp":"2026-06-26T16:00:00Z","type":"response_item","payload":{}}`)
	write("empty.jsonl", "")
	write("other.jsonl", `{"timestamp":"2026-06-26T16:00:00Z","type":"session_meta","payload":{"cwd":"/w/other"}}`)

	got := CodexHarness(root).Sources([]string{"/w/repo"})
	want := Sources{Files: []string{match}}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("Sources = %+v, want %+v", got, want)
	}

	// brain + repo both matched → both files, sorted.
	brain := write("brain.jsonl", `{"timestamp":"2026-06-26T16:00:00Z","type":"session_meta","payload":{"cwd":"/w/brain"}}`)
	got2 := CodexHarness(root).Sources([]string{"/w/brain", "/w/repo"})
	want2 := Sources{Files: []string{brain, match}} // sorted: brain.jsonl < match.jsonl
	if !reflect.DeepEqual(got2, want2) {
		t.Fatalf("Sources = %+v, want %+v", got2, want2)
	}

	// Empty root / empty cwds → no files, no panic.
	if g := CodexHarness("").Sources([]string{"/w/repo"}); len(g.Files) != 0 {
		t.Errorf("empty root should yield no files, got %+v", g)
	}
	if g := CodexHarness(root).Sources(nil); len(g.Files) != 0 {
		t.Errorf("no cwds should yield no files, got %+v", g)
	}
}

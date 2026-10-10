package judge

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func fixture(t *testing.T, name string) []byte {
	t.Helper()
	b, err := os.ReadFile(filepath.Join("testdata", "stream", name))
	if err != nil {
		t.Fatal(err)
	}
	return b
}

// #300: the stream yields every assistant message, so a block-less postscript
// can't erase the verdict, and the result's error flag is kept apart from the
// review's prose.
func TestReadStream(t *testing.T) {
	for _, c := range []struct {
		file          string
		stream        bool
		messages      int
		verdict       string // ParseVerdictBlock over Text(), "" for none
		findings      bool
		isError       bool
		textHasPrefix string
	}{
		{"clean.jsonl", true, 1, "FIX-THEN-SHIP", true, false, "Reviewed"},
		{"postscript.jsonl", true, 2, "FIX-THEN-SHIP", true, false, "Reviewed"},
		{"background_wait.jsonl", true, 1, "", false, false, "I'll wait"},
		{"api_error.jsonl", true, 1, "", false, true, "API Error"},
		{"truncated.jsonl", true, 1, "FIX-THEN-SHIP", true, false, "Reviewed"},
		{"not_a_stream.txt", false, 0, "FIX-THEN-SHIP", true, false, "Reviewed"},
	} {
		t.Run(c.file, func(t *testing.T) {
			r := ReadStream(fixture(t, c.file))
			if r.Stream != c.stream || len(r.Messages) != c.messages {
				t.Fatalf("stream %v messages %d: %+v", r.Stream, len(r.Messages), r)
			}
			text := r.Text()
			if !strings.HasPrefix(text, c.textHasPrefix) {
				t.Fatalf("text: %q", text)
			}
			got, _, _ := ParseVerdictBlock(text)
			if got != c.verdict {
				t.Fatalf("verdict %q, want %q", got, c.verdict)
			}
			if strings.Contains(text, "```findings") != c.findings {
				t.Fatalf("findings block presence: %q", text)
			}
			if (r.Result != nil && r.Result.IsError) != c.isError {
				t.Fatalf("result error: %+v", r.Result)
			}
		})
	}
	// A truncated stream keeps its unparsed tail as text after the messages.
	if r := ReadStream(fixture(t, "truncated.jsonl")); !strings.Contains(r.Text(), `"cut of`) {
		t.Fatalf("truncated tail lost: %q", r.Text())
	}
}

// HasVerdict recognises every recipe's verdict: a valid block or a VERDICT:
// line, plan-quality's CLEAN/INFO/FAILURE included.
func TestHasVerdict(t *testing.T) {
	for text, want := range map[string]bool{
		"```verdict\nverdict: SHIP\nconfidence: high\n```": true,
		"VERDICT: CLEAN":             true,
		"VERDICT: FAILURE (because)": true,
		"VERDICT: INFO":              true,
		"I'll wait for the background test run to notify.": false,
		"":                                false,
		"```verdict\nverdict: MAYBE\n```": false,
	} {
		if got := HasVerdict(text); got != want {
			t.Errorf("HasVerdict(%q) = %v, want %v", text, got, want)
		}
	}
}

// Over joined messages the VERDICT: line fallback agrees with the blocks:
// the latest wins.
func TestParseVerdictTokenTakesTheLastLine(t *testing.T) {
	if got, ok := ParseVerdictToken("VERDICT: SHIP\n\nlater I changed my mind\n\nVERDICT: REWORK"); !ok || got != "REWORK" {
		t.Fatalf("got %q %v", got, ok)
	}
}

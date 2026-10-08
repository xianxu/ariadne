package tracker

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/xianxu/ariadne/cmd/sdlc/internal/testfix"
)

func record(when, op string, paths ...string) string {
	return "\x01" + when + "\x00#1: tracker: update card\n\nTracker-Operation: " + op + "\n\x00\n" + strings.Join(paths, "\n") + "\n"
}

// #284: the newest claim-kind commit per wanted card; other operations, other
// cards and malformed records are skipped; reading stops once all are found.
func TestClaimTimes(t *testing.T) {
	a, b := "workshop/issue-cards/000001-a.md", "workshop/issue-cards/000002-b.md"
	stream := record("2026-10-07T12:00:00-07:00", "set-aaa", a) + // not a claim
		record("2026-10-07T11:00:00-07:00", "claim-bbb", a, b) + // a set claim
		"\x01garbage-without-fields" +
		record("2026-10-06T09:00:00-07:00", "reclaim-ccc", a) + // older: ignored
		record("not-a-time", "claim-ddd", b)
	got, err := ClaimTimes(strings.NewReader(stream), map[string]bool{a: true, b: true})
	if err != nil {
		t.Fatal(err)
	}
	want, _ := time.Parse(time.RFC3339, "2026-10-07T11:00:00-07:00")
	if !got[a].Equal(want) || !got[b].Equal(want) {
		t.Fatalf("got %v", got)
	}
	if got, _ := ClaimTimes(strings.NewReader(record("2026-10-07T11:00:00-07:00", "unclaim-x", a)), map[string]bool{a: true}); len(got) != 0 {
		t.Fatalf("an unclaim is not a claim: %v", got)
	}
}

// ClaimTimes stops reading once every wanted card is resolved.
func TestClaimTimesStopsEarly(t *testing.T) {
	a := "workshop/issue-cards/000001-a.md"
	r := &countingReader{r: strings.NewReader(record("2026-10-07T11:00:00-07:00", "claim-x", a) + strings.Repeat(record("2026-01-01T00:00:00Z", "claim-y", a), 20000))}
	if _, err := ClaimTimes(r, map[string]bool{a: true}); err != nil {
		t.Fatal(err)
	}
	if r.n > 1<<20 {
		t.Fatalf("read %d bytes after the answer was known", r.n)
	}
}

type countingReader struct {
	r interface{ Read([]byte) (int, error) }
	n int
}

func (c *countingReader) Read(p []byte) (int, error) {
	n, err := c.r.Read(p)
	c.n += n
	return n, err
}

// FuzzClaimTimes: any stream, malformed or truncated, never panics.
func FuzzClaimTimes(f *testing.F) {
	f.Add(record("2026-10-07T11:00:00-07:00", "claim-x", "p"))
	f.Add("\x01\x00\x00")
	f.Add("\x01x\x00Tracker-Operation: claim-\x00")
	f.Fuzz(func(t *testing.T, s string) {
		_, _ = ClaimTimes(strings.NewReader(s), map[string]bool{"p": true})
	})
}

// #284 BR-33: a history read that fails is an error, never an empty answer.
func TestRepositoryClaimTimesReportsAFailedRead(t *testing.T) {
	r, root, _ := fixture(t)
	if _, err := r.ClaimTimes(map[string]bool{testPath: true}); err != nil {
		t.Fatalf("a readable history: %v", err)
	}
	testfix.Git(t, root, "update-ref", "-d", "refs/remotes/publication/issue-tracker")
	if got, err := r.ClaimTimes(map[string]bool{testPath: true}); err == nil {
		t.Fatalf("a failed read answered %v", got)
	}
}

// #284 BR-39: a parse error mid-stream (a record beyond the scanner's 16 MB)
// stops git and returns the error — it never waits on a full pipe.
func TestRepositoryClaimTimesStopsGitOnAParseError(t *testing.T) {
	r, root, _ := fixture(t)
	// A commit whose message (17 MB) is past the scanner's limit, made with
	// plumbing (an argv message that size is refused by the OS) on top of the
	// local tracking ref the history read walks.
	tracking := "refs/remotes/publication/issue-tracker"
	msgFile := filepath.Join(t.TempDir(), "msg")
	if err := os.WriteFile(msgFile, []byte("#1: tracker: update card\n\n"+strings.Repeat("x", 17<<20)+"\n\nTracker-Operation: claim-huge\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	// The commit must touch a card, or the log's path limit skips it.
	blobFile := filepath.Join(t.TempDir(), "card")
	if err := os.WriteFile(blobFile, []byte(strings.Replace(testCard, "Tracker title", "Huge", 1)), 0o600); err != nil {
		t.Fatal(err)
	}
	blob := strings.TrimSpace(testfix.Capture(t, root, "hash-object", "-w", blobFile))
	index := filepath.Join(t.TempDir(), "index")
	plumb := func(args ...string) string {
		cmd := exec.Command("git", append([]string{"-C", root}, args...)...)
		cmd.Env = append(os.Environ(), "GIT_INDEX_FILE="+index)
		out, err := cmd.Output()
		if err != nil {
			t.Fatalf("git %v: %v", args, err)
		}
		return strings.TrimSpace(string(out))
	}
	plumb("read-tree", tracking)
	plumb("update-index", "--cacheinfo", "100644,"+blob+","+testPath)
	tree := plumb("write-tree")
	parent := strings.TrimSpace(testfix.Capture(t, root, "rev-parse", tracking))
	commit := strings.TrimSpace(testfix.Capture(t, root, "commit-tree", tree, "-p", parent, "-F", msgFile))
	testfix.Git(t, root, "update-ref", tracking, commit)
	done := make(chan error, 1)
	go func() {
		_, err := r.ClaimTimes(map[string]bool{"workshop/issue-cards/000999-absent.md": true})
		done <- err
	}()
	select {
	case err := <-done:
		if err == nil {
			t.Fatal("an oversized record must be a read error")
		}
	case <-time.After(30 * time.Second):
		t.Fatal("ClaimTimes hung on a parse error")
	}
}

package tracker

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/xianxu/ariadne/cmd/sdlc/internal/issue"
	"github.com/xianxu/ariadne/cmd/sdlc/internal/testfix"
)

// Cards are exact blob projections, not checkout files. Neither external
// attributes nor an unrelated caller checkout may transform their bytes.
func TestRepositoryUpdateCardPreservesExactBytesUnderAttributes(t *testing.T) {
	for _, source := range []string{"external", "worktree", "index"} {
		t.Run(source, func(t *testing.T) {
			r, root, _ := fixture(t)
			snapshot, err := r.Snapshot()
			if err != nil {
				t.Fatal(err)
			}
			old, _ := snapshot.Card("000252")

			// The clean filter is deliberately non-idempotent; applying it even
			// once changes a valid card's title. Explicit text also normalizes
			// CRLF within the otherwise LF document's Problem section.
			testfix.Git(t, root, "config", "filter.tracker-test.clean", "sed 's/Tracker title/Filtered Tracker title/g'")
			testfix.Git(t, root, "config", "filter.tracker-test.required", "true")
			attributes := "*.md text eol=lf filter=tracker-test\n"
			attributesPath := filepath.Join(root, ".gitattributes")
			if source == "external" {
				attributesPath = filepath.Join(t.TempDir(), "attributes")
				testfix.Git(t, root, "config", "core.attributesFile", attributesPath)
			}
			if err := os.WriteFile(attributesPath, []byte(attributes), 0o600); err != nil {
				t.Fatal(err)
			}
			if source == "index" {
				testfix.Git(t, root, "add", ".gitattributes")
				if err := os.Remove(attributesPath); err != nil {
					t.Fatal(err)
				}
			}
			if attrs := testfix.Capture(t, root, "check-attr", "filter", "--", testPath); !strings.Contains(attrs, "tracker-test") {
				t.Fatalf("fixture did not activate caller attributes: %s", attrs)
			}
			beforeHEAD := testfix.Capture(t, root, "rev-parse", "HEAD")
			beforeStatus := testfix.Capture(t, root, "status", "--porcelain=v1")
			indexPath := strings.TrimSpace(testfix.Capture(t, root, "rev-parse", "--git-path", "index"))
			if !filepath.IsAbs(indexPath) {
				indexPath = filepath.Join(root, indexPath)
			}
			beforeIndex, err := os.ReadFile(indexPath)
			if err != nil {
				t.Fatal(err)
			}

			input := []byte(strings.Replace(testCard, "status: open", "status: working", 1) + "Preserve this CRLF payload.\r\n")
			if _, err := issue.ParseCard(input); err != nil {
				t.Fatalf("fixture card must be valid before writing: %v", err)
			}
			if err := r.UpdateCard(old, input, "exact-bytes-"+source, func(string, string) error { return nil }); err != nil {
				t.Fatal(err)
			}

			// Read the remote object itself: checkout conversion cannot mask a
			// transformed write or make this assertion match by round-tripping.
			remote := strings.TrimSpace(testfix.Capture(t, root, "remote", "get-url", "publication"))
			stored, err := exec.Command("git", "--git-dir", remote, "cat-file", "blob", "refs/heads/issue-tracker:"+testPath).Output()
			if err != nil {
				t.Fatal(err)
			}
			if !bytes.Equal(stored, input) {
				t.Fatalf("%s attributes transformed tracker bytes:\nwant %q\ngot  %q", source, input, stored)
			}
			afterIndex, err := os.ReadFile(indexPath)
			if err != nil {
				t.Fatal(err)
			}
			if !bytes.Equal(beforeIndex, afterIndex) || beforeHEAD != testfix.Capture(t, root, "rev-parse", "HEAD") || beforeStatus != testfix.Capture(t, root, "status", "--porcelain=v1") {
				t.Fatal("tracker write changed caller index, HEAD, or worktree state")
			}
			if source != "index" {
				got, err := os.ReadFile(attributesPath)
				if err != nil || string(got) != attributes {
					t.Fatal("tracker write changed caller attributes")
				}
			}
		})
	}
}

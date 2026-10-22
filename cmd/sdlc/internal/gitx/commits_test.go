package gitx

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/xianxu/ariadne/cmd/sdlc/internal/testfix"
)

func TestReadCommitObjectsIgnoresReplacement(t *testing.T) {
	repo := testfix.Repo(t, testfix.InitialCommit())
	testfix.Git(t, repo, "commit", "--allow-empty", "-qm", "#548: original")
	original := strings.TrimSpace(testfix.Capture(t, repo, "rev-parse", "HEAD"))
	want := testfix.Capture(t, repo, "cat-file", "commit", original)
	testfix.Git(t, repo, "commit", "--allow-empty", "-qm", "#999: replacement")
	replacement := strings.TrimSpace(testfix.Capture(t, repo, "rev-parse", "HEAD"))
	testfix.Git(t, repo, "replace", original, replacement)
	if replaced := testfix.Capture(t, repo, "cat-file", "commit", original); replaced == want {
		t.Fatal("replacement control did not replace original")
	}
	got, err := ReadCommitObjects(t.Context(), repo, []string{original, replacement, original})
	if err != nil || len(got) != 3 {
		t.Fatalf("read: %#v, %v", got, err)
	}
	if got[0].OID != original || string(got[0].Content) != want || !bytes.Contains(got[1].Content, []byte("#999: replacement")) || string(got[2].Content) != want {
		t.Fatalf("wrong immutable objects: %#v", got)
	}
}

func TestReadCommitObjectsRejectsUnavailableAndNoncommit(t *testing.T) {
	repo := testfix.Repo(t, testfix.InitialCommit())
	tree := strings.TrimSpace(testfix.Capture(t, repo, "rev-parse", "HEAD^{tree}"))
	for _, oid := range []string{strings.Repeat("a", 40), tree} {
		if got, err := ReadCommitObjects(t.Context(), repo, []string{oid}); err == nil || got != nil {
			t.Fatalf("accepted %s: %#v, %v", oid, got, err)
		}
	}
	ctx, cancel := context.WithCancel(t.Context())
	cancel()
	if _, err := ReadCommitObjects(ctx, repo, []string{tree}); !errors.Is(err, context.Canceled) {
		t.Fatalf("cancellation lost: %v", err)
	}
}

func TestReadCommitObjectsInputValidation(t *testing.T) {
	for _, oid := range []string{"HEAD", "", strings.Repeat("0", 40), strings.Repeat("A", 40), strings.Repeat("a", 39), strings.Repeat("a", 40) + "\n", strings.Repeat("z", 64)} {
		if _, err := ReadCommitObjects(t.Context(), "nonexistent", []string{oid}); err == nil || !strings.Contains(err.Error(), "object ID") {
			t.Errorf("invalid ID %q: %v", oid, err)
		}
	}
	if got, err := ReadCommitObjects(t.Context(), "nonexistent", nil); err != nil || got != nil {
		t.Fatalf("empty read: %#v, %v", got, err)
	}
	if _, err := ReadCommitObjects(t.Context(), "nonexistent", make([]string, 129)); err == nil || !strings.Contains(err.Error(), "128") {
		t.Fatalf("oversized batch: %v", err)
	}
}

func TestParseCommitObjectsFrames(t *testing.T) {
	oid := strings.Repeat("a", 40)
	frame := oid + " commit 3\na\x00b\n"
	for name, raw := range map[string]string{
		"empty": "", "missing": oid + " missing\n", "wrong oid": strings.Repeat("b", 40) + " commit 0\n\n",
		"wrong type": oid + " blob 0\n\n", "negative": oid + " commit -1\n", "overflow": oid + " commit 18446744073709551616\n",
		"oversized": fmt.Sprintf("%s commit %d\n", oid, (1<<20)+1), "truncated": oid + " commit 3\nab",
		"delimiter": oid + " commit 3\na\x00bX", "trailing": frame + "x", "extra frame": frame + frame,
		"extra field": oid + " commit 3 extra\na\x00b\n", "unterminated header": oid + " commit 0",
	} {
		t.Run(name, func(t *testing.T) {
			if got, err := parseCommitObjects([]byte(raw), []string{oid}); err == nil || got != nil {
				t.Fatalf("accepted invalid frame: %#v, %v", got, err)
			}
		})
	}
	for _, size := range []int{0, 3, 1 << 20} {
		content := bytes.Repeat([]byte{'x'}, size)
		raw := append([]byte(fmt.Sprintf("%s commit %d\n", oid, size)), content...)
		raw = append(raw, '\n')
		got, err := parseCommitObjects(raw, []string{oid})
		if err != nil || len(got) != 1 || got[0].OID != oid || !bytes.Equal(got[0].Content, content) {
			t.Fatalf("size %d: %#v, %v", size, got, err)
		}
	}
	sha256 := strings.Repeat("b", 64)
	if _, err := parseCommitObjects([]byte(sha256+" commit 0\n\n"), []string{sha256}); err != nil {
		t.Fatal(err)
	}
}

func TestParseCommitObjectsAggregateIncludesFraming(t *testing.T) {
	oid := strings.Repeat("a", 40)
	var raw bytes.Buffer
	var ids []string
	for range 32 {
		fmt.Fprintf(&raw, "%s commit %d\n", oid, 1<<20)
		raw.Write(bytes.Repeat([]byte{'x'}, 1<<20))
		raw.WriteByte('\n')
		ids = append(ids, oid)
	}
	if _, err := parseCommitObjects(raw.Bytes(), ids); !errors.Is(err, ErrOutputLimit) {
		t.Fatalf("framing must count toward aggregate: %v", err)
	}
	// Retain exactly the byte ceiling, adjusting the last frame's content size.
	frameHeader := fmt.Sprintf("%s commit %d\n", oid, 1<<20)
	lastSize := (1 << 20) - (raw.Len() - (32 << 20))
	start := 31 * (len(frameHeader) + (1 << 20) + 1)
	bounded := append([]byte(nil), raw.Bytes()[:start]...)
	bounded = append(bounded, []byte(fmt.Sprintf("%s commit %d\n", oid, lastSize))...)
	bounded = append(bounded, bytes.Repeat([]byte{'x'}, lastSize)...)
	bounded = append(bounded, '\n')
	if len(bounded) != 32<<20 {
		t.Fatalf("invalid boundary fixture: %d", len(bounded))
	}
	if got, err := parseCommitObjects(bounded, ids); err != nil || len(got) != 32 {
		t.Fatalf("exact aggregate limit: %d commits, %v", len(got), err)
	}
}

func TestReadCommitObjectsEnforcesProcessOutputCap(t *testing.T) {
	dir := t.TempDir()
	// Reuse the executable Git seam: the producer emits beyond the cap rather
	// than handing an already allocated oversized response to the parser.
	script := "#!/bin/sh\ncat >/dev/null\nexec dd if=/dev/zero bs=1048576 count=33 2>/dev/null\n"
	if err := os.WriteFile(filepath.Join(dir, "git"), []byte(script), 0755); err != nil {
		t.Fatal(err)
	}
	t.Setenv("PATH", dir+string(os.PathListSeparator)+os.Getenv("PATH"))
	got, err := ReadCommitObjects(t.Context(), dir, []string{strings.Repeat("a", 40)})
	if got != nil || !errors.Is(err, ErrOutputLimit) {
		t.Fatalf("process cap: %#v, %v", got, err)
	}
}

func TestReadCommitObjectsBatchCardinalityBoundary(t *testing.T) {
	repo := testfix.Repo(t, testfix.InitialCommit())
	oid := strings.TrimSpace(testfix.Capture(t, repo, "rev-parse", "HEAD"))
	ids := make([]string, 128)
	for i := range ids {
		ids[i] = oid
	}
	got, err := ReadCommitObjects(t.Context(), repo, ids)
	if err != nil || len(got) != 128 {
		t.Fatalf("128-object boundary: %d, %v", len(got), err)
	}
	ids[127] = strings.Repeat("a", 40)
	if got, err := ReadCommitObjects(t.Context(), repo, ids); err == nil || got != nil {
		t.Fatalf("partial result on missing object: %d, %v", len(got), err)
	}
}

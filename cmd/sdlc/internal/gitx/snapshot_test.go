package gitx

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/xianxu/ariadne/cmd/sdlc/internal/testfix"
)

func TestSnapshotPinsTreeAndBatchReadsLiteralBytes(t *testing.T) {
	repo, _ := trunkFixture(t, "seed\n")
	tf, err := NewTrunkFileContext(context.Background(), repo, "origin", "main")
	if err != nil {
		t.Fatal(err)
	}
	want := "literal $Format:%H$\n\x00tail\n"
	err = tf.UpdateMany("snapshot", func(*TrunkView) (TrunkWrite, error) {
		return TrunkWrite{Write: map[string][]byte{
			"cards/one.md": []byte(want), ".gitattributes": []byte("cards/* export-ignore export-subst\n"),
		}}, nil
	})
	if err != nil {
		t.Fatal(err)
	}
	view, err := tf.Snapshot()
	if err != nil {
		t.Fatal(err)
	}
	all, err := view.Files("")
	if err != nil || len(all) < 2 {
		t.Fatalf("root snapshot = %#v, %v", all, err)
	}
	err = tf.Update("cards/one.md", "later", func([]byte) ([]byte, error) { return []byte("changed"), nil })
	if err != nil {
		t.Fatal(err)
	}
	files, err := view.Files("cards")
	if err != nil {
		t.Fatal(err)
	}
	if len(files) != 1 || files[0].Path != "cards/one.md" || string(files[0].Content) != want || files[0].Mode != "100644" {
		t.Fatalf("snapshot = %#v", files)
	}
	blob, err := view.ReadBlob(files[0].OID)
	if err != nil || string(blob) != want {
		t.Fatalf("blob = %q, %v", blob, err)
	}
}

func FuzzSnapshotFrames(f *testing.F) {
	oid := strings.Repeat("a", 40)
	f.Add([]byte("100644 blob "+oid+"\tcards/one.md\x00"), []byte(oid+" blob 1\nx\n"))
	f.Add([]byte("garbage"), []byte("garbage"))
	f.Fuzz(func(t *testing.T, tree, blobs []byte) {
		files, err := parseSnapshotTree(tree)
		if err != nil {
			return
		}
		result, err := parseSnapshotBlobs(blobs, files)
		if err != nil {
			return
		}
		if len(result) != len(files) {
			t.Fatal("changed snapshot cardinality")
		}
		for i := range files {
			if result[i].OID != files[i].OID || result[i].Path != files[i].Path || len(result[i].Content) > 1<<20 {
				t.Fatal("accepted mismatched or oversized snapshot")
			}
		}
	})
}

func TestSnapshotBatchRejectsMalformedFrames(t *testing.T) {
	for _, raw := range []string{"", "abc missing\n", "abc blob -1\n", "abc blob 5\nxx", "abc blob 1\nxX"} {
		if _, err := parseSnapshotBlobs([]byte(raw), []TreeFile{{OID: "abc"}}); err == nil {
			t.Fatalf("accepted %q", raw)
		}
	}
}

// #290: a snapshot right after an unchanged presence probe reads the tip the
// probe saw without fetching; a moved remote is fetched; the probe's tip is used
// once; and tracker fetches never trigger automatic maintenance.
func TestSnapshotSkipsTheFetchWhenTheTipIsUnchanged(t *testing.T) {
	repo, origin := trunkFixture(t, "seed\n")
	trace := filepath.Join(t.TempDir(), "trace")
	t.Setenv("GIT_TRACE", trace)
	fetches := func() (n int, noMaint bool) {
		raw, _ := os.ReadFile(trace)
		noMaint = true
		for _, l := range strings.Split(string(raw), "\n") {
			if strings.Contains(l, "built-in: git fetch") {
				n++
				noMaint = noMaint && strings.Contains(l, "--no-auto-maintenance")
			}
		}
		return n, noMaint
	}
	tf, err := NewTrunkFileContext(context.Background(), repo, "origin", "main")
	if err != nil {
		t.Fatal(err)
	}
	tip := strings.TrimSpace(testfix.Capture(t, origin, "rev-parse", "main"))
	if ok, err := tf.RemoteExists(); !ok || err != nil {
		t.Fatalf("presence: %v %v", ok, err)
	}
	view, err := tf.Snapshot()
	if n, _ := fetches(); err != nil || n != 0 || view.Ref() != tip {
		t.Fatalf("unchanged tip: %d fetches, ref %s want %s, %v", n, view.Ref(), tip, err)
	}

	other := filepath.Join(t.TempDir(), "other")
	testfix.Git(t, "", "clone", "-q", origin, other)
	testfix.Git(t, other, "-c", "user.name=t", "-c", "user.email=t@t", "commit", "-q", "--allow-empty", "-m", "moved")
	testfix.Git(t, other, "push", "-q", "origin", "main")
	moved := strings.TrimSpace(testfix.Capture(t, origin, "rev-parse", "main"))
	if _, err := tf.RemoteExists(); err != nil {
		t.Fatal(err)
	}
	view, err = tf.Snapshot()
	if n, noMaint := fetches(); err != nil || n != 1 || !noMaint || view.Ref() != moved {
		t.Fatalf("moved tip: %d fetches (no-maintenance %v), ref %s want %s, %v", n, noMaint, view.Ref(), moved, err)
	}

	if _, err := tf.Snapshot(); err != nil {
		t.Fatal(err)
	}
	if n, _ := fetches(); n != 2 {
		t.Fatalf("a probe's tip was reused by a later snapshot: %d fetches, want 2", n)
	}
}

package gitx

import (
	"context"
	"strings"
	"testing"
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

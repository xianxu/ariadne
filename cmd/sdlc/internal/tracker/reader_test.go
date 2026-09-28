package tracker

import (
	"bytes"
	"errors"
	"fmt"
	"strings"
	"testing"

	"github.com/xianxu/ariadne/cmd/sdlc/internal/gitx"
	"github.com/xianxu/ariadne/cmd/sdlc/internal/issue"
)

// Build a valid tracker whose cat-file batch response is exactly budget bytes.
// The OID/header delimiters consume budget even though they are not card data.
func budgetFiles(t *testing.T, budget int) []gitx.TreeFile {
	t.Helper()
	files := []gitx.TreeFile{file(ManifestPath, string(ManifestBytes()))}
	wire := func(size int) int { return len(fmt.Sprintf("%040d blob %d\n", 0, size)) + size + 1 }
	remaining := budget - wire(len(files[0].Content))
	for id := 1; remaining > 0; id++ {
		size := 1 << 20
		for wire(size) > remaining {
			size -= wire(size) - remaining
		}
		if size < len(testCard) {
			t.Fatal("budget fixture leaves insufficient card space")
		}
		key := fmt.Sprintf("%06d", id)
		raw := strings.ReplaceAll(testCard, "000252", key)
		raw += strings.Repeat("x", size-len(raw))
		files = append(files, file("workshop/issue-cards/"+key+"-test.md", raw))
		remaining -= wire(len(raw))
	}
	return files
}

func TestSnapshotStorageEnvelope(t *testing.T) {
	files := budgetFiles(t, 32<<20)
	snapshot, err := parseSnapshot("tip", files)
	if err != nil {
		t.Fatalf("exact wire budget refused: %v", err)
	}
	last := len(files) - 1
	current, _ := snapshot.Card(fmt.Sprintf("%06d", last))
	if err := snapshot.validateReplacement(current, current.Raw); err != nil {
		t.Fatalf("same-size replacement refused: %v", err)
	}
	if err := snapshot.validateReplacement(current, append(bytes.Clone(current.Raw), 'x')); !errors.Is(err, gitx.ErrOutputLimit) {
		t.Fatalf("growing replacement at batch limit: %v", err)
	}
	if err := snapshot.validateReplacement(current, current.Raw[:len(current.Raw)-1]); err != nil {
		t.Fatalf("shrinking replacement refused: %v", err)
	}
	files[last] = file(files[last].Path, string(files[last].Content)+"x")
	if _, err := parseSnapshot("tip", files); !errors.Is(err, gitx.ErrOutputLimit) {
		t.Fatalf("wire budget overflow: %v", err)
	}
	oversize := file(testPath, testCard+strings.Repeat("x", 1<<20))
	if _, err := parseSnapshot("tip", []gitx.TreeFile{file(ManifestPath, string(ManifestBytes())), oversize}); !errors.Is(err, gitx.ErrOutputLimit) {
		t.Fatalf("blob overflow: %v", err)
	}
	tooMany := make([]gitx.TreeFile, 10002)
	if _, err := parseSnapshot("tip", tooMany); !errors.Is(err, gitx.ErrOutputLimit) {
		t.Fatalf("entry overflow: %v", err)
	}
}

const testCard = "---\nid: 000252\nstatus: open\n---\n\n# Tracker title\n\n## Problem\nOriginal report.\n"
const testPath = "workshop/issue-cards/000252-test.md"

func file(path, raw string) gitx.TreeFile {
	oid, _ := issue.CardBlobOID([]byte(raw), "sha1")
	return gitx.TreeFile{Path: path, Mode: "100644", OID: oid, Content: []byte(raw)}
}

func BenchmarkSnapshotTenThousandCards(b *testing.B) {
	files := make([]gitx.TreeFile, 0, 10001)
	files = append(files, file(ManifestPath, string(ManifestBytes())))
	for id := 1; id <= 10000; id++ {
		key := fmt.Sprintf("%06d", id)
		files = append(files, file("workshop/issue-cards/"+key+"-sample.md", strings.ReplaceAll(testCard, "000252", key)))
	}
	b.ReportAllocs()
	b.ResetTimer()
	for range b.N {
		snapshot, err := parseSnapshot("tip", files)
		if err != nil || snapshot.MaxID() != 10000 {
			b.Fatalf("snapshot: max=%d err=%v", snapshot.MaxID(), err)
		}
	}
}

func TestSnapshotRejectsMalformedAuthority(t *testing.T) {
	for name, files := range map[string][]gitx.TreeFile{
		"missing manifest":       {file(testPath, testCard)},
		"unknown version":        {file(ManifestPath, `{"version":2}`)},
		"duplicate version":      {file(ManifestPath, `{"version":1,"version":2}`)},
		"trailing JSON":          {file(ManifestPath, `{"version":1}{}`)},
		"unknown manifest field": {file(ManifestPath, `{"version":1,"other":true}`)},
		"duplicate id":           {file(ManifestPath, string(ManifestBytes())), file(testPath, testCard), file("workshop/issue-cards/000252-other.md", testCard)},
		"mismatched id":          {file(ManifestPath, string(ManifestBytes())), file("workshop/issue-cards/000251-test.md", testCard)},
		"nested path":            {file(ManifestPath, string(ManifestBytes())), file("workshop/issue-cards/nested/000252-test.md", testCard)},
		"unexpected file":        {file(ManifestPath, string(ManifestBytes())), file("README", "unexpected")},
		"malformed card":         {file(ManifestPath, string(ManifestBytes())), file(testPath, strings.Replace(testCard, "status: open", "status: impossible", 1))},
	} {
		t.Run(name, func(t *testing.T) {
			if _, err := parseSnapshot("tip", files); err == nil {
				t.Fatal("accepted malformed tracker")
			}
		})
	}
	for _, mode := range []string{"120000", "100755", "160000"} {
		t.Run(mode, func(t *testing.T) {
			card := file(testPath, testCard)
			card.Mode = mode
			if _, err := parseSnapshot("tip", []gitx.TreeFile{file(ManifestPath, string(ManifestBytes())), card}); err == nil {
				t.Fatal("accepted non-ordinary card")
			}
		})
	}
}

func TestSnapshotRecordsAreIsolated(t *testing.T) {
	files := []gitx.TreeFile{file(ManifestPath, string(ManifestBytes())), file(testPath, testCard)}
	snapshot, err := parseSnapshot("tip", files)
	if err != nil {
		t.Fatal(err)
	}
	record, ok := snapshot.Card("000252")
	if !ok || record.Path != testPath || record.Card.Title != "Tracker title" || snapshot.Ref() != "tip" {
		t.Fatalf("snapshot: %+v", record)
	}
	record.Raw[0] = 'x'
	record.Card.Title = "changed"
	files[1].Content[0] = 'x'
	again, _ := snapshot.Card("000252")
	if string(again.Raw) != testCard || again.Card.Title != "Tracker title" {
		t.Fatal("snapshot mutated through returned/source record")
	}
	if snapshot.MaxID() != 252 || len(snapshot.Records()) != 1 {
		t.Fatal("wrong inventory")
	}
}

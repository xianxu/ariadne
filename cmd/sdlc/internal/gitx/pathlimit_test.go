package gitx

import (
	"strings"
	"testing"
)

func TestCheckPathLimits(t *testing.T) {
	ok := "workshop/issues/000001-" + strings.Repeat("a", 245) + ".md" // 255-byte name
	if err := CheckPathLimits([]string{ok}); err != nil {
		t.Fatal(err)
	}
	long := "workshop/issues/000001-" + strings.Repeat("a", 246) + ".md"
	err := CheckPathLimits([]string{ok, long})
	if err == nil || !strings.Contains(err.Error(), "256 bytes") || !strings.Contains(err.Error(), "255-byte") {
		t.Fatalf("err = %v", err)
	}
	if err := CheckPathLimits([]string{strings.Repeat("d", 256) + "/x.md"}); err == nil {
		t.Fatal("an over-long directory component passed")
	}
}

// #566: parley.nvim#421's 331-byte name was pushed to main and broke every
// checkout. The refusal precedes any candidate, so origin never moves.
func TestUpdateManyRefusesOverlongNameBeforePush(t *testing.T) {
	repo, origin := trunkFixture(t, "seed\n")
	before := commitCount(t, origin)
	tf, err := NewTrunkFile(repo, "origin", "main")
	if err != nil {
		t.Fatal(err)
	}
	name := "workshop/issues/000421-" + strings.Repeat("x", 321) + ".md"
	err = tf.UpdateMany("file it", func(*TrunkView) (TrunkWrite, error) {
		return TrunkWrite{Write: map[string][]byte{name: []byte("x\n")}}, nil
	})
	if err == nil || !strings.Contains(err.Error(), "file-name limit") {
		t.Fatalf("err = %v", err)
	}
	if got := commitCount(t, origin) - before; got != 0 {
		t.Fatalf("origin gained %d commits", got)
	}
}

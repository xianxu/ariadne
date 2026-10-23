package tracker

import "testing"

// Every combination of the two refs maps to one step or a refusal (#566). The
// expected table is written out, not derived from the implementation.
func TestNextRenameStep(t *testing.T) {
	want := map[RenameObservation]RenameStep{
		{TrackerOld: true, MainOld: true}: RenameCard,
		{TrackerNew: true, MainOld: true}: RenameDetails,
		{TrackerOld: true, MainNew: true}: RenameCard,
		{TrackerNew: true, MainNew: true}: RenameDone,
	}
	for i := 0; i < 16; i++ {
		o := RenameObservation{TrackerOld: i&1 != 0, TrackerNew: i&2 != 0, MainOld: i&4 != 0, MainNew: i&8 != 0}
		got, err := NextRenameStep(o)
		if step, ok := want[o]; ok {
			if err != nil || got != step {
				t.Errorf("%+v → %v, %v; want %v", o, got, err, step)
			}
		} else if err == nil {
			t.Errorf("%+v → %v; want a refusal", o, got)
		}
	}
}

func TestDetailNames(t *testing.T) {
	paths := []string{"w/000008-old.md", "w/000008-new.md", "w/000009-old.md", "w/README.md"}
	olds, hasNew := DetailNames(paths, "000008", "w/000008-new.md")
	if !hasNew || len(olds) != 1 || olds[0] != "w/000008-old.md" {
		t.Fatalf("olds %v hasNew %v", olds, hasNew)
	}
	if olds, hasNew := DetailNames(paths[2:], "000008", "w/000008-new.md"); hasNew || len(olds) != 0 {
		t.Fatalf("olds %v hasNew %v", olds, hasNew)
	}
}

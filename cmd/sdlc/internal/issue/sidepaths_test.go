package issue

import (
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

func TestPathsByIDSideTreesHaveNoIdentity(t *testing.T) {
	listing := strings.Join([]string{
		"workshop/issues/000424-real.md",
		"workshop/issues/000424-real.side",
		"workshop/issues/000424-real.side/000424-context.md",
		"workshop/issues/000424-real.side/sub/000999-evidence.json",
		"workshop/issues/000888-orphan.side",
		"workshop/history/issues/000424-duplicate.md",
		"workshop/issues/000424-real.md", // repeated paths are not duplicates
		"workshop/issues/000425-existing-parser.txt",
		"workshop/issues/000426-similar.sidecar",
		"workshop/issues/1234567-too-many-digits.md",
	}, "\n")
	want := map[int][]string{
		424: {"workshop/issues/000424-real.md", "workshop/history/issues/000424-duplicate.md"},
		425: {"workshop/issues/000425-existing-parser.txt"},
		426: {"workshop/issues/000426-similar.sidecar"},
	}
	got := PathsByID(listing)
	if !reflect.DeepEqual(got, want) {
		t.Errorf("PathsByID = %#v, want %#v", got, want)
	}
	duplicates := DuplicatesIn(got)
	if len(duplicates) != 1 || duplicates[0].ID != 424 || len(duplicates[0].Paths) != 2 {
		t.Errorf("real duplicate lost or side tree counted: %#v", duplicates)
	}
}

func TestLocalPathsByIDSkipsDirectoriesAndSideMembers(t *testing.T) {
	issues := filepath.Join(t.TempDir(), "issues")
	side := filepath.Join(issues, "000424-real.side")
	mustMkdir(t, issues, side, filepath.Join(issues, "000999-unrelated-directory"))
	writeFiles(t, issues, "000424-real.md", "000425-existing-parser.txt")
	writeFiles(t, side, "000888-evidence.md")
	got, found, err := LocalPathsByID([]string{issues, side})
	if err != nil {
		t.Fatal(err)
	}
	want := map[int][]string{
		424: {filepath.Join(issues, "000424-real.md")},
		425: {filepath.Join(issues, "000425-existing-parser.txt")},
	}
	if found != 2 || !reflect.DeepEqual(got, want) {
		t.Errorf("LocalPathsByID = %#v, found %d; want %#v, found 2", got, found, want)
	}
}

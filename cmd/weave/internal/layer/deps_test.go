package layer

import (
	"reflect"
	"testing"
)

// ParseDeps extracts the substrate edges from a construct/deps file's text,
// in file order — the sole layer-edge source feeding Resolve. Grammar ported
// from construct/scripts/lib-deps.sh:deps_substrate_targets (ARCH-DRY): '#'
// comments and blanks ignored, whitespace-split columns, only `substrate` rows
// contribute, `data` rows are skipped.

func TestParseDepsSubstrateRow(t *testing.T) {
	got, err := ParseDeps("substrate ../ariadne\n")
	if err != nil {
		t.Fatalf("ParseDeps: unexpected error: %v", err)
	}
	want := []string{"../ariadne"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("ParseDeps = %v, want %v", got, want)
	}
}

func TestParseDepsMultipleSubstratesInOrder(t *testing.T) {
	got, err := ParseDeps("substrate ../ariadne\nsubstrate ../nous\n")
	if err != nil {
		t.Fatalf("ParseDeps: unexpected error: %v", err)
	}
	want := []string{"../ariadne", "../nous"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("ParseDeps = %v, want %v", got, want)
	}
}

func TestParseDepsIgnoresDataCommentsBlanks(t *testing.T) {
	// data rows, '#' comments (whole-line and trailing), and blank lines are
	// all dropped — only the two substrate rows survive, in order.
	content := `# layer deps for this repo
substrate ../ariadne

data git@github.com:xianxu/you-decide.git data/life/politics/you-decide
substrate ../nous   # the nous edge
`
	got, err := ParseDeps(content)
	if err != nil {
		t.Fatalf("ParseDeps: unexpected error: %v", err)
	}
	want := []string{"../ariadne", "../nous"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("ParseDeps = %v, want %v", got, want)
	}
}

func TestParseDepsEmptyContent(t *testing.T) {
	got, err := ParseDeps("")
	if err != nil {
		t.Fatalf("ParseDeps: unexpected error: %v", err)
	}
	if len(got) != 0 {
		t.Fatalf("ParseDeps(\"\") = %v, want empty", got)
	}
}

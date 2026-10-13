package shelllint

import (
	"strings"
	"testing"
)

func TestMktempFindings(t *testing.T) {
	cases := []struct {
		name, src, reason string // reason "" = no finding
	}{
		// Unsafe: macOS mktemp without a path template ignores TMPDIR (#342).
		{"bare -d", `WORK=$(mktemp -d)`, "template"},
		{"bare file", `export L=$(mktemp)`, "template"},
		{"redirect then or", `work="$(mktemp -d 2>/dev/null || true)"`, "template"},
		{"makefile escaped", `	d1="$$(mktemp -d)"; d2="$$(mktemp -d "$${TMPDIR:-/tmp}/b.XXXXXX")"; \`, "template"},
		{"-t prefix", `d=$(mktemp -d -t foo.XXXXXX)`, "template"},
		{"literal /tmp template", `out="$(mktemp /tmp/s.XXXXXX)"`, "template"},
		{"command position", `mktemp -d`, "template"},
		{"after &&", `true && mktemp -d`, "template"},
		{"cd into substitution", `ROOT="$(cd "$(mktemp -d "${TMPDIR:-/tmp}/c.XXXXXX")" && pwd -P)"`, "cd"},
		// Safe.
		{"tmpdir template", `d=$(mktemp -d "${TMPDIR:-/tmp}/x.XXXXXX") || exit 1`, ""},
		{"unbraced tmpdir", `f=$(mktemp "$TMPDIR/x.XXXXXX")`, ""},
		{"makefile tmpdir", `	sync_bin=$$(mktemp "$${TMPDIR:-/tmp}/s.XXXXXX"); \`, ""},
		{"inside a made dir", `f=$(mktemp "$work/part.XXXX")`, ""},
		{"comment line", `# WORK=$(mktemp -d) is the bug`, ""},
		{"trailing comment", `x=1 # then mktemp -d`, ""},
		{"word in a string", `echo "run mktemp -d to see"`, ""},
		{"variable invocation", `"$mktemp" -d`, ""},
		{"identifier", `real_mktemp=1; c.mktemp=2`, ""},
		{"cd into a checked var", `d=$(mktemp -d "${TMPDIR:-/tmp}/x.XXXXXX") || exit 1; cd "${d:?}"`, ""},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got := MktempFindings(c.src)
			if c.reason == "" {
				if len(got) != 0 {
					t.Fatalf("unexpected findings %+v", got)
				}
				return
			}
			if len(got) != 1 || got[0].Line != 1 || !strings.Contains(got[0].Reason, c.reason) {
				t.Fatalf("findings %+v, want one on line 1 mentioning %q", got, c.reason)
			}
		})
	}
}

func TestMktempFindingsLineNumbers(t *testing.T) {
	src := "#!/bin/sh\nset -eu\nd=$(mktemp -d)\n# mktemp -d\nf=$(mktemp)\n"
	got := MktempFindings(src)
	if len(got) != 2 || got[0].Line != 3 || got[1].Line != 5 {
		t.Fatalf("findings %+v, want lines 3 and 5", got)
	}
}

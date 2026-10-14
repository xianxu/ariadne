package shelllint

import (
	"strings"
	"testing"
)

func TestArrayFindings(t *testing.T) {
	const strict = "#!/usr/bin/env bash\nset -euo pipefail\n"
	cases := []struct {
		name, line string
		want       int // findings on that line
	}{
		// Unsafe under set -u on bash 3.2 when the array is empty (#353).
		{"quoted @", `for x in "${xs[@]}"; do :; done`, 1},
		{"unquoted @", `printf '%s\n' ${xs[@]}`, 1},
		{"star", `echo "${names[*]}"`, 1},
		{"assignment", `SEGMENTS=("${TRIMMED[@]}")`, 1},
		{"case operator", `echo "${xs[@]^}"`, 1},
		{"two on a line", `cp "${a[@]}" "${b[@]}"`, 2},
		{"inside a substitution", `n=$(wc -l "${files[@]}")`, 1},
		// Safe.
		{"guarded", `for s in ${seen[@]+"${seen[@]}"}; do :; done`, 0},
		{"guarded in quotes", `for s in "${seen[@]+"${seen[@]}"}"; do :; done`, 0},
		{"colon guard", `x=${xs[@]:+"${xs[@]}"}`, 0},
		{"guarded substring", `queue=(${queue[@]+"${queue[@]:1}"})`, 0},
		{"count", `[ ${#xs[@]} -gt 0 ]`, 0},
		{"default", `echo "${xs[@]:-none}"`, 0},
		// Measured safe on bash 3.2 for an empty array.
		{"substring", `queue=("${queue[@]:1}")`, 0},
		{"pattern removal", `echo "${xs[@]#prefix}" "${xs[@]%.sh}"`, 0},
		{"replacement", `echo "${xs[@]/a/b}"`, 0},
		{"positional", `exec "$@"; echo "$*"`, 0},
		{"scalar", `echo "${name} ${name:-x}"`, 0},
		{"single-quoted text", `echo 'use "${xs[@]}" here'`, 0},
		{"comment", `# "${arr[@]}" is unsafe`, 0},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got := ArrayFindings(strict + c.line + "\n")
			n := 0
			for _, f := range got {
				if f.Line != 3 || !strings.Contains(f.Reason, "set -u") {
					t.Fatalf("unexpected finding %+v", f)
				}
				n++
			}
			if n != c.want {
				t.Fatalf("%d findings %+v, want %d", n, got, c.want)
			}
		})
	}
}

// Only a script that enables nounset is checked, in any of its spellings.
func TestArrayFindingsNeedsNounset(t *testing.T) {
	line := `for x in "${xs[@]}"; do :; done`
	for _, header := range []string{"set -u", "set -eu", "set -o nounset", "  set -euo pipefail  # strict", "#!/bin/bash -eu", "#!/usr/bin/env -S bash -u"} {
		if got := ArrayFindings(header + "\n" + line + "\n"); len(got) != 1 {
			t.Errorf("%q: findings %+v, want 1", header, got)
		}
	}
	for _, header := range []string{"set -e", "set -o pipefail", "# set -u", "echo 'set -u'", "#!/bin/bash -e", "#!/usr/bin/env bash"} {
		if got := ArrayFindings(header + "\n" + line + "\n"); len(got) != 0 {
			t.Errorf("%q: findings %+v, want none", header, got)
		}
	}
}

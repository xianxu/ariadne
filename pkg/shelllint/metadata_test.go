package shelllint

import (
	"strings"
	"testing"
)

func TestMetadataPipeFindings(t *testing.T) {
	for _, tc := range []struct {
		name, src string
		bad       bool
	}{
		{"revision", `printf '%s\n' "$metadata" | grep -Eq "build[[:space:]]+vcs.revision=$PIN_REF$"`, true},
		{"modified", `printf '%s\n' "$metadata" | grep -Eq 'build[[:space:]]+vcs.modified=false$'`, true},
		{"split option", `printf '%s\n' "$metadata" | grep -E -q 'vcs.modified=false$'`, true},
		{"long quiet", `cat data | grep --quiet 'vcs.revision=x'`, true},
		{"long silent", `cat data | grep --silent 'vcs.revision=x'`, true},
		{"escaped field", `cat data | grep -q 'vcs\.revision=x'`, true},
		{"full input", `printf '%s\n' "$metadata" | grep -E 'vcs.modified=false$' >/dev/null`, false},
		{"file input", `grep -Eq 'vcs.modified=false$' metadata.txt`, false},
		{"here string", `grep -Eq 'vcs.modified=false$' <<< "$metadata"`, false},
		{"or not pipe", `false || grep -q 'vcs.modified=false$' data`, false},
		{"quoted example", `echo 'cat data | grep -q vcs.modified=false'`, false},
		{"comment", `# cat data | grep -q vcs.modified=false`, false},
		{"comment field", `cat data | grep -q other # vcs.modified`, false},
		{"unrelated pipe", `cat data | grep -q other`, false},
		{"pattern option", `cat data | grep -e 'vcs.revision=-q'`, false},
		{"attached pattern", `cat data | grep -evcs.revision=-q`, false},
		{"quiet after pattern option", `cat data | grep -e 'vcs.revision=x' -q`, true},
		{"quiet after long pattern", `cat data | grep --regexp 'vcs.revision=x' --quiet`, true},
		{"quiet after count value", `cat data | grep -m 1 -q 'vcs.revision=x'`, true},
		{"option terminator", `cat data | grep -- 'vcs.revision=-q'`, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			got := MetadataPipeFindings(tc.src)
			if !tc.bad {
				if len(got) != 0 {
					t.Fatalf("unexpected findings: %+v", got)
				}
				return
			}
			if len(got) != 1 || got[0].Line != 1 || !strings.Contains(got[0].Reason, "#534") {
				t.Fatalf("want one incident finding on line 1, got %+v", got)
			}
		})
	}
}

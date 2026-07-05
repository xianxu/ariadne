package helptext

import (
	"regexp"
	"strings"
	"testing"

	"github.com/xianxu/ariadne/pkg/vocab"
)

// docSectionRE pulls "Name" out of the help doc's "Body sections" list lines,
// which look like `    ## Done when    acceptance criteria ...` — a `##` heading,
// the section name, then a 2+-space gutter before the description. The name may
// contain single spaces ("Done when", "Side quests"); the lazy group stops at the
// first 2+-space run. Line-anchored (?m), so the mid-line `## Plan` inside a
// description is never captured.
var docSectionRE = regexp.MustCompile(`(?m)^ *## (.+?)  +\S`)

// TestIssueHelpDocumentsEveryScaffoldSection enforces the invariant chain's right
// edge: `sdlc issue --help` documents a SUPERSET of the creation template, so
// every modeled scaffold.sections entry must appear in the help doc (the doc may
// list more — e.g. Estimate, Side quests). Catches "added a section to issue.cue
// but forgot to document it for humans."
func TestIssueHelpDocumentsEveryScaffoldSection(t *testing.T) {
	documented := map[string]bool{}
	for _, m := range docSectionRE.FindAllStringSubmatch(MustGet("issue"), -1) {
		documented[strings.TrimSpace(m[1])] = true
	}
	for _, s := range vocab.Issue().Sections() {
		if !documented[s.Name] {
			t.Errorf("issue.md `sdlc issue --help` omits section %q modeled in issue.cue "+
				"scaffold.sections; document it in the Body-sections list (documented=%v)", s.Name, documented)
		}
	}
}

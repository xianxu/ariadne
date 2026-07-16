package project

import (
	"fmt"
	"strings"

	"github.com/xianxu/ariadne/pkg/vocab"
)

// ScaffoldSpec is the pure input to a new project record.
type ScaffoldSpec struct {
	Name, Goal, DoneWhen, Today string
}

// RenderScaffold renders frontmatter plus the model-owned ordered sections.
func RenderScaffold(s ScaffoldSpec) string {
	m := vocab.Project()
	var b strings.Builder
	b.WriteString("---\n")
	fmt.Fprintf(&b, "type: project\nname: %s\ngoal: %s\ndone_when: %s\n", s.Name, s.Goal, s.DoneWhen)
	fmt.Fprintf(&b, "status: %s\ncreated: %s\nupdated: %s\n---\n\n# %s\n", m.InitialStatus(), s.Today, s.Today, s.Name)
	for _, section := range m.Sections() {
		fmt.Fprintf(&b, "\n## %s\n", section.Name)
		if section.Seed != "" {
			fmt.Fprintf(&b, "\n%s\n", section.Seed)
		}
	}
	return b.String()
}

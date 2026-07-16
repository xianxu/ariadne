package project

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/xianxu/ariadne/cmd/sdlc/internal/issue"
)

var (
	taskRowRE  = regexp.MustCompile(`^- \[([ x.\-~])\] (.*)$`)
	refGroupRE = regexp.MustCompile(`\[([^\]]+)\]`)
)

// Task is one checkbox row in a project's Breakdown section. LineIdx is the
// row's zero-based index in the markdown body (after frontmatter).
type Task struct {
	LineIdx int
	State   byte
	Title   string
	RefText string
}

type sectionSpan struct {
	start int
	end   int
}

// Doc is a parsed project file. The body lines remain the render source of
// truth so parsing and non-mutating reads never reflow the document.
type Doc struct {
	fm       string
	lines    []string
	Tasks    []Task
	sections map[string]sectionSpan
}

// ParseDoc parses frontmatter, section spans, and task rows without changing
// the source bytes.
func ParseDoc(text string) (*Doc, error) {
	fm, body, err := issue.Parse(text)
	if err != nil {
		return nil, err
	}
	return parseDocBody(fm, body), nil
}

func parseDocBody(fm, body string) *Doc {
	d := &Doc{
		fm:       fm,
		lines:    strings.Split(body, "\n"),
		sections: make(map[string]sectionSpan),
	}

	var current string
	for i, line := range d.lines {
		if strings.HasPrefix(line, "## ") {
			if current != "" {
				span := d.sections[current]
				span.end = i
				d.sections[current] = span
			}
			current = strings.TrimSpace(strings.TrimPrefix(line, "## "))
			d.sections[current] = sectionSpan{start: i + 1, end: len(d.lines)}
			continue
		}

		match := taskRowRE.FindStringSubmatch(line)
		if match == nil {
			continue
		}
		remainder := match[2]
		title := strings.TrimSpace(remainder)
		refText := ""
		refs := refGroupRE.FindAllStringSubmatchIndex(remainder, -1)
		if len(refs) > 0 {
			last := refs[len(refs)-1]
			title = strings.TrimSpace(remainder[:last[0]])
			refText = remainder[last[2]:last[3]]
		}
		d.Tasks = append(d.Tasks, Task{
			LineIdx: i,
			State:   match[1][0],
			Title:   title,
			RefText: refText,
		})
	}

	return d
}

// FM returns a trimmed frontmatter field value, or an empty string when absent.
func (d *Doc) FM(field string) string {
	value, _ := issue.GetField(d.fm, field)
	return value
}

// SetFM updates one frontmatter field through the shared issue helper.
func (d *Doc) SetFM(field, value string) {
	d.fm = issue.SetField(d.fm, field, value)
}

// SectionBody returns the source body under a level-two section heading.
func (d *Doc) SectionBody(name string) string {
	span, ok := d.sections[name]
	if !ok {
		return ""
	}
	return strings.Join(d.lines[span.start:span.end], "\n")
}

// AppendToSection adds a block after the section's existing content, separated
// by one blank line, then rebuilds spans and task indexes from the new body.
func (d *Doc) AppendToSection(name, block string) error {
	span, ok := d.sections[name]
	if !ok {
		return fmt.Errorf("section %q not found", name)
	}
	block = strings.Trim(block, "\n")
	if block == "" {
		return nil
	}

	insertAt := span.end
	for insertAt > span.start && d.lines[insertAt-1] == "" {
		insertAt--
	}
	lines := append([]string(nil), d.lines[:insertAt]...)
	if insertAt > span.start {
		lines = append(lines, "")
	}
	lines = append(lines, strings.Split(block, "\n")...)
	lines = append(lines, "")
	lines = append(lines, d.lines[span.end:]...)

	reparsed, err := ParseDoc(issue.Compose(d.fm, strings.Join(lines, "\n")))
	if err != nil {
		return err
	}
	*d = *reparsed
	return nil
}

// SetTaskState rewrites only the selected task row's checkbox state.
func (d *Doc) SetTaskState(i int, state byte) {
	if i < 0 || i >= len(d.Tasks) {
		return
	}
	task := &d.Tasks[i]
	line := d.lines[task.LineIdx]
	if len(line) < 4 {
		return
	}
	d.lines[task.LineIdx] = line[:3] + string(state) + line[4:]
	task.State = state
}

// Render reassembles the current frontmatter and line-preserved markdown body.
func (d *Doc) Render() string {
	return issue.Compose(d.fm, strings.Join(d.lines, "\n"))
}

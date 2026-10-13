package project

import (
	"fmt"
	"strings"
)

// Summary is the pure read model shared by project list/show.
type Summary struct {
	Path, Type, Name, Goal, DoneWhen, Status string
	TL, Created, Updated                     string
	Done, Total                              int
}

func Summarize(path string, d *Doc) Summary {
	s := Summary{Path: path, Type: d.FM("type"), Name: d.FM("name"), Goal: d.FM("goal"), DoneWhen: d.FM("done_when"), Status: d.FM("status"), Created: d.FM("created"), Updated: d.FM("updated"), Total: len(d.Tasks)}
	// An unreadable tl renders as absent here; validate and close report it.
	if owners, err := d.TLOwners(); err == nil {
		s.TL = strings.Join(owners, ", ")
	}
	for _, task := range d.Tasks {
		if task.State == 'x' {
			s.Done++
		}
	}
	return s
}

func RenderListRow(s Summary) string {
	tl := s.TL
	if tl == "" {
		tl = "-"
	}
	return fmt.Sprintf("%s  %s  %s\n", s.Name, s.Status, tl)
}

func RenderShow(s Summary) string {
	var b strings.Builder
	fmt.Fprintf(&b, "%s\n---\n", s.Path)
	fields := []struct{ name, value string }{{"type", s.Type}, {"name", s.Name}, {"goal", s.Goal}, {"done_when", s.DoneWhen}, {"status", s.Status}, {"tl", s.TL}, {"created", s.Created}, {"updated", s.Updated}}
	for _, f := range fields {
		if f.value != "" {
			fmt.Fprintf(&b, "%s: %s\n", f.name, f.value)
		}
	}
	fmt.Fprintf(&b, "---\ntasks: %d/%d done\n", s.Done, s.Total)
	return b.String()
}

package project

import (
	"fmt"
	"strings"
)

// Summary is the pure read model shared by project list/show.
type Summary struct {
	Path, Type, Name, Goal, DoneWhen, Status  string
	Deadline, PlannedFinish, Created, Updated string
	Done, Total                               int
}

func Summarize(path string, d *Doc) Summary {
	s := Summary{Path: path, Type: d.FM("type"), Name: d.FM("name"), Goal: d.FM("goal"), DoneWhen: d.FM("done_when"), Status: d.FM("status"), Deadline: d.FM("deadline"), PlannedFinish: d.FM("planned_finish"), Created: d.FM("created"), Updated: d.FM("updated"), Total: len(d.Tasks)}
	for _, task := range d.Tasks {
		if task.State == 'x' {
			s.Done++
		}
	}
	return s
}

func RenderListRow(s Summary) string {
	deadline := s.Deadline
	if deadline == "" {
		deadline = "-"
	}
	return fmt.Sprintf("%s  %s  %s\n", s.Name, s.Status, deadline)
}

func RenderShow(s Summary) string {
	var b strings.Builder
	fmt.Fprintf(&b, "%s\n---\n", s.Path)
	fields := []struct{ name, value string }{{"type", s.Type}, {"name", s.Name}, {"goal", s.Goal}, {"done_when", s.DoneWhen}, {"status", s.Status}, {"deadline", s.Deadline}, {"planned_finish", s.PlannedFinish}, {"created", s.Created}, {"updated", s.Updated}}
	for _, f := range fields {
		if f.value != "" {
			fmt.Fprintf(&b, "%s: %s\n", f.name, f.value)
		}
	}
	fmt.Fprintf(&b, "---\ntasks: %d/%d done\n", s.Done, s.Total)
	return b.String()
}

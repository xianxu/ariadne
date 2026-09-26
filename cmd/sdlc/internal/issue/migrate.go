// migrate.go — card derivation for the one-time tracker migration (#252).
//
// Active details become card + mirrored details through the same SplitCard the
// rest of the system uses; the only normalization is giving a pre-template file
// its `## Problem` heading. Archived details are never rewritten — their own
// frontmatter stays their terminal authority — so their cards are built
// tolerantly, and every value the builder had to infer is reported rather than
// silently invented.
package issue

import (
	"errors"
	"fmt"
	"slices"
	"strings"

	"github.com/xianxu/ariadne/pkg/vocab"
	"go.yaml.in/yaml/v3"
)

// MigrateActiveDetails splits legacy active details into a card and mirrored
// details. Details without a `## Problem` section get the heading inserted above
// their preamble (the text between the H1 and the next section), or — with no
// preamble — a labelled placeholder section; the inferences list says which.
// Any invalid owned field refuses.
func MigrateActiveDetails(details []byte, objectFormat string) (card, mirrored []byte, inferences []string, err error) {
	normalized, inferred, err := withProblemHeading(details)
	if err != nil {
		return nil, nil, nil, err
	}
	card, mirrored, err = SplitCardWithFormat(normalized, objectFormat)
	if err != nil {
		return nil, nil, nil, err
	}
	if inferred != "" {
		inferences = append(inferences, inferred)
	}
	return card, mirrored, inferences, nil
}

// NoProblemPlaceholder is the Problem a migrated issue gets when its details
// never stated one.
const NoProblemPlaceholder = "No problem statement was recorded before the issue tracker migration."

// withProblemHeading returns details unchanged when they have a Problem
// section; else with `## Problem` inserted before the preamble's first line, or
// a placeholder Problem section under the title when there is no preamble. The
// string names the inference ("" for none).
func withProblemHeading(details []byte) ([]byte, string, error) {
	fm, body, err := Parse(string(details))
	if err != nil {
		return nil, "", invalidCard("frontmatter", err.Error())
	}
	lines := strings.Split(body, "\n")
	fenced := FenceSpans(lines, UnterminatedIsProse)
	h1, start := -1, -1
	for i, line := range lines {
		if fenced[i] {
			continue
		}
		if strings.TrimSpace(line) == "## Problem" {
			return details, "", nil
		}
		if h1 < 0 && strings.HasPrefix(line, "# ") {
			h1 = i
			continue
		}
		if h1 >= 0 && start < 0 && strings.HasPrefix(line, "## ") {
			break // the preamble ended without content
		}
		if h1 >= 0 && start < 0 && strings.TrimSpace(line) != "" {
			start = i
		}
	}
	if h1 < 0 {
		return nil, "", invalidCard("title", "expected exactly one nonempty Markdown H1")
	}
	inference := "inserted `## Problem` above the preamble"
	insert := []string{"## Problem", ""}
	if start < 0 {
		start, inference = h1+1, "no problem statement: inserted a placeholder `## Problem`"
		insert = []string{"", "## Problem", "", NoProblemPlaceholder}
	}
	next := append(append(append([]string{}, lines[:start]...), insert...), lines[start:]...)
	return []byte(Compose(fm, strings.Join(next, "\n"))), inference, nil
}

// ArchivedCard builds the card for an archived details file named file
// (NNNNNN-slug.md). Valid owned fields are copied as written; each inference —
// a dropped invalid field, N/A hours for closed work, an inferred status,
// title or Problem — is reported.
func ArchivedCard(details []byte, file, objectFormat string) (card []byte, inferences []string, err error) {
	id, slug, ok := ParseFilename(file)
	if !ok {
		return nil, nil, fmt.Errorf("archived %s: not an issue filename", file)
	}
	infer := func(format string, args ...any) { inferences = append(inferences, fmt.Sprintf(format, args...)) }
	fm, body, perr := Parse(string(details))
	if perr != nil {
		fm, body = "", string(details)
		infer("no frontmatter")
	}
	fields := map[string]*yaml.Node{}
	if fm != "" {
		var doc yaml.Node
		if yerr := yaml.Unmarshal([]byte(fm), &doc); yerr != nil || len(doc.Content) != 1 || doc.Content[0].Kind != yaml.MappingNode {
			infer("unreadable frontmatter")
		} else {
			m := doc.Content[0]
			for i := 0; i+1 < len(m.Content); i += 2 {
				if _, dup := fields[m.Content[i].Value]; dup {
					return nil, nil, fmt.Errorf("archived %s: duplicate frontmatter key %q", file, m.Content[i].Value)
				}
				fields[m.Content[i].Value] = m.Content[i+1]
			}
		}
	}
	out := &yaml.Node{Kind: yaml.MappingNode}
	add := func(name string, value *yaml.Node) {
		out.Content = append(out.Content, &yaml.Node{Kind: yaml.ScalarNode, Tag: "!!str", Value: name}, value)
	}
	str := func(v string) *yaml.Node {
		return &yaml.Node{Kind: yaml.ScalarNode, Tag: "!!str", Value: v}
	}
	status := ""
	for _, field := range vocab.Issue().CardFields() {
		switch field.Kind {
		case "title":
			continue
		case "id":
			if n, ok := fields["id"]; ok && (validateCardScalar(field, n) != nil || n.Value != id) {
				infer("id %q replaced by the filename's %s", n.Value, id)
			}
			quoted := str(id)
			quoted.Style = yaml.SingleQuotedStyle // six digits, never an (octal) integer
			add("id", quoted)
			continue
		}
		n, ok := fields[field.Name]
		if field.Kind == "status" {
			if ok && validateCardScalar(field, n) == nil {
				status = n.Value
			} else {
				status = "done"
				infer("status inferred done (archived)")
			}
			add("status", str(status))
			continue
		}
		if !ok || (n.Kind == yaml.ScalarNode && n.Tag == "!!null") {
			if field.Kind == "actual" && (status == "done" || status == "codecomplete") {
				add(field.Name, str(ActualNotApplicableSentinel))
				infer("%s: N/A (closed without measured hours)", field.Name)
			}
			continue
		}
		if validateCardScalar(field, n) != nil {
			if field.Kind == "actual" && (status == "done" || status == "codecomplete") {
				add(field.Name, str(ActualNotApplicableSentinel))
				infer("%s %q unreadable: N/A", field.Name, n.Value)
			} else {
				infer("%s %q dropped (invalid)", field.Name, n.Value)
			}
			continue
		}
		add(field.Name, n)
	}
	title, problem := archivedTitleAndProblem(body, slug, file, infer)
	var buf strings.Builder
	enc := yaml.NewEncoder(&buf)
	enc.SetIndent(2)
	if err := enc.Encode(out); err != nil {
		return nil, nil, err
	}
	card = []byte(Compose(strings.TrimSuffix(buf.String(), "\n"), "\n# "+title+"\n\n## Problem\n"+problem))
	if _, err := ParseCard(card); err != nil {
		return nil, nil, fmt.Errorf("archived %s: %w", file, err)
	}
	if _, err := CardBlobOID(card, objectFormat); err != nil {
		return nil, nil, err
	}
	return card, inferences, nil
}

// archivedTitleAndProblem reads the H1 and the Problem (else the preamble)
// from an archived body, falling back to the slug and a pointer to the file.
func archivedTitleAndProblem(body, slug, file string, infer func(string, ...any)) (string, string) {
	lines := strings.Split(body, "\n")
	fenced := FenceSpans(lines, UnterminatedIsProse)
	var titles []string
	problems, h1 := 0, -1
	var preamble []string
	inPreamble := false
	for i, line := range lines {
		if fenced[i] {
			if inPreamble {
				preamble = append(preamble, line)
			}
			continue
		}
		switch {
		case strings.HasPrefix(line, "# ") && strings.TrimSpace(line[2:]) != "":
			titles = append(titles, strings.TrimSpace(line[2:]))
			if h1 < 0 {
				h1, inPreamble = i, true
				continue
			}
			inPreamble = false
		case strings.HasPrefix(line, "## "):
			if strings.TrimSpace(line) == "## Problem" {
				problems++
			}
			inPreamble = false
		}
		if inPreamble {
			preamble = append(preamble, line)
		}
	}
	title := ""
	switch len(titles) {
	case 1:
		title = titles[0]
	case 0:
		title = strings.ReplaceAll(slug, "-", " ")
		infer("title from the slug")
	default:
		title = titles[0]
		infer("title from the first of %d H1s", len(titles))
	}
	if problems == 1 {
		if p, ok := SectionBody(body, "Problem"); ok && strings.TrimSpace(p) != "" {
			return title, p
		}
	}
	if p := strings.TrimSpace(strings.Join(preamble, "\n")); p != "" {
		infer("Problem from the preamble")
		return title, "\n" + p + "\n"
	}
	infer("Problem points at the archived file")
	return title, "\nArchived before the issue tracker; the archived details file " + file + " is the record.\n"
}

// ErrNotLegacy marks details that already carry a card mirror.
var ErrNotLegacy = errors.New("details already carry a card mirror")

// ReconcileLegacyDetails attaches the imported card's mirror to a pre-cutover
// branch's legacy details, proving first that every card-owned field the branch
// holds equals the imported card's (Problem is excluded: details may revise it).
func ReconcileLegacyDetails(details, importedCard []byte, objectFormat string) ([]byte, error) {
	if HasMirror(details) {
		return nil, ErrNotLegacy
	}
	normalized, _, err := withProblemHeading(details)
	if err != nil {
		return nil, err
	}
	d, err := parseCardDocument(normalized)
	if err != nil {
		return nil, err
	}
	c, err := parseCardDocument(importedCard)
	if err != nil {
		return nil, err
	}
	var differ []string
	for _, field := range vocab.Issue().CardFields() {
		if field.Kind == "title" {
			if d.title != c.title {
				differ = append(differ, "title")
			}
			continue
		}
		if !sameCardValue(d.fields[field.Name], c.fields[field.Name]) {
			differ = append(differ, field.Name)
		}
	}
	if len(differ) > 0 {
		slices.Sort(differ)
		return nil, fmt.Errorf("card-owned fields differ from the imported card: %s", strings.Join(differ, ", "))
	}
	oid, err := CardBlobOID(importedCard, objectFormat)
	if err != nil {
		return nil, err
	}
	return []byte(Compose(d.fm+"\n"+mirrorLine(oid), d.body)), nil
}

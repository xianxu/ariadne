package main

import (
	"flag"
	"fmt"
	"os"
	"regexp"
	"strings"

	"github.com/xianxu/ariadne/pkg/frontmatter"
)

// validate.go — `vocabulary validate-instance --type <noun> <file>`: the
// instance-conformance validator (#124). It extracts a typed-markdown file's
// frontmatter and `cue vet`s it against the noun's schema definition (#<Noun>),
// rendering cue's verbose output as clear, per-field diagnostics. Generic over the
// noun — the only per-datatype input is the construct/vocabulary/<noun>.cue the
// DAG-merge resolves. Well-formedness only (frontmatter shape); the LLM owns
// semantic quality. No write-back: it reports; the LLM/human fixes + re-validates.

// Diagnostic is one clear, per-field conformance problem — the actionable form the
// Done-when requires ("actionable enough that an LLM fixes the file and
// re-validates"). Field is "" when cue doesn't name a field for the failure.
type Diagnostic struct {
	Field   string
	Message string
}

func (d Diagnostic) String() string {
	if d.Field != "" {
		return d.Field + ": " + d.Message
	}
	return d.Message
}

// runValidateInstance is the CLI entry point.
func runValidateInstance(args []string) error {
	fs := flag.NewFlagSet("validate-instance", flag.ExitOnError)
	noun := fs.String("type", "", "datatype noun (e.g. issue) — selects #<Noun> in its .cue")
	_ = fs.Parse(args)
	if *noun == "" {
		return fmt.Errorf("validate-instance: --type <noun> is required")
	}
	rest := fs.Args()
	if len(rest) != 1 {
		return fmt.Errorf("validate-instance: want exactly one <file> argument, got %d", len(rest))
	}
	file := rest[0]

	paths, err := resolveVocab()
	if err != nil {
		return err
	}
	schema, ok := paths[*noun]
	if !ok {
		return fmt.Errorf("no vocabulary noun %q (have: %v)", *noun, sortedKeys(paths))
	}

	diags, err := validateInstanceFile(osCue{}, file, schema, *noun)
	if err != nil {
		return err
	}
	if len(diags) > 0 {
		fmt.Fprintf(os.Stderr, "%s does not conform to #%s:\n", file, titleCase(*noun))
		for _, d := range diags {
			fmt.Fprintf(os.Stderr, "  - %s\n", d.String())
		}
		return fmt.Errorf("%d conformance error(s) in %s", len(diags), file)
	}
	return nil
}

// validateInstanceFile reads a markdown file, extracts its frontmatter to a `.yaml`
// temp (cue infers YAML by extension), vets it against #<Noun> in schemaPath, and
// returns the per-field diagnostics (empty = conformant). The CueRunner is injected
// so the unit tests run without the `cue` binary (ARCH-PURE: the IO seam is the
// runner; the transform below is pure).
func validateInstanceFile(cue CueRunner, file, schemaPath, noun string) ([]Diagnostic, error) {
	raw, err := os.ReadFile(file)
	if err != nil {
		return nil, fmt.Errorf("read %s: %w", file, err)
	}
	fm, _, err := frontmatter.Split(string(raw))
	if err != nil {
		return nil, fmt.Errorf("%s: %w", file, err)
	}
	tmp, err := os.CreateTemp("", "conform-*.yaml") // .yaml so cue parses it as YAML
	if err != nil {
		return nil, err
	}
	defer os.Remove(tmp.Name())
	if _, err := tmp.WriteString(fm); err != nil {
		tmp.Close()
		return nil, err
	}
	tmp.Close()

	out, err := cue.VetInstance(tmp.Name(), schemaPath, "#"+titleCase(noun))
	if err != nil {
		return nil, err
	}
	if strings.TrimSpace(out) == "" {
		return nil, nil // clean
	}
	return parseCueDiagnostics(out), nil
}

var (
	// `status: conflicting values "working" and "in-progress":`
	cueConflictRE = regexp.MustCompile(`^(\S+): conflicting values "([^"]*)" and "([^"]*)":`)
	// `status: 6 errors in empty disjunction:` — header above the conflict lines.
	cueEmptyDisjRE = regexp.MustCompile(`^(\S+): \d+ errors in empty disjunction:`)
	// `actual_hours: field is required but not present`
	cueRequiredRE = regexp.MustCompile(`^(\S+): field is required but not present`)
	// `target: field not allowed:` — only a closed schema emits this (#Issue is open).
	cueNotAllowedRE = regexp.MustCompile(`^(\S+): field not allowed:`)
	// `unresolved disjunction "open" | "working" | … (type string):` — an OPEN struct's
	// way of reporting a missing required enum field (cue names no field here).
	cueUnresolvedDisjRE = regexp.MustCompile(`^unresolved disjunction (.+) \(type [^)]+\):`)
)

// parseCueDiagnostics maps `cue vet -d` combined output → a deduped, ordered list
// of clear per-field diagnostics. Pure (string → []Diagnostic).
//
// FIXTURE-COUPLED: the recognized shapes are captured from cue v0.16.1 (see
// validate_test.go). A cue upgrade may reword them — if a bump breaks the parse,
// re-capture the fixtures. Unrecognized message lines pass through verbatim so a
// new shape is surfaced, never silently dropped.
func parseCueDiagnostics(out string) []Diagnostic {
	var diags []Diagnostic
	seen := map[string]bool{}
	add := func(d Diagnostic) {
		k := d.Field + "\x00" + d.Message
		if seen[k] {
			return
		}
		seen[k] = true
		diags = append(diags, d)
	}

	// Enum conflicts repeat one line per enum member; gather the want-set per field
	// and emit one diagnostic. Track first-seen order for determinism.
	enumActual := map[string]string{}
	enumWants := map[string][]string{}
	var enumOrder []string

	for _, line := range strings.Split(out, "\n") {
		line = strings.TrimRight(line, "\r")
		if line == "" || line[0] == ' ' || line[0] == '\t' {
			continue // blank, or an indented location line
		}
		if m := cueConflictRE.FindStringSubmatch(line); m != nil {
			field, want, actual := m[1], m[2], m[3]
			if _, ok := enumActual[field]; !ok {
				enumOrder = append(enumOrder, field)
			}
			enumActual[field] = actual
			enumWants[field] = appendUnique(enumWants[field], want)
			continue
		}
		if cueEmptyDisjRE.MatchString(line) {
			continue // header; the conflict lines carry the detail
		}
		if m := cueRequiredRE.FindStringSubmatch(line); m != nil {
			add(Diagnostic{Field: m[1], Message: "required field is missing"})
			continue
		}
		if m := cueNotAllowedRE.FindStringSubmatch(line); m != nil {
			add(Diagnostic{Field: m[1], Message: "unknown field (not allowed by the schema)"})
			continue
		}
		if m := cueUnresolvedDisjRE.FindStringSubmatch(line); m != nil {
			add(Diagnostic{Message: "a required field is missing or has an invalid value (want one of: " + normalizeDisjunction(m[1]) + ")"})
			continue
		}
		add(Diagnostic{Message: line}) // unrecognized — surface it, don't drop it
	}
	for _, field := range enumOrder {
		add(Diagnostic{Field: field, Message: `"` + enumActual[field] + `" is not valid (want: ` + strings.Join(enumWants[field], "|") + ")"})
	}
	return diags
}

// normalizeDisjunction turns `"open" | "working" | "blocked"` into
// `open|working|blocked` for a compact diagnostic.
func normalizeDisjunction(s string) string {
	parts := strings.Split(s, "|")
	for i, p := range parts {
		parts[i] = strings.Trim(strings.TrimSpace(p), `"`)
	}
	return strings.Join(parts, "|")
}

func appendUnique(xs []string, x string) []string {
	for _, e := range xs {
		if e == x {
			return xs
		}
	}
	return append(xs, x)
}

// titleCase upper-cases the first byte of a single-word noun (issue → Issue,
// pensive → Pensive) to form its #Definition name.
func titleCase(s string) string {
	if s == "" {
		return s
	}
	return strings.ToUpper(s[:1]) + s[1:]
}

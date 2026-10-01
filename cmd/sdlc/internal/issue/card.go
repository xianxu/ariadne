package issue

import (
	"fmt"
	"io"
	"math"
	"reflect"
	"regexp"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/xianxu/ariadne/pkg/vocab"
	"go.yaml.in/yaml/v3"
)

// Card is a validated read-only snapshot. Changing its exported strings does
// not update its private projection; callers edit serialized bytes and reparse.
// Frontmatter retains the source spelling, including legacy unquoted IDs.
type Card struct {
	ID, Title, Problem, Frontmatter string
	doc                             *cardDocument
}

type cardSpan struct{ start, end int }
type cardDocument struct {
	fm, body, title, problem string
	fields                   map[string]*yaml.Node
	spans                    map[string]cardSpan
	titleSpan                cardSpan
}

// ValidationError identifies malformed card/detail input before any effects.
type ValidationError struct{ Field, Reason string }

func (e *ValidationError) Error() string {
	return fmt.Sprintf("invalid issue %s: %s", e.Field, e.Reason)
}

func invalidCard(field, reason string) error {
	for _, owned := range vocab.Issue().CardFields() {
		if owned.Name == field {
			reason += "; edit via `" + owned.Setter + "`"
			break
		}
	}
	return &ValidationError{field, reason}
}

// ParseCard validates a serialized tracker card. Unknown user fields are
// refused; structured transaction metadata belongs in the versioned envelope.
func ParseCard(raw []byte) (Card, error) {
	d, err := parseCardDocument(raw)
	if err != nil {
		return Card{}, err
	}
	var unexpectedSection string
	ScanMarkdownLines(strings.Split(d.body, "\n"), UnterminatedIsProse, func(_ int, line string) {
		if strings.HasPrefix(line, "## ") && strings.TrimSpace(line) != "## Problem" {
			unexpectedSection = line
		}
	})
	if unexpectedSection != "" {
		return Card{}, invalidCard("body", "card contains detail section "+unexpectedSection)
	}
	allowed := map[string]bool{}
	for _, f := range vocab.Issue().CardFields() {
		if f.Kind != "title" {
			allowed[f.Name] = true
		}
	}
	internal := vocab.Issue().Card.Internal
	for name, node := range d.fields {
		if allowed[name] {
			continue
		}
		if name != internal.Field {
			return Card{}, invalidCard(name, "not a card field")
		}
		if node.Kind != yaml.MappingNode {
			return Card{}, invalidCard(name, "expected a versioned transaction mapping")
		}
		version := ""
		for i := 0; i < len(node.Content); i += 2 {
			if node.Content[i].Value == "version" && node.Content[i+1].Tag == "!!int" {
				version = node.Content[i+1].Value
			}
		}
		if version != strconv.Itoa(internal.Version) {
			return Card{}, invalidCard(name, "unsupported or missing transaction version")
		}
	}
	return Card{ID: d.fields["id"].Value, Title: d.title, Problem: d.problem, Frontmatter: d.fm, doc: d}, nil
}

// SplitCard extracts authoritative metadata and the original Problem, then
// initializes the detail mirror. SHA-1 is the default Git object format;
// repositories using SHA-256 must call SplitCardWithFormat explicitly.
func SplitCard(details []byte) (card, mirroredDetails []byte, err error) {
	return SplitCardWithFormat(details, "sha1")
}

func SplitCardWithFormat(details []byte, objectFormat string) (card, mirroredDetails []byte, err error) {
	d, err := parseCardDocument(details)
	if err != nil {
		return nil, nil, err
	}
	if _, exists := d.fields[MirrorField]; exists {
		return nil, nil, invalidCard(MirrorField, "already mirrored; use RefreshMirror")
	}
	var fm strings.Builder
	for _, field := range vocab.Issue().CardFields() {
		if span, ok := d.spans[field.Name]; ok && field.Kind != "title" {
			fm.WriteString(d.fm[span.start:span.end])
			fm.WriteByte('\n')
		}
	}
	card = []byte(Compose(strings.TrimSuffix(fm.String(), "\n"), "\n# "+d.title+"\n\n## Problem\n"+d.problem))
	if _, err := ParseCard(card); err != nil {
		return nil, nil, err
	}
	oid, err := CardBlobOID(card, objectFormat)
	if err != nil {
		return nil, nil, err
	}
	mirroredDetails = []byte(Compose(d.fm+"\n"+mirrorLine(oid), d.body))
	return card, mirroredDetails, nil
}

func parseCardDocument(raw []byte) (*cardDocument, error) {
	fm, body, err := Parse(string(raw))
	if err != nil {
		return nil, invalidCard("frontmatter", err.Error())
	}
	var node yaml.Node
	dec := yaml.NewDecoder(strings.NewReader(fm))
	if err := dec.Decode(&node); err != nil {
		return nil, invalidCard("frontmatter", err.Error())
	}
	var extra yaml.Node
	if err := dec.Decode(&extra); err != io.EOF {
		return nil, invalidCard("frontmatter", "expected exactly one YAML mapping")
	}
	if len(node.Content) != 1 || node.Content[0].Kind != yaml.MappingNode || node.Content[0].Style&yaml.FlowStyle != 0 {
		return nil, invalidCard("frontmatter", "expected a block mapping with one top-level field per line")
	}
	root := node.Content[0]
	if err := validateCardYAML(root); err != nil {
		return nil, err
	}
	d := &cardDocument{fm: fm, body: body, fields: map[string]*yaml.Node{}, spans: map[string]cardSpan{}}
	lines := strings.Split(fm, "\n")
	offsets := make([]int, len(lines)+1)
	for i, line := range lines {
		offsets[i+1] = offsets[i] + len(line) + 1
	}
	for i := 0; i < len(root.Content); i += 2 {
		key, value := root.Content[i], root.Content[i+1]
		if key.Column != 1 {
			return nil, invalidCard(key.Value, "top-level field must start in column one")
		}
		end := len(lines)
		if i+2 < len(root.Content) {
			end = root.Content[i+2].Line - 1
		}
		// Comments and whitespace preceding the next field are not owned by
		// this one. Keep them byte-identical when replacing its scalar.
		for end > key.Line && (strings.TrimSpace(lines[end-1]) == "" || strings.HasPrefix(strings.TrimSpace(lines[end-1]), "#")) {
			end--
		}
		stop := offsets[end] - 1
		if stop > len(fm) {
			stop = len(fm)
		}
		d.fields[key.Value], d.spans[key.Value] = value, cardSpan{offsets[key.Line-1], stop}
	}
	if err := d.parseHeadings(); err != nil {
		return nil, err
	}
	for _, field := range vocab.Issue().CardFields() {
		if field.Kind == "title" {
			continue
		}
		n, ok := d.fields[field.Name]
		if !ok {
			if field.Required {
				return nil, invalidCard(field.Name, "required field is missing")
			}
			continue
		}
		if err := validateCardScalar(field, n); err != nil {
			return nil, err
		}
	}
	status := d.fields["status"].Value
	if status == "done" || status == "codecomplete" {
		if n := d.fields["actual_hours"]; n == nil || n.Tag == "!!null" {
			return nil, invalidCard("actual_hours", "closed work requires measured hours or N/A")
		}
	}
	return d, nil
}

// Duplicate keys anywhere (including unowned mappings) make interpretation
// ambiguous. Aliases/merge keys also introduce hidden ownership, so refuse.
func validateCardYAML(n *yaml.Node) error {
	if n.Kind == yaml.AliasNode || n.Anchor != "" {
		return invalidCard("frontmatter", "YAML aliases and anchors are not supported")
	}
	if n.Kind == yaml.MappingNode {
		seen := map[string]bool{}
		for i := 0; i < len(n.Content); i += 2 {
			key := n.Content[i]
			if key.Kind != yaml.ScalarNode || key.Tag != "!!str" {
				return invalidCard("frontmatter", "mapping keys must be strings, not merges")
			}
			if seen[key.Value] {
				return invalidCard(key.Value, "duplicate YAML key")
			}
			seen[key.Value] = true
		}
	}
	for _, child := range n.Content {
		if err := validateCardYAML(child); err != nil {
			return err
		}
	}
	return nil
}

var cardIDRE = regexp.MustCompile(`^[0-9]{6}$`)

func validateCardScalar(field vocab.CardField, n *yaml.Node) error {
	if field.Kind == "claimant" {
		_, err := parseClaimant(n)
		return err
	}
	fail := func() error { return invalidCard(field.Name, "invalid "+field.Kind+" scalar") }
	if n.Kind != yaml.ScalarNode {
		return fail()
	}
	if n.Tag == "!!null" && !field.Required {
		return nil
	}
	switch field.Kind {
	case "id":
		// Do not decode an unquoted ID as an integer: YAML's octal semantics
		// would turn 000252 into 170. Identity is its six decimal digits.
		if (n.Tag != "!!str" && n.Tag != "!!int" && n.Tag != "!!float") || !cardIDRE.MatchString(n.Value) {
			return fail()
		}
	case "status":
		if n.Tag != "!!str" || !slices.Contains(vocab.Issue().AllStatuses(), n.Value) {
			return fail()
		}
	case "estimate", "actual":
		if field.Kind == "actual" && n.Tag == "!!str" && n.Value == ActualNotApplicableSentinel {
			return nil
		}
		if n.Tag != "!!int" && n.Tag != "!!float" {
			return fail()
		}
		var number float64
		if n.Decode(&number) != nil || math.IsInf(number, 0) || math.IsNaN(number) || number <= 0 {
			return fail()
		}
	case "date", "timestamp":
		if n.Tag != "!!str" && n.Tag != "!!timestamp" {
			return fail()
		}
		layout := time.RFC3339
		if field.Kind == "date" {
			layout = "2006-01-02"
		}
		if _, err := time.Parse(layout, n.Value); err != nil {
			return fail()
		}
	case "github":
		if n.Tag != "!!int" && n.Tag != "!!str" {
			return fail()
		}
		if n.Value == "" && n.Tag == "!!str" {
			return nil
		}
		v, err := strconv.ParseUint(n.Value, 10, 64)
		if err != nil || v == 0 {
			return fail()
		}
	default:
		return invalidCard(field.Name, "unknown field schema kind")
	}
	return nil
}

func (d *cardDocument) parseHeadings() error {
	lines := strings.Split(d.body, "\n")
	spans := FenceSpans(lines, UnterminatedIsProse)
	pos, titles, problems := 0, 0, 0
	for i, line := range lines {
		if !spans[i] && strings.HasPrefix(line, "# ") {
			titles++
			d.title = strings.TrimSpace(strings.TrimPrefix(line, "# "))
			d.titleSpan = cardSpan{pos, pos + len(line)}
		}
		if !spans[i] && strings.TrimSpace(line) == "## Problem" {
			problems++
		}
		pos += len(line) + 1
	}
	if titles != 1 || d.title == "" {
		return invalidCard("title", "expected exactly one nonempty Markdown H1")
	}
	if problems != 1 {
		return invalidCard("Problem", "expected exactly one Problem section")
	}
	d.problem, _ = SectionBody(d.body, "Problem")
	return nil
}

// sameCardValue compares presence, YAML type and decoded scalar value. Numeric
// spelling changes do not change a value, but quoted numbers change its type.
func sameCardValue(a, b *yaml.Node) bool {
	if a == nil || b == nil {
		return a == b
	}
	if a.Tag != b.Tag || a.Kind != b.Kind {
		return false
	}
	if a.Kind == yaml.MappingNode { // a structured field (#277 claimant)
		var av, bv any
		return a.Decode(&av) == nil && b.Decode(&bv) == nil && reflect.DeepEqual(av, bv)
	}
	if a.Value == b.Value {
		return true
	}
	if a.Tag == "!!null" {
		return true
	}
	if a.Tag == "!!int" || a.Tag == "!!float" {
		var av, bv any
		if a.Decode(&av) == nil && b.Decode(&bv) == nil {
			return fmt.Sprint(av) == fmt.Sprint(bv)
		}
	}
	return false
}

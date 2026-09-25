package issue

import (
	"crypto/sha1"
	"crypto/sha256"
	"fmt"
	"regexp"
	"sort"
	"strings"

	"github.com/xianxu/ariadne/pkg/vocab"
	"go.yaml.in/yaml/v3"
)

// MirrorField records the exact blob last projected into this detail file.
const MirrorField = "card_mirror"

// OwnershipError refuses a local edit before a stale mirror can overwrite it.
type OwnershipError struct{ Field, Setter string }

func (e *OwnershipError) Error() string {
	return fmt.Sprintf("%s is owned by the card; use `%s` (restore the mirrored value before refreshing)", e.Field, e.Setter)
}

// CardBlobOID hashes exact bytes using Git's blob header. No repository IO is
// involved. The caller supplies its repository object format, sha1 or sha256.
func CardBlobOID(raw []byte, objectFormat string) (string, error) {
	input := append([]byte(fmt.Sprintf("blob %d\x00", len(raw))), raw...)
	switch objectFormat {
	case "sha1":
		return fmt.Sprintf("%x", sha1.Sum(input)), nil
	case "sha256":
		return fmt.Sprintf("%x", sha256.Sum256(input)), nil
	default:
		return "", invalidCard(MirrorField, "unsupported Git object format "+objectFormat)
	}
}

func mirrorLine(oid string) string {
	return MirrorField + ": '" + oid + "' # card fields mirrored from issue-cards; edit via sdlc"
}

var mirrorOIDRE = regexp.MustCompile(`^(?:[0-9a-f]{40}|[0-9a-f]{64})$`)

// RefreshMirror first proves the supplied baseline is the exact card blob
// named by details, then compares local owned values against that baseline.
// Only an unchanged projection may refresh. Problem and unowned YAML/body bytes
// remain branch-owned. Baseline/current resolution and freshness are the
// caller's responsibility; the pure core verifies content identity itself.
func RefreshMirror(details, baselineCard, currentCard []byte) ([]byte, error) {
	d, err := parseCardDocument(details)
	if err != nil {
		return nil, err
	}
	marker := d.fields[MirrorField]
	if marker == nil || marker.Kind != yaml.ScalarNode || marker.Tag != "!!str" || !mirrorOIDRE.MatchString(marker.Value) {
		return nil, invalidCard(MirrorField, "missing or invalid baseline; reconcile this detail with the tracker")
	}
	format := "sha1"
	if len(marker.Value) == 64 {
		format = "sha256"
	}
	oid, _ := CardBlobOID(baselineCard, format)
	if oid != marker.Value {
		return nil, invalidCard(MirrorField, "baseline bytes do not match recorded Git blob OID; resolve the recorded tracker blob")
	}
	baseline, err := ParseCard(baselineCard)
	if err != nil {
		return nil, err
	}
	current, err := ParseCard(currentCard)
	if err != nil {
		return nil, err
	}
	if baseline.ID != current.ID {
		return nil, invalidCard("id", "current card and baseline identify different issues")
	}
	for _, field := range vocab.Issue().CardFields() {
		equal := sameCardValue(d.fields[field.Name], baseline.doc.fields[field.Name])
		if field.Kind == "title" {
			equal = d.title == baseline.Title
		}
		if !equal {
			return nil, &OwnershipError{field.Name, field.Setter}
		}
	}
	type edit struct {
		cardSpan
		text string
	}
	var edits []edit
	appendFields := ""
	for _, field := range vocab.Issue().CardFields() {
		if field.Kind == "title" {
			continue
		}
		old, oldOK := d.spans[field.Name]
		next, nextOK := current.doc.spans[field.Name]
		if sameCardValue(d.fields[field.Name], current.doc.fields[field.Name]) {
			continue
		}
		text := ""
		if nextOK {
			text = current.Frontmatter[next.start:next.end]
		}
		if oldOK {
			if !nextOK && old.end < len(d.fm) && d.fm[old.end] == '\n' {
				old.end++
			}
			edits = append(edits, edit{old, text})
		} else if nextOK {
			appendFields += "\n" + text
		}
	}
	currentOID, _ := CardBlobOID(currentCard, format)
	if marker.Value != currentOID {
		edits = append(edits, edit{d.spans[MirrorField], mirrorLine(currentOID)})
	}
	sort.Slice(edits, func(i, j int) bool { return edits[i].start > edits[j].start })
	fm := d.fm
	for _, e := range edits {
		fm = fm[:e.start] + e.text + fm[e.end:]
	}
	fm += appendFields
	body := d.body
	if d.title != current.Title {
		body = body[:d.titleSpan.start] + "# " + current.Title + body[d.titleSpan.end:]
	}
	result := []byte(Compose(fm, body))
	if _, err := parseCardDocument(result); err != nil {
		return nil, err
	}
	return result, nil
}

// MirrorBaselineOID reads the reference a repository adapter must resolve.
func MirrorBaselineOID(details []byte) (string, error) {
	d, err := parseCardDocument(details)
	if err != nil {
		return "", err
	}
	n := d.fields[MirrorField]
	if n == nil || n.Kind != yaml.ScalarNode || n.Tag != "!!str" || !mirrorOIDRE.MatchString(strings.TrimSpace(n.Value)) {
		return "", invalidCard(MirrorField, "missing or invalid baseline")
	}
	return n.Value, nil
}

// HasMirror reports whether details claim a card mirror at all, independent of
// whether that claim is valid. It is a textual top-level key check, so malformed
// frontmatter cannot make a mirrored file look like a pre-tracker one and skip
// its ownership checks; validation belongs to MirrorBaselineOID/RefreshMirror.
func HasMirror(details []byte) bool {
	fm, _, err := Parse(string(details))
	if err != nil {
		return false
	}
	for _, line := range strings.Split(fm, "\n") {
		if strings.HasPrefix(line, MirrorField+":") {
			return true
		}
	}
	return false
}

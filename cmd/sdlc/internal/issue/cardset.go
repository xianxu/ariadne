package issue

import (
	"fmt"
	"strings"

	"github.com/xianxu/ariadne/pkg/vocab"
)

// SetCardField replaces or adds one card-owned frontmatter field and
// revalidates the whole card, so a setter can never publish a card the tracker
// reader would refuse. Unowned names and the title (an H1) are refused here.
func SetCardField(raw []byte, name, value string) ([]byte, error) {
	owned := false
	for _, f := range vocab.Issue().CardFields() {
		if f.Name == name && f.Kind != "title" {
			owned = true
		}
	}
	if !owned {
		return nil, fmt.Errorf("%q is not a card frontmatter field", name)
	}
	if strings.ContainsAny(value, "\r\n") {
		return nil, invalidCard(name, "value must be one line")
	}
	fm, body, err := Parse(string(raw))
	if err != nil {
		return nil, invalidCard("frontmatter", err.Error())
	}
	out := []byte(Compose(SetField(fm, name, value), body))
	if _, err := ParseCard(out); err != nil {
		return nil, err
	}
	return out, nil
}

// SetCardTitle replaces the card's single H1. The card path (and so the
// details filename and branch name) keeps its original slug: identity is the
// ID, and renaming paths would break every reference to them.
func SetCardTitle(raw []byte, title string) ([]byte, error) {
	title = strings.TrimSpace(title)
	if title == "" || strings.ContainsAny(title, "\r\n") {
		return nil, invalidCard("title", "must be one nonempty line")
	}
	d, err := parseCardDocument(raw)
	if err != nil {
		return nil, err
	}
	body := d.body[:d.titleSpan.start] + "# " + title + d.body[d.titleSpan.end:]
	out := []byte(Compose(d.fm, body))
	if _, err := ParseCard(out); err != nil {
		return nil, err
	}
	return out, nil
}

package issue

import (
	"errors"
	"fmt"
	"regexp"
	"strings"

	"github.com/xianxu/ariadne/pkg/vocab"
	"go.yaml.in/yaml/v3"
)

// Handoff is a card's single initial-details handoff record (#252): which
// operation published the first details to main, from where, and — once
// confirmed — as which main commit. It lives in the card's versioned internal
// envelope, is never mirrored into details, and lets a fresh clone recognise the
// source branch without the originating checkout's private recovery refs.
type Handoff struct {
	Token        string `yaml:"token"`
	Repository   string `yaml:"repository"`
	SourceBranch string `yaml:"source_branch"`
	SourceBase   string `yaml:"source_base"`
	SourceHEAD   string `yaml:"source_head"`
	SourcePath   string `yaml:"source_path"`
	SourceBlob   string `yaml:"source_blob"`
	Destination  string `yaml:"destination"`
	MainCommit   string `yaml:"main_commit,omitempty"`
}

// Completion binds a card's codecomplete to the exact close that produced it
// (#252): the reviewed HEAD and the evidence commit on the closing branch, in one
// repository. Publishing verbs select issues by this binding — a status alone
// cannot say which PR owns it — and write done only for the same token.
type Completion struct {
	Token          string `yaml:"token"`
	Repository     string `yaml:"repository"`
	ReviewedHEAD   string `yaml:"reviewed_head"`
	EvidenceCommit string `yaml:"evidence_commit"`
	LandedCommit   string `yaml:"landed_commit,omitempty"`
}

type trackerEnvelope struct {
	Version    int         `yaml:"version"`
	Handoff    *Handoff    `yaml:"handoff,omitempty"`
	Completion *Completion `yaml:"completion,omitempty"`
	Release    *Release    `yaml:"release,omitempty"`   // #284
	Abandoned  *Abandoned  `yaml:"abandoned,omitempty"` // #286
	// Extra keeps every member this binary does not know, so a rewrite never
	// drops a record a newer sdlc wrote (#284).
	Extra map[string]yaml.Node `yaml:",inline"`
}

var handoffOID = regexp.MustCompile(`^(?:[0-9a-f]{40}|[0-9a-f]{64})$`)

func (h Handoff) validate() error {
	for name, v := range map[string]string{"token": h.Token, "repository": h.Repository, "source_branch": h.SourceBranch, "source_path": h.SourcePath, "destination": h.Destination} {
		if strings.TrimSpace(v) == "" || strings.ContainsAny(v, "\r\n\x00") {
			return invalidCard("tracker.handoff."+name, "required single-line value")
		}
	}
	for name, v := range map[string]string{"source_base": h.SourceBase, "source_head": h.SourceHEAD, "source_blob": h.SourceBlob} {
		if !handoffOID.MatchString(v) {
			return invalidCard("tracker.handoff."+name, "expected a full object ID")
		}
	}
	if h.MainCommit != "" && !handoffOID.MatchString(h.MainCommit) {
		return invalidCard("tracker.handoff.main_commit", "expected a full object ID")
	}
	return nil
}

func (c Completion) validate() error {
	for name, v := range map[string]string{"token": c.Token, "repository": c.Repository} {
		if strings.TrimSpace(v) == "" || strings.ContainsAny(v, "\r\n\x00") {
			return invalidCard("tracker.completion."+name, "required single-line value")
		}
	}
	for name, v := range map[string]string{"reviewed_head": c.ReviewedHEAD, "evidence_commit": c.EvidenceCommit} {
		if !handoffOID.MatchString(v) {
			return invalidCard("tracker.completion."+name, "expected a full object ID")
		}
	}
	if c.LandedCommit != "" && !handoffOID.MatchString(c.LandedCommit) {
		return invalidCard("tracker.completion.landed_commit", "expected a full object ID")
	}
	return nil
}

// cardEnvelope decodes the card's internal transaction envelope (zero if absent).
func cardEnvelope(d *cardDocument) (trackerEnvelope, error) {
	node := d.fields[vocab.Issue().Card.Internal.Field]
	if node == nil {
		return trackerEnvelope{Version: vocab.Issue().Card.Internal.Version}, nil
	}
	var env trackerEnvelope
	if err := node.Decode(&env); err != nil {
		return env, invalidCard("tracker", err.Error())
	}
	if env.Handoff != nil {
		if err := env.Handoff.validate(); err != nil {
			return env, err
		}
	}
	if env.Completion != nil {
		if err := env.Completion.validate(); err != nil {
			return env, err
		}
	}
	if env.Release != nil {
		if err := env.Release.validate(); err != nil {
			return env, err
		}
	}
	if env.Abandoned != nil {
		if err := env.Abandoned.validate(); err != nil {
			return env, err
		}
	}
	return env, nil
}

// updateEnvelope rewrites the whole envelope after one member changes, keeping
// every other member and every other card byte.
func updateEnvelope(card []byte, change func(*trackerEnvelope) error) ([]byte, error) {
	d, err := parseCardDocument(card)
	if err != nil {
		return nil, err
	}
	env, err := cardEnvelope(d)
	if err != nil {
		return nil, err
	}
	if err := change(&env); err != nil {
		return nil, err
	}
	env.Version = vocab.Issue().Card.Internal.Version
	field := vocab.Issue().Card.Internal.Field
	fm := d.withoutField(field)
	block, err := yaml.Marshal(map[string]trackerEnvelope{field: env})
	if err != nil {
		return nil, err
	}
	out := []byte(Compose(strings.TrimRight(fm, "\n")+"\n"+strings.TrimRight(string(block), "\n"), d.body))
	if _, err := ParseCard(out); err != nil {
		return nil, fmt.Errorf("card envelope: %w", err)
	}
	return out, nil
}

// CardHandoff returns the card's handoff record, if any.
func CardHandoff(card []byte) (Handoff, bool, error) {
	d, err := parseCardDocument(card)
	if err != nil {
		return Handoff{}, false, err
	}
	env, err := cardEnvelope(d)
	if err != nil || env.Handoff == nil {
		return Handoff{}, false, err
	}
	return *env.Handoff, true, nil
}

// SetCardHandoff writes the handoff record. A card holds at most one: a record
// from another operation is refused (creation happens once), while the same
// operation may complete its own record (adding main_commit).
func SetCardHandoff(card []byte, h Handoff) ([]byte, error) {
	if err := h.validate(); err != nil {
		return nil, err
	}
	return updateEnvelope(card, func(env *trackerEnvelope) error {
		if prior := env.Handoff; prior != nil {
			completing := prior.Token == h.Token && prior.MainCommit == "" && h.MainCommit != ""
			p := *prior
			p.MainCommit = h.MainCommit
			if !completing || p != h {
				return errors.New("card already records its initial handoff; details were created once")
			}
		}
		env.Handoff = &h
		return nil
	})
}

// CardCompletion returns the card's completion binding, if any.
func CardCompletion(card []byte) (Completion, bool, error) {
	d, err := parseCardDocument(card)
	if err != nil {
		return Completion{}, false, err
	}
	env, err := cardEnvelope(d)
	if err != nil || env.Completion == nil {
		return Completion{}, false, err
	}
	return *env.Completion, true, nil
}

// SetCardCompletion records a close's binding. A new close (a new token)
// replaces an older generation; the same close may only add its landed commit.
func SetCardCompletion(card []byte, c Completion) ([]byte, error) {
	if err := c.validate(); err != nil {
		return nil, err
	}
	return updateEnvelope(card, func(env *trackerEnvelope) error {
		if prior := env.Completion; prior != nil && prior.Token == c.Token {
			p := *prior
			p.LandedCommit = c.LandedCommit
			if p != c || (prior.LandedCommit != "" && prior.LandedCommit != c.LandedCommit) {
				return errors.New("a close's completion binding cannot be rewritten")
			}
		}
		env.Completion = &c
		return nil
	})
}

// DetailsFromCard derives initial details for an issue whose creator left no
// local details: the card's fields and body mirrored exactly (baseline = this
// card blob), followed by the template's remaining sections. The internal
// envelope is transaction data and is not copied.
func DetailsFromCard(card []byte, objectFormat, today string) ([]byte, error) {
	d, err := parseCardDocument(card)
	if err != nil {
		return nil, err
	}
	fm := d.withoutField(vocab.Issue().Card.Internal.Field)
	oid, err := CardBlobOID(card, objectFormat)
	if err != nil {
		return nil, err
	}
	rendered := Render(ScaffoldSpec{ID: d.fields["id"].Value, Title: d.title, Today: today})
	_, renderedBody, err := Parse(rendered)
	if err != nil {
		return nil, err
	}
	sections := vocab.Issue().Sections()
	rest := ""
	for i, s := range sections {
		if s.Name == "Problem" && i+1 < len(sections) {
			if at := strings.Index(renderedBody, "## "+sections[i+1].Name+"\n"); at >= 0 {
				rest = renderedBody[at:]
			}
		}
	}
	body := strings.TrimRight(d.body, "\n") + "\n\n" + rest
	out := []byte(Compose(strings.TrimRight(fm, "\n")+"\ndeps: []\n"+mirrorLine(oid), body))
	if _, err := MirrorBaselineOID(out); err != nil {
		return nil, err
	}
	return out, nil
}

// withoutField returns the frontmatter with one top-level field (and its
// continuation lines) removed; every other byte is kept.
func (d *cardDocument) withoutField(name string) string {
	span, ok := d.spans[name]
	if !ok {
		return d.fm
	}
	end := span.end
	if end < len(d.fm) && d.fm[end] == '\n' {
		end++
	}
	return d.fm[:span.start] + d.fm[end:]
}

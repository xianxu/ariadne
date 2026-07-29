package gatestate

import (
	"fmt"
	"strings"

	"go.yaml.in/yaml/v3"

	"github.com/xianxu/ariadne/pkg/frontmatter"
)

// The gate ledger sidecar is TWO PROJECTIONS OF ONE LEDGER: YAML frontmatter (the machine
// view, and the only thing ParseSidecar reads) followed by generated prose (the human
// view). The prose is derived from the same Ledger at render time, so the document cannot
// disagree with itself — there is one source of truth in the file, not two.

// canonical returns l with every agent-authored string normalized. ParseFindingsBlock
// already normalizes on the way in, so this is defence in depth at the WRITE boundary: no
// code path may produce a ledger document that cannot be read back, because that would
// permanently destroy the gate's memory for the issue (see normalizeText for the
// yaml/v3 leading-newline emitter bug that makes this reachable).
func canonical(l Ledger) Ledger {
	out := l
	out.Rounds = make([]Round, len(l.Rounds))
	for i, r := range l.Rounds {
		r.Forced = normalizeText(r.Forced)
		r.ProtocolError = normalizeText(r.ProtocolError)
		r.Agent = normalizeText(r.Agent)
		r.Timestamp = normalizeText(r.Timestamp)
		nf := make([]Finding, len(r.New))
		for j, f := range r.New {
			f.Title = normalizeText(f.Title)
			f.Detail = normalizeText(f.Detail)
			nf[j] = f
		}
		r.New = nf
		if len(r.New) == 0 {
			r.New = nil
		}
		nd := make([]Disposition, len(r.Dispositions))
		for j, d := range r.Dispositions {
			d.Note = normalizeText(d.Note)
			d.ID = strings.TrimSpace(d.ID)
			nd[j] = d
		}
		r.Dispositions = nd
		if len(r.Dispositions) == 0 {
			r.Dispositions = nil
		}
		out.Rounds[i] = r
	}
	if len(out.Rounds) == 0 {
		out.Rounds = nil
	}
	return out
}

// Render writes the ledger as a self-contained markdown document. Pure: the caller supplies
// the repo name; no clock, no filesystem. The output is always CANONICAL and always
// re-readable by ParseSidecar (FuzzRenderParseRoundTrip pins that).
func Render(rawLedger Ledger, repo string) string {
	l := canonical(rawLedger)
	var b strings.Builder

	fm, err := yaml.Marshal(l)
	if err != nil {
		// Ledger holds only strings, ints, bools and slices of those — Marshal cannot
		// fail on it. Surface rather than silently emitting a headless document.
		panic("gatestate: marshal ledger: " + err.Error())
	}
	b.WriteString("---\n")
	b.Write(fm)
	b.WriteString("---\n\n")

	if repo == "" {
		repo = "<unknown-repo>"
	}
	fmt.Fprintf(&b, "# Gate ledger — %s#%d (%s)\n\n", repo, l.IssueNum, l.Gate)
	b.WriteString("Findings this gate raised, the stable ids the binary assigned them, and how\n")
	b.WriteString("later rounds disposed of them. Generated — edit the gate, not this file.\n")

	for _, r := range l.Rounds {
		outcome := "passed"
		if r.Blocked {
			outcome = "BLOCKED"
		}
		fmt.Fprintf(&b, "\n## Round %d — %s (%s) — %s\n", r.N, r.Timestamp, r.Agent, outcome)
		if r.ProtocolError != "" {
			fmt.Fprintf(&b, "\n**Protocol error:** %s — this round contributed no findings.\n", r.ProtocolError)
		}
		if r.Forced != "" {
			fmt.Fprintf(&b, "\n**Forced past** (`--force`): %s\n", r.Forced)
		}
		if len(r.Dispositions) > 0 {
			b.WriteString("\n### Disposed\n\n")
			for _, d := range r.Dispositions {
				fmt.Fprintf(&b, "- %s — %s", d.ID, d.State)
				if d.Note != "" {
					fmt.Fprintf(&b, " — %s", d.Note)
				}
				b.WriteString("\n")
			}
		}
		if len(r.New) > 0 {
			b.WriteString("\n### Raised\n\n")
			for _, f := range r.New {
				fmt.Fprintf(&b, "- **%s** [%s] %s\n", f.ID, f.Severity, f.Title)
				if f.Detail != "" {
					fmt.Fprintf(&b, "  %s\n", strings.ReplaceAll(f.Detail, "\n", "\n  "))
				}
			}
		}
	}

	open := OpenFindings(l)
	b.WriteString("\n## Open findings\n\n")
	if len(open) == 0 {
		b.WriteString("(none — every finding has been disposed)\n")
		return b.String()
	}
	for _, f := range open {
		fmt.Fprintf(&b, "- **%s** [%s] %s\n", f.ID, f.Severity, f.Title)
	}
	return b.String()
}

// ParseSidecar reads a rendered sidecar back into a Ledger. It reads ONLY the frontmatter
// — the prose below is a derived projection, never a parse target.
//
// A document with no frontmatter, or frontmatter that doesn't parse, is an ERROR. The
// caller must not treat it as an empty ledger: silently resetting would erase every
// disposition and re-open findings the operator already addressed, which is precisely the
// forgetting this package exists to prevent.
func ParseSidecar(text string) (Ledger, error) {
	fm, _, err := frontmatter.Split(text)
	if err != nil {
		return Ledger{}, fmt.Errorf("gate ledger has no YAML frontmatter: %w", err)
	}
	var l Ledger
	if err := yaml.Unmarshal([]byte(fm), &l); err != nil {
		return Ledger{}, fmt.Errorf("gate ledger frontmatter does not parse: %w", err)
	}
	return l, nil
}

package project

import (
	"fmt"
	"regexp"
	"strings"
)

// GuardCtx is what a lifecycle guard may consult besides the doc. IssueTerminal
// resolves an mvp_scope ref ("repo#N") to whether that issue has reached a
// terminal status; the caller injects it from the cross-repo card lookup, so
// this package stays free of IO.
type GuardCtx struct {
	Today         string
	IssueTerminal func(ref string) (bool, error)
}

type GuardFunc func(d *Doc, ctx GuardCtx) error

var (
	retroHeadingRE = regexp.MustCompile(`(?m)^### \d{4}-\d{2}-\d{2} — retro\b`)
)

// Guards is the registry of the guards project.cue's lifecycle names (#329).
func Guards() map[string]GuardFunc {
	return map[string]GuardFunc{
		// defined → executing: the TL has laid out at least one train.
		"trains-present": func(d *Doc, _ GuardCtx) error {
			p := d.Trains("")
			if len(p.Trains) == 0 {
				msg := "## Trains must hold a `| Train | Cars, in order | Slot |` table with at least one train"
				if len(p.Warnings) > 0 {
					msg += " (" + strings.Join(p.Warnings, "; ") + ")"
				}
				return fmt.Errorf("%s", msg)
			}
			return nil
		},
		// executing → closing: every mvp_scope issue has landed.
		"mvp-landed": func(d *Doc, ctx GuardCtx) error {
			meta, err := d.Metadata()
			if err != nil {
				return err
			}
			if len(meta.MVPScope) == 0 {
				return fmt.Errorf("mvp_scope is empty; list the MVP issues before wrapping")
			}
			if ctx.IssueTerminal == nil {
				return fmt.Errorf("no issue lookup available to check mvp_scope")
			}
			mvp := MVPProgress(meta.MVPScope, ctx.IssueTerminal)
			if len(mvp.Errors) > 0 {
				return mvp.Errors[0]
			}
			if len(mvp.Open) > 0 {
				return fmt.Errorf("mvp_scope issues not landed: %s", strings.Join(mvp.Open, ", "))
			}
			return nil
		},
		"retro-recorded": func(d *Doc, _ GuardCtx) error {
			if !retroHeadingRE.MatchString(d.SectionBody("Log")) {
				return fmt.Errorf("Log must contain a dated retro heading")
			}
			return nil
		},
	}
}

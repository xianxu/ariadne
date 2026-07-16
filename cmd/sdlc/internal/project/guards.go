package project

import (
	"fmt"
	"regexp"
	"strings"
)

type GuardCtx struct {
	Evidence map[string]string
	Today    string
}

type GuardFunc func(d *Doc, ctx GuardCtx) error

var (
	retroHeadingRE = regexp.MustCompile(`(?m)^### \d{4}-\d{2}-\d{2} — retro\b`)
)

func Guards() map[string]GuardFunc {
	return map[string]GuardFunc{
		"prd-present": func(d *Doc, _ GuardCtx) error {
			body := strings.TrimSpace(d.SectionBody("PRD"))
			if body == "" || strings.HasPrefix(body, "<") {
				return fmt.Errorf("PRD must contain substantive prose")
			}
			return nil
		},
		"phase-a-estimate": func(d *Doc, _ GuardCtx) error {
			_, present, err := ParsePhaseA(d.SectionBody("Estimate"))
			if err != nil {
				return err
			}
			if !present {
				return fmt.Errorf("Estimate must contain **phase-a:** <N>h")
			}
			return nil
		},
		"baseline-set": func(d *Doc, _ GuardCtx) error {
			if d.FM("deadline") == "" || d.FM("planned_finish") == "" {
				return fmt.Errorf("deadline and planned_finish must both be set")
			}
			return nil
		},
		"reality-check":    evidenceGuard("reality-check"),
		"issues-cover-prd": evidenceGuard("issues-cover-prd"),
		"retro-recorded": func(d *Doc, _ GuardCtx) error {
			if !retroHeadingRE.MatchString(d.SectionBody("Log")) {
				return fmt.Errorf("Log must contain a dated retro heading")
			}
			return nil
		},
		"fog-factor-recorded": func(*Doc, GuardCtx) error { return fmt.Errorf("fog factor is recorded by `sdlc project close`") },
	}
}

func evidenceGuard(name string) GuardFunc {
	return func(_ *Doc, ctx GuardCtx) error {
		if strings.TrimSpace(ctx.Evidence[name]) == "" {
			return fmt.Errorf("%s evidence is required", name)
		}
		return nil
	}
}

package issue

import (
	"path"
	"strings"
)

const (
	LegacyRetained           = "retained"
	LegacyWindowlessClose    = "windowless-close"
	LegacyWindowlessTerminal = "windowless-terminal"
)

// LegacyAdoption pins the imported numeric generation to its pre-cutover
// source. Close-backed records name the exact completion; terminal-only
// records preserve provenance without manufacturing a completion.
type LegacyAdoption struct {
	Repository      string `yaml:"repository"`
	SourceMain      string `yaml:"source_main"`
	SourcePath      string `yaml:"source_path"`
	SourceBlob      string `yaml:"source_blob"`
	TerminalCommit  string `yaml:"terminal_commit"`
	CompletionToken string `yaml:"completion_token,omitempty"`
	WindowClass     string `yaml:"window_class"`
}

func (a LegacyAdoption) validate() error {
	if strings.TrimSpace(a.Repository) == "" || strings.ContainsAny(a.Repository, "\r\n\x00") {
		return invalidCard("tracker.legacy_adoption.repository", "required single-line value")
	}
	for name, value := range map[string]string{"source_main": a.SourceMain, "source_blob": a.SourceBlob, "terminal_commit": a.TerminalCommit} {
		if !handoffOID.MatchString(value) {
			return invalidCard("tracker.legacy_adoption."+name, "expected a full object ID")
		}
	}
	p := a.SourcePath
	if p == "" || p == "." || path.IsAbs(p) || path.Clean(p) != p || p == ".." || strings.HasPrefix(p, "../") || strings.TrimSpace(p) != p || strings.ContainsAny(p, "\\:\r\n\x00") {
		return invalidCard("tracker.legacy_adoption.source_path", "expected a canonical repository-relative path")
	}
	switch a.WindowClass {
	case LegacyRetained, LegacyWindowlessClose:
		if strings.TrimSpace(a.CompletionToken) == "" || strings.ContainsAny(a.CompletionToken, "\r\n\x00") {
			return invalidCard("tracker.legacy_adoption.completion_token", "required single-line value for a close-backed generation")
		}
	case LegacyWindowlessTerminal:
		if a.CompletionToken != "" {
			return invalidCard("tracker.legacy_adoption.completion_token", "terminal-only generation cannot have a completion token")
		}
	default:
		return invalidCard("tracker.legacy_adoption.window_class", "unknown legacy window class")
	}
	return nil
}

func (a LegacyAdoption) validateCompletion(c *Completion) error {
	if a.WindowClass == LegacyWindowlessTerminal {
		if c != nil {
			return invalidCard("tracker.legacy_adoption", "terminal-only generation cannot have a completion")
		}
	} else if c == nil || c.Token != a.CompletionToken || c.Repository != a.Repository {
		return invalidCard("tracker.legacy_adoption", "completion binding mismatch")
	}
	return nil
}

// CardLegacyAdoption returns explicit legacy provenance, if present. Absence
// does not imply a legacy generation or grant a missing measurement window.
func CardLegacyAdoption(card []byte) (LegacyAdoption, bool, error) {
	d, err := parseCardDocument(card)
	if err != nil {
		return LegacyAdoption{}, false, err
	}
	env, err := cardEnvelope(d)
	if err != nil || env.LegacyAdoption == nil {
		return LegacyAdoption{}, false, err
	}
	return *env.LegacyAdoption, true, nil
}

// SetCardLegacyAdoption records provenance for the current imported generation.
// Close-backed callers install Completion first; terminal-only callers leave it absent.
func SetCardLegacyAdoption(card []byte, a LegacyAdoption) ([]byte, error) {
	if err := a.validate(); err != nil {
		return nil, err
	}
	return updateEnvelope(card, func(env *trackerEnvelope) error {
		if err := a.validateCompletion(env.Completion); err != nil {
			return err
		}
		env.LegacyAdoption = &a
		return nil
	})
}

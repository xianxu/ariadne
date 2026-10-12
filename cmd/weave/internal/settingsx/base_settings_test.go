package settingsx

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestBaseSettingsExcludedVerbsMatchArguments guards the shipped base-layer
// policy (#334). Claude Code reads sandbox.excludedCommands entries with
// Bash(...) rule syntax: a bare word is an EXACT match, so `"couch"` exempts
// only `couch` typed with no arguments, and `couch --send-to …` stays
// sandboxed. A verb the base pre-allows with arguments (`Bash(X:*)` or
// `Bash(X *)` in permissions.allow) and also excludes from the sandbox is
// meant to run unsandboxed with arguments, so it must be listed as `X *`.
func TestBaseSettingsExcludedVerbsMatchArguments(t *testing.T) {
	path := filepath.Join("..", "..", "..", "..", ".claude", "settings.ariadne.json")
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read base settings: %v", err)
	}
	var s struct {
		Permissions struct {
			Allow []string `json:"allow"`
		} `json:"permissions"`
		Sandbox struct {
			ExcludedCommands []string `json:"excludedCommands"`
		} `json:"sandbox"`
	}
	if err := json.Unmarshal(raw, &s); err != nil {
		t.Fatalf("parse base settings: %v", err)
	}
	allowedWithArgs := map[string]bool{}
	for _, rule := range s.Permissions.Allow {
		inner, ok := strings.CutPrefix(rule, "Bash(")
		if !ok {
			continue
		}
		inner = strings.TrimSuffix(inner, ")")
		for _, suffix := range []string{":*", " *"} {
			if verb, ok := strings.CutSuffix(inner, suffix); ok {
				allowedWithArgs[verb] = true
			}
		}
	}
	excluded := map[string]bool{}
	for _, entry := range s.Sandbox.ExcludedCommands {
		excluded[entry] = true
	}
	for _, entry := range s.Sandbox.ExcludedCommands {
		if allowedWithArgs[entry] {
			t.Errorf("sandbox.excludedCommands lists %q, an exact match that exempts only the bare verb; "+
				"permissions.allow pre-allows it with arguments, so list it as %q", entry, entry+" *")
		}
	}
	// The workbench verbs #318 meant to exempt must actually be exempt.
	for _, verb := range []string{"sdlc", "couch"} {
		if !excluded[verb+" *"] {
			t.Errorf("sandbox.excludedCommands lacks %q; agents' `%s …` calls run sandboxed", verb+" *", verb)
		}
	}
}

package reviewusage

import (
	"bytes"
	"crypto/rand"
	"encoding/json"
	"errors"
	"fmt"
	"regexp"
	"strings"

	"github.com/xianxu/ariadne/cmd/sdlc/internal/jsonstrict"
)

const IdentityLimit = 64 << 10
const claudeIdentityLimit = 4 << 20

var codexBanner = regexp.MustCompile(`^OpenAI Codex (v[0-9]+\.[0-9]+\.[0-9]+|\(v[0-9]+\.[0-9]+\.[0-9]+\))$`)

func NewID() (string, error) {
	var b [16]byte
	if _, err := rand.Read(b[:]); err != nil {
		return "", err
	}
	b[6] = (b[6] & 0x0f) | 0x40
	b[8] = (b[8] & 0x3f) | 0x80
	return fmt.Sprintf("%x-%x-%x-%x-%x", b[:4], b[4:6], b[6:8], b[8:10], b[10:]), nil
}

// CaptureIdentity trusts only the CLI's protocol metadata, never quoted review
// prose, model output, a cwd/time guess or a prompt-echoed session ID.
func CaptureIdentity(agent string, stdout, stderr []byte) (string, error) {
	switch agent {
	case "claude", "":
		return claudeIdentity(stdout)
	case "codex":
		return codexIdentity(stdout, stderr)
	default:
		return "", errors.New("unsupported-session-identity-protocol")
	}
}
func claudeIdentity(raw []byte) (string, error) {
	if len(raw) > claudeIdentityLimit {
		return "", errors.New("session-identity-output-limit")
	}
	id := ""
	for _, line := range bytes.Split(raw, []byte("\n")) {
		var r struct {
			Type      string `json:"type"`
			Subtype   string `json:"subtype"`
			SessionID string `json:"session_id"`
		}
		if json.Unmarshal(line, &r) != nil || !(r.Type == "result" || r.Type == "system" && r.Subtype == "init") {
			continue
		}
		if r.SessionID == "" {
			continue
		}
		if err := jsonstrict.Scan(json.NewDecoder(bytes.NewReader(line)), 16); err != nil {
			return "", errors.New("invalid-session-identity-event")
		}
		if !sessionUUID.MatchString(r.SessionID) {
			return "", errors.New("invalid-session-identity")
		}
		if id != "" && id != r.SessionID {
			return "", errors.New("conflicting-session-identity")
		}
		id = r.SessionID
	}
	if id == "" {
		return "", errors.New("missing-session-identity")
	}
	return id, nil
}
func codexIdentity(stdout, stderr []byte) (string, error) {
	// Codex emits startup metadata on stderr. Stdout is plain model prose
	// and cannot authenticate a fabricated startup banner.
	id, err := codexPreamble(stderr)
	if err != nil {
		return "", errors.New("missing-or-invalid-codex-startup-identity")
	}
	return id, nil
}
func codexPreamble(raw []byte) (string, error) {
	if len(raw) > IdentityLimit {
		raw = raw[:IdentityLimit]
	}
	lines := strings.Split(strings.ReplaceAll(string(raw), "\r\n", "\n"), "\n")
	// With piped stdin, current Codex prints this fixed CLI notice before
	// its banner. Accept one exact notice, never search past arbitrary prose.
	if len(lines) > 0 && lines[0] == "Reading additional input from stdin..." {
		lines = lines[1:]
	}
	if len(lines) < 3 || !codexBanner.MatchString(lines[0]) || lines[1] != "--------" {
		return "", errors.New("missing startup preamble")
	}
	allowed := map[string]bool{"workdir": true, "model": true, "provider": true, "approval": true, "sandbox": true, "reasoning effort": true, "reasoning summaries": true, "session id": true}
	seen := map[string]bool{}
	id := ""
	for _, line := range lines[2:] {
		if line == "--------" {
			if id != "" && seen["workdir"] && seen["model"] && seen["provider"] {
				return id, nil
			}
			return "", errors.New("incomplete startup metadata")
		}
		key, value, ok := strings.Cut(line, ": ")
		if !ok || !allowed[key] || seen[key] || strings.TrimSpace(value) == "" {
			return "", errors.New("invalid startup metadata")
		}
		seen[key] = true
		if key == "session id" {
			if !sessionUUID.MatchString(value) {
				return "", errors.New("invalid startup session UUID")
			}
			id = value
		}
	}
	return "", errors.New("truncated startup preamble")
}

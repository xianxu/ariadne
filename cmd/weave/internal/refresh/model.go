// Package refresh advances an existing layer graph using captured Git evidence.
package refresh

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/xianxu/ariadne/pkg/layergraph"
)

type Snapshot struct {
	path, common, gitdir, branch, head, origin, target string
	rows                                               []layergraph.Dependency
}
type Prepared struct{ snapshots []Snapshot }
type Phase uint8

const (
	inspecting Phase = iota
	ready
	applying
	compiling
	complete
	stopped
)

type event uint8

const (
	checked event = iota
	validated
	confirmed
	finished
	compiled
	failed
)

func advance(p Phase, e event) (Phase, error) {
	if e == failed && p != complete && p != stopped {
		return stopped, nil
	}
	switch {
	case p == inspecting && e == checked:
		return ready, nil
	case p == ready && e == validated:
		return applying, nil
	case p == applying && e == confirmed:
		return applying, nil
	case p == applying && e == finished:
		return compiling, nil
	case p == compiling && e == compiled:
		return complete, nil
	}
	return p, fmt.Errorf("invalid refresh transition %d/%d", p, e)
}
func eligibility(equal, ancestor, rebase bool) bool { return equal || ancestor || rebase }
func parseRecord(s string) (string, error) {
	if !strings.HasSuffix(s, "\n") || strings.ContainsAny(strings.TrimSuffix(s, "\n"), "\n\x00") {
		return "", fmt.Errorf("invalid Git record framing")
	}
	return strings.TrimSuffix(s, "\n"), nil
}

var oidPattern = regexp.MustCompile(`^(?:[0-9a-f]{40}|[0-9a-f]{64})$`)

func parseOID(s string) (string, error) {
	v, e := parseRecord(s)
	if e != nil || !oidPattern.MatchString(v) || strings.Trim(v, "0") == "" {
		return "", fmt.Errorf("invalid full Git object ID")
	}
	return v, nil
}
func parseBranch(s string) (string, error) {
	v, e := parseRecord(s)
	if e != nil || !strings.HasPrefix(v, "refs/heads/") || strings.TrimPrefix(v, "refs/heads/") == "" || strings.ContainsAny(v, " ~^:?*[\\\t\r") || strings.Contains(v, "..") || strings.Contains(v, "@{") || strings.Contains(v, "//") || strings.HasSuffix(v, "/") || strings.HasSuffix(v, ".") {
		return "", fmt.Errorf("invalid symbolic branch")
	}
	for _, r := range v {
		if r <= 0x20 || r == 0x7f {
			return "", fmt.Errorf("invalid symbolic branch")
		}
	}
	for _, part := range strings.Split(v, "/") {
		if strings.HasPrefix(part, ".") || strings.HasSuffix(part, ".lock") {
			return "", fmt.Errorf("invalid symbolic branch")
		}
	}
	return v, nil
}

func parseNULRecords(s string) ([]string, error) {
	if s == "" {
		return nil, nil
	}
	if !strings.HasSuffix(s, "\x00") {
		return nil, fmt.Errorf("invalid NUL record framing")
	}
	rows := strings.Split(strings.TrimSuffix(s, "\x00"), "\x00")
	for _, row := range rows {
		if row == "" {
			return nil, fmt.Errorf("empty NUL record")
		}
	}
	return rows, nil
}

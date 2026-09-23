package workspace

import (
	"fmt"
	"path/filepath"
	"strconv"
	"strings"
	"unicode"
	"unicode/utf8"
)

type Address struct {
	Repo string
	Slot int
}

// validRepoName accepts one exact portable basename, including spaces and
// Unicode. Addresses never interpret names as paths or ref expressions.
func validRepoName(s string) bool {
	if s == "" || s == "." || s == ".." || !utf8.ValidString(s) || strings.ContainsAny(s, "/\\:") {
		return false
	}
	for _, r := range s {
		if unicode.IsControl(r) {
			return false
		}
	}
	return true
}

func (a Address) String() string { return a.Repo + ":" + strconv.Itoa(a.Slot) }
func ParseAddress(s string) (Address, error) {
	parts := strings.Split(s, ":")
	if len(parts) > 2 || s == "" {
		return Address{}, fmt.Errorf("invalid workspace address %q", s)
	}
	a := Address{Repo: parts[0]}
	if a.Repo != "" && !validRepoName(a.Repo) {
		return Address{}, fmt.Errorf("invalid repository name %q", a.Repo)
	}
	if len(parts) == 1 {
		return a, nil
	}
	n, e := strconv.Atoi(parts[1])
	if e != nil || n < 0 || strconv.Itoa(n) != parts[1] {
		return Address{}, fmt.Errorf("invalid workspace slot %q", parts[1])
	}
	a.Slot = n
	return a, nil
}
func SlotEnvironmentPath(fleet, repo string, slot int) (string, error) {
	if !validRepoName(repo) || slot < 1 {
		return "", fmt.Errorf("invalid numbered slot %q:%d", repo, slot)
	}
	return filepath.Join(fleet, "worktree", repo+"-slot"+strconv.Itoa(slot)), nil
}

func SlotPath(fleet, repo string, slot int) (string, error) {
	env, err := SlotEnvironmentPath(fleet, repo, slot)
	if err != nil {
		return "", err
	}
	return filepath.Join(env, repo), nil
}

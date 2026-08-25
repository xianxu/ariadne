package gitx

import (
	"bytes"
	"fmt"
	"strings"
)

// Worktree is one record from `git worktree list --porcelain -z`. Branch is the
// short local branch name when checked out; Detached and Bare represent the
// two mutually exclusive non-branch states. Locked and Prunable are pointers
// so an attribute with no reason remains distinct from an absent attribute.
type Worktree struct {
	Path     string
	HEAD     string
	Branch   string
	Detached bool
	Bare     bool
	Locked   *string
	Prunable *string
}

// ParseWorktrees parses the complete NUL-delimited `git worktree list --porcelain -z` grammar.
// It preserves attributes that current callers do not render so future callers
// do not have to reparse Git output. A final record need not end in a blank
// separator, but every record must otherwise be complete and well-formed.
func ParseWorktrees(porcelain []byte) ([]Worktree, error) {
	var result []Worktree
	var current worktreeRecord
	inRecord := false

	finish := func() error {
		if !inRecord {
			return nil
		}
		worktree, err := current.finish()
		if err != nil {
			return err
		}
		result = append(result, worktree)
		current = worktreeRecord{}
		inRecord = false
		return nil
	}

	for fieldNum, line := range bytes.Split(porcelain, []byte{0}) {
		if len(line) == 0 {
			if err := finish(); err != nil {
				return nil, fmt.Errorf("worktree record ending at field %d: %w", fieldNum+1, err)
			}
			continue
		}
		if !inRecord {
			if !bytes.HasPrefix(line, []byte("worktree ")) {
				return nil, fmt.Errorf("field %d: record must begin with worktree", fieldNum+1)
			}
			path := string(line[len("worktree "):])
			if path == "" {
				return nil, fmt.Errorf("field %d: worktree path is empty", fieldNum+1)
			}
			current = worktreeRecord{worktree: Worktree{Path: path}}
			inRecord = true
			continue
		}
		if bytes.HasPrefix(line, []byte("worktree ")) {
			return nil, fmt.Errorf("field %d: worktree record is missing a NUL record separator", fieldNum+1)
		}
		if err := current.add(string(line)); err != nil {
			return nil, fmt.Errorf("field %d: %w", fieldNum+1, err)
		}
	}
	if err := finish(); err != nil {
		return nil, err
	}
	return result, nil
}

type worktreeRecord struct {
	worktree     Worktree
	headSeen     bool
	stateSeen    bool
	lockedSeen   bool
	prunableSeen bool
}

func (r *worktreeRecord) add(line string) error {
	switch {
	case strings.HasPrefix(line, "HEAD "):
		if r.headSeen {
			return fmt.Errorf("duplicate HEAD")
		}
		head := strings.TrimPrefix(line, "HEAD ")
		if head == "" || strings.ContainsAny(head, " \t\r\n") {
			return fmt.Errorf("HEAD is malformed")
		}
		r.worktree.HEAD = head
		r.headSeen = true
		return nil
	case strings.HasPrefix(line, "branch "):
		if r.stateSeen {
			return fmt.Errorf("multiple checkout states")
		}
		const prefix = "branch refs/heads/"
		if !strings.HasPrefix(line, prefix) {
			return fmt.Errorf("branch is not a local refs/heads ref")
		}
		branch := strings.TrimPrefix(line, prefix)
		if branch == "" || strings.ContainsAny(branch, " \t\r\n") {
			return fmt.Errorf("branch name is malformed")
		}
		r.worktree.Branch = branch
		r.stateSeen = true
		return nil
	case line == "detached":
		if r.stateSeen {
			return fmt.Errorf("multiple checkout states")
		}
		r.worktree.Detached = true
		r.stateSeen = true
		return nil
	case line == "bare":
		if r.stateSeen {
			return fmt.Errorf("multiple checkout states")
		}
		r.worktree.Bare = true
		r.stateSeen = true
		return nil
	case line == "locked" || strings.HasPrefix(line, "locked "):
		if r.lockedSeen {
			return fmt.Errorf("duplicate locked attribute")
		}
		reason := ""
		if line != "locked" {
			reason = strings.TrimPrefix(line, "locked ")
		}
		r.worktree.Locked = &reason
		r.lockedSeen = true
		return nil
	case line == "prunable" || strings.HasPrefix(line, "prunable "):
		if r.prunableSeen {
			return fmt.Errorf("duplicate prunable attribute")
		}
		reason := ""
		if line != "prunable" {
			reason = strings.TrimPrefix(line, "prunable ")
		}
		r.worktree.Prunable = &reason
		r.prunableSeen = true
		return nil
	default:
		return fmt.Errorf("unknown worktree attribute %q", line)
	}
}

func (r worktreeRecord) finish() (Worktree, error) {
	if r.worktree.Path == "" {
		return Worktree{}, fmt.Errorf("worktree path is empty")
	}
	if r.worktree.Bare {
		if r.headSeen {
			return Worktree{}, fmt.Errorf("bare worktree must not contain HEAD")
		}
		return r.worktree, nil
	}
	if !r.headSeen {
		return Worktree{}, fmt.Errorf("worktree is missing HEAD")
	}
	if !r.stateSeen {
		return Worktree{}, fmt.Errorf("worktree is missing checkout state")
	}
	return r.worktree, nil
}

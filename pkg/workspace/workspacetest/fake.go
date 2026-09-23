// Package workspacetest provides stateful local Git observations for workspace tests.
package workspacetest

import (
	"fmt"
	"github.com/xianxu/ariadne/pkg/workspace"
	"path/filepath"
	"strings"
)

type Repository struct {
	CommonDir string
	Worktrees []workspace.Worktree
	Refs      map[string]string
}
type Fake struct {
	Repositories []*Repository
	Reads        int
	Commands     []string
	// Before and After permit deterministic mutation and fault injection at a read ordinal.
	Before  func(*Fake, int)
	After   func(*Fake, int)
	Faults  map[int]error
	Outputs map[int][]byte
}

func (f *Fake) GitInDir(dir string, args ...string) (out []byte, err error) {
	f.Reads++
	f.Commands = append(f.Commands, strings.Join(args, " "))
	n := f.Reads
	if f.Before != nil {
		f.Before(f, n)
	}
	defer func() {
		if f.After != nil {
			f.After(f, n)
		}
	}()
	if e := f.Faults[n]; e != nil {
		return nil, e
	}
	if b, ok := f.Outputs[n]; ok {
		return b, nil
	}
	dir, e := workspace.CanonicalPath(dir)
	if e != nil {
		return nil, e
	}
	for _, r := range f.Repositories {
		for _, w := range r.Worktrees {
			if dir != w.Path && !strings.HasPrefix(dir, w.Path+string(filepath.Separator)) {
				continue
			}
			switch strings.Join(args, " ") {
			case "rev-parse --show-toplevel":
				return []byte(w.Path + "\n"), nil
			case "rev-parse --git-common-dir":
				return []byte(r.CommonDir + "\n"), nil
			case "worktree list --porcelain -z":
				return Porcelain(r.Worktrees), nil
			case "symbolic-ref -q HEAD":
				if w.Detached {
					return nil, exitError(1)
				}
				return []byte("refs/heads/" + w.Branch + "\n"), nil
			case "rev-parse --verify HEAD":
				if strings.Trim(w.HEAD, "0") == "" {
					return nil, fmt.Errorf("unborn")
				}
				return []byte(w.HEAD + "\n"), nil
			}
			if len(args) == 3 && args[0] == "for-each-ref" && args[1] == "--format=%(refname)" {
				if oid := r.Refs[strings.TrimPrefix(args[2], "refs/heads/")]; oid != "" {
					return []byte(args[2] + "\n"), nil
				}
				return nil, nil
			}
			if len(args) == 4 && args[0] == "rev-parse" && args[1] == "--verify" && args[2] == "--end-of-options" {
				ref := strings.TrimSuffix(args[3], "^{commit}")
				ref = strings.TrimPrefix(ref, "refs/heads/")
				if ref == "HEAD" {
					if strings.Trim(w.HEAD, "0") != "" {
						return []byte(w.HEAD + "\n"), nil
					}
				} else if oid := r.Refs[ref]; oid != "" {
					return []byte(oid + "\n"), nil
				}
				return nil, fmt.Errorf("missing commit ref %s", ref)
			}
			return nil, fmt.Errorf("unsupported fake Git command: %v", args)
		}
	}
	return nil, fmt.Errorf("not a registered worktree: %s", dir)
}
func Porcelain(trees []workspace.Worktree) []byte {
	var b strings.Builder
	for _, w := range trees {
		fmt.Fprintf(&b, "worktree %s\x00", w.Path)
		if w.Bare {
			b.WriteString("bare\x00")
		} else {
			fmt.Fprintf(&b, "HEAD %s\x00", w.HEAD)
			if w.Detached {
				b.WriteString("detached\x00")
			} else {
				fmt.Fprintf(&b, "branch refs/heads/%s\x00", w.Branch)
			}
		}
		if w.Locked != nil {
			fmt.Fprintf(&b, "locked %s\x00", *w.Locked)
		}
		if w.Prunable != nil {
			fmt.Fprintf(&b, "prunable %s\x00", *w.Prunable)
		}
		b.WriteByte(0)
	}
	return []byte(b.String())
}

type exitError int

func (e exitError) Error() string { return fmt.Sprintf("exit status %d", int(e)) }
func (e exitError) ExitCode() int { return int(e) }

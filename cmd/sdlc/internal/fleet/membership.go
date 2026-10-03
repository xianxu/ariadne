package fleet

import (
	"errors"
	"io"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/xianxu/ariadne/pkg/layergraph"
)

// MemberState is what the walk found at a declared member's path.
type MemberState string

const (
	MemberPresent MemberState = "present" // the directory exists
	MemberMissing MemberState = "missing" // declared, not there
	MemberOutside MemberState = "outside" // declared outside the environment root
)

// MemberDecl is one declared substrate dependency of a numbered slot.
type MemberDecl struct {
	Path  string
	State MemberState
}

// DeclaredMembers walks a numbered slot's substrate dependencies from its
// host (#289): weave's `construct/deps` grammar (layergraph.ParseRows),
// `substrate` rows only, each path relative to the checkout declaring it,
// transitively, in discovery order. A member must sit directly under envRoot
// (weave's slot policy); an absent member is still returned (missing). Cycles
// end on a visited set. A declaration that cannot be read or parsed is
// returned per declaring directory, so its checkout can be judged unknown.
// Pure over read (content, found, err) and stat.
func DeclaredMembers(host, envRoot string, read func(dir string) (string, bool, error), stat func(string) error) ([]MemberDecl, map[string]string) {
	var members []MemberDecl
	declErrs := map[string]string{}
	visited := map[string]bool{filepath.Clean(host): true}
	queue := []string{filepath.Clean(host)}
	for len(queue) > 0 {
		dir := queue[0]
		queue = queue[1:]
		content, found, err := read(dir)
		if err != nil {
			declErrs[dir] = err.Error()
			continue
		}
		if !found {
			continue // a leaf: no declared dependencies
		}
		rows, err := layergraph.ParseRows(content)
		if err != nil {
			declErrs[dir] = err.Error()
			continue
		}
		for _, row := range rows {
			if row.Kind != "substrate" {
				continue
			}
			p := row.Path
			if !filepath.IsAbs(p) {
				p = filepath.Join(dir, p)
			}
			p = filepath.Clean(p)
			if visited[p] {
				continue
			}
			visited[p] = true
			switch err := stat(p); {
			case filepath.Dir(p) != filepath.Clean(envRoot):
				members = append(members, MemberDecl{Path: p, State: MemberOutside})
			case errors.Is(err, fs.ErrNotExist):
				members = append(members, MemberDecl{Path: p, State: MemberMissing})
			default:
				// Present (or unreadable: the row collection reports that).
				members = append(members, MemberDecl{Path: p, State: MemberPresent})
				queue = append(queue, p)
			}
		}
	}
	return members, declErrs
}

// readDeclaration reads a checkout's `construct/deps` with weave's bound.
func readDeclaration(dir string) (string, bool, error) {
	f, err := os.Open(filepath.Join(dir, "construct", "deps"))
	if errors.Is(err, fs.ErrNotExist) {
		return "", false, nil
	}
	if err != nil {
		return "", false, err
	}
	defer f.Close()
	const limit = 1 << 20
	buf := make([]byte, limit+1)
	n, err := io.ReadFull(f, buf)
	if err != nil && !errors.Is(err, io.ErrUnexpectedEOF) && !errors.Is(err, io.EOF) {
		return "", false, err
	}
	if n > limit {
		return "", false, errors.New("construct/deps exceeds 1 MiB")
	}
	return string(buf[:n]), true, nil
}

func statPath(p string) error {
	_, err := os.Stat(p)
	return err
}

package agentid

import (
	"bufio"
	"bytes"
	"context"
	"errors"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/xianxu/ariadne/cmd/sdlc/internal/qodertranscript"
)

const qoderCandidateLimit = 256
const qoderReadLimit = 16 << 20
const qoderSessionReadLimit = 4 << 20

type qoderCandidate struct{ id, model, effort string }

// selectQoderSession is deliberately independent of file/record recency.
// Replayed or newer observations cannot replace the invoking launch identity.
func selectQoderSession(candidates []qoderCandidate, bound string) Identity {
	out := Identity{Agent: "qoder"}
	var chosen *qoderCandidate
	for _, c := range candidates {
		if bound != "" && c.id != bound {
			continue
		}
		if chosen != nil && *chosen != c {
			return out
		}
		v := c
		chosen = &v
	}
	if chosen != nil {
		out.Model, out.Effort = chosen.model, chosen.effort
	}
	return out
}

func (d Detector) qoderIdentity(parent context.Context) Identity {
	unknown := Identity{Agent: "qoder"}
	cwd := d.CWD
	if cwd == "" {
		var err error
		cwd, err = os.Getwd()
		if err != nil {
			return unknown
		}
	}
	cwd = qoderWorkspace(cwd)
	now := time.Now()
	if d.Now != nil {
		now = d.Now()
	}
	ctx, cancel := context.WithTimeout(parent, 2*time.Second)
	defer cancel()
	bound := d.Getenv("PAIR_SESSION_ID")
	// Match names as data rather than interpolating an environment value into a glob.
	if strings.ContainsAny(bound, "/\\\x00") || len(bound) > qodertranscript.MaxStringBytes {
		return unknown
	}
	var candidates []qoderCandidate
	rootsByID := map[string]string{}
	count, total, entries := 0, 0, 0
	root := filepath.Join(d.Home, ".qoder", "projects")
	err := qoderEntries(ctx, root, func(project os.DirEntry) error {
		entries++
		if entries > 1024 {
			return qodertranscript.ErrLimit
		}
		if !project.IsDir() {
			return nil
		}
		return qoderEntries(ctx, filepath.Join(root, project.Name()), func(file os.DirEntry) error {
			entries++
			if entries > 1024 {
				return qodertranscript.ErrLimit
			}
			if file.IsDir() || file.Type()&os.ModeSymlink != 0 || !strings.HasSuffix(file.Name(), ".jsonl") {
				return nil
			}
			count++
			if count > qoderCandidateLimit {
				return qodertranscript.ErrLimit
			}
			if bound != "" && file.Name() != bound+".jsonl" {
				return nil
			}
			path := filepath.Join(root, project.Name(), file.Name())
			state, n, err := readQoderIdentity(ctx, path, qoderReadLimit-total)
			total += n
			if err != nil || total > qoderReadLimit {
				return qodertranscript.ErrIncomplete
			}
			meta, err := state.Metadata()
			if err != nil {
				return err
			}
			root := qoderWorkspace(meta.CWD)
			if previous, ok := rootsByID[meta.SessionID]; ok && previous != root {
				return qodertranscript.ErrConflict
			}
			rootsByID[meta.SessionID] = root
			if root != cwd {
				if bound != "" && meta.SessionID == bound {
					return qodertranscript.ErrConflict
				}
				return nil
			}
			if meta.LastActivity.After(now) || file.Name() != meta.SessionID+".jsonl" {
				return qodertranscript.ErrConflict
			}
			if bound != "" && meta.SessionID != bound {
				return qodertranscript.ErrConflict
			}
			model, effort, err := state.ConfigAt(now)
			if err != nil {
				return err
			}
			if !qoderIdentityPart(model) || !qoderIdentityPart(effort) {
				return qodertranscript.ErrInvalidRecord
			}
			candidates = append(candidates, qoderCandidate{meta.SessionID, model, effort})
			return nil
		})
	})
	if err != nil || ctx.Err() != nil {
		return unknown
	}
	return selectQoderSession(candidates, bound)
}

func qoderWorkspace(path string) string {
	if canonical, err := filepath.EvalSymlinks(path); err == nil {
		path = canonical
	}
	return filepath.Clean(path)
}

func qoderIdentityPart(s string) bool {
	for _, c := range s {
		if !(c >= 'a' && c <= 'z' || c >= 'A' && c <= 'Z' || c >= '0' && c <= '9' || strings.ContainsRune("._:-", c)) {
			return false
		}
	}
	return true
}

func qoderEntries(ctx context.Context, path string, visit func(os.DirEntry) error) error {
	f, err := os.Open(path)
	if err != nil {
		return err
	}
	defer f.Close()
	for {
		if err := ctx.Err(); err != nil {
			return err
		}
		entries, err := f.ReadDir(32)
		for _, entry := range entries {
			if e := visit(entry); e != nil {
				return e
			}
		}
		if errors.Is(err, io.EOF) {
			return nil
		}
		if err != nil {
			return err
		}
	}
}

// readQoderIdentity scans a complete stat-pinned source within the identity
// budget. Sampling prefix/tail alone cannot rule out conflicting metadata in
// the middle, so an oversized source stays unknown rather than being sampled.
func readQoderIdentity(ctx context.Context, path string, remaining int) (qodertranscript.SessionState, int, error) {
	var state qodertranscript.SessionState
	f, err := os.Open(path)
	if err != nil {
		return state, 0, err
	}
	defer f.Close()
	before, err := f.Stat()
	if err != nil || !before.Mode().IsRegular() {
		return state, 0, qodertranscript.ErrIncomplete
	}
	if before.Size() > qoderSessionReadLimit || before.Size() > int64(remaining) {
		return state, 0, qodertranscript.ErrLimit
	}
	r := &qoderCountReader{ctx: ctx, r: io.LimitReader(f, before.Size())}
	sc := bufio.NewScanner(r)
	sc.Buffer(make([]byte, 4096), qoderSessionReadLimit)
	partial := false
	sc.Split(func(data []byte, atEOF bool) (int, []byte, error) {
		advance, token, err := bufio.ScanLines(data, atEOF)
		partial = atEOF && advance == len(data) && len(token) > 0 && !bytes.HasSuffix(data, []byte("\n"))
		return advance, token, err
	})
	for sc.Scan() {
		if err := ctx.Err(); err != nil {
			return state, r.n, err
		}
		if partial {
			return state, r.n, qodertranscript.ErrIncomplete
		}
		if len(bytes.TrimSpace(sc.Bytes())) == 0 {
			continue
		}
		rec, err := qodertranscript.ParseRecord(sc.Bytes())
		if err != nil {
			return state, r.n, err
		}
		if err = state.Step(rec); err != nil {
			return state, r.n, err
		}
	}
	if err := sc.Err(); err != nil {
		return state, r.n, err
	}
	after, err := os.Stat(path)
	if err != nil || !os.SameFile(before, after) || before.Size() != after.Size() || before.Size() != int64(r.n) || !before.ModTime().Equal(after.ModTime()) {
		return state, r.n, qodertranscript.ErrIncomplete
	}
	if err := ctx.Err(); err != nil {
		return state, r.n, err
	}
	return state, r.n, nil
}

type qoderCountReader struct {
	ctx context.Context
	r   io.Reader
	n   int
}

func (r *qoderCountReader) Read(p []byte) (int, error) {
	if err := r.ctx.Err(); err != nil {
		return 0, err
	}
	n, err := r.r.Read(p)
	r.n += n
	return n, err
}

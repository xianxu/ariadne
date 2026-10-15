package transcripts

import (
	"bufio"
	"context"
	"crypto/sha256"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/xianxu/ariadne/cmd/sdlc/internal/qodertranscript"
)

type qoderHarness struct{ root string }

// QoderHarness selects sessions from the recorded workspace, never the lossy
// project-directory label. root is ~/.qoder/projects in production.
func QoderHarness(root string) Harness { return qoderHarness{root: root} }
func (qoderHarness) Name() string      { return "qoder" }
func (h qoderHarness) Sources(cwds []string) Sources {
	allowed := map[string]bool{}
	for _, cwd := range cwds {
		if cwd != "" {
			allowed[normalizedPath(cwd)] = true
		}
	}
	if h.root == "" || len(allowed) == 0 {
		return Sources{}
	}
	var out Sources
	_ = filepath.WalkDir(h.root, func(path string, d os.DirEntry, err error) error {
		if err != nil || d.IsDir() || filepath.Ext(path) != ".jsonl" {
			return nil
		}
		s, err := readQoderEvidence(context.Background(), path, -1, nil)
		if err == nil && allowed[normalizedPath(s.CWD)] {
			out.Files = append(out.Files, path)
		}
		return nil
	})
	sort.Strings(out.Files)
	return out
}

const qoderRecordLimit = 250000

type countedReader struct {
	r io.Reader
	n int64
}

func (r *countedReader) Read(p []byte) (int, error) {
	n, err := r.r.Read(p)
	r.n += int64(n)
	return n, err
}

// readQoderEvidence resolves immutable metadata across the entire sampled prefix.
// The old first-metadata byte caps cannot prove identity: copied metadata or a
// workspace contradiction can occur after a large message. Keep bounded lines,
// record count and cancellation instead, and hash the same prefix we interpret.
// expected pins usage's preselected file identity and length; negative length
// selects the current stat-sized prefix for actual discovery.
func readQoderEvidence(ctx context.Context, path string, length int64, expected os.FileInfo) (SessionEvidence, error) {
	s := SessionEvidence{Harness: "qoder", File: path}
	f, err := os.Open(path)
	if err != nil {
		return s, err
	}
	defer f.Close()
	info, err := f.Stat()
	if err != nil {
		return s, err
	}
	if !info.Mode().IsRegular() {
		return s, fmt.Errorf("nonregular qoder source")
	}
	if expected != nil && (!os.SameFile(expected, info) || info.Size() < length || info.Size() == length && !info.ModTime().Equal(expected.ModTime())) {
		return s, fmt.Errorf("qoder source changed before metadata scan")
	}
	if length < 0 {
		length = info.Size()
	}
	s.Length = length
	hash := sha256.New()
	counted := &countedReader{r: io.LimitReader(contextReader{ctx: ctx, r: f}, length)}
	scanner := bufio.NewScanner(io.TeeReader(counted, hash))
	scanner.Buffer(make([]byte, 4096), UsageLineLimit)
	var state qodertranscript.SessionState
	records := 0
	for scanner.Scan() {
		if err := ctx.Err(); err != nil {
			return s, err
		}
		if strings.TrimSpace(scanner.Text()) == "" {
			continue
		}
		records++
		if records > qoderRecordLimit {
			return s, fmt.Errorf("qoder metadata record limit exceeded")
		}
		rec, err := qodertranscript.ParseRecord(scanner.Bytes())
		if err != nil {
			return s, err
		}
		// Workspace ambiguity must not erase the known usage session. Sticky state
		// retains contradictions; final Metadata decides actual admission.
		_ = state.Step(rec)
	}
	if err := scanner.Err(); err != nil {
		return s, fmt.Errorf("qoder metadata scan incomplete")
	}
	if err := ctx.Err(); err != nil {
		return s, err
	}
	if counted.n != length {
		return s, fmt.Errorf("qoder source truncated during metadata scan")
	}
	after, err := f.Stat()
	if err != nil {
		return s, err
	}
	if after.Size() < length || after.Size() == info.Size() && !after.ModTime().Equal(info.ModTime()) {
		return s, fmt.Errorf("qoder source changed during metadata scan")
	}
	s.Digest = fmt.Sprintf("%x", hash.Sum(nil))
	id, idErr := state.Identity()
	if idErr != nil {
		return s, idErr
	}
	s.ID = id
	m, err := state.Metadata()
	s.CWD, s.Started = m.CWD, m.StartedAt
	return s, err
}

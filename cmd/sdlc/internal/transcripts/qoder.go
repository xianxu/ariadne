package transcripts

import (
	"bufio"
	"context"
	"crypto/sha256"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/xianxu/ariadne/cmd/sdlc/internal/qodertranscript"
)

type qoderHarness struct{ root string }

// QoderHarness selects sessions from the recorded workspace, never the lossy
// project-directory label. root is ~/.qoder/projects in production.
func QoderHarness(root string) Harness { return qoderHarness{root: root} }
func (qoderHarness) Name() string      { return "qoder" }

// Sources uses the same strict evidence discovery as owned actuals. The empty
// root guard also prevents the generic discovery fallback from recursing.
func (h qoderHarness) Sources(ctx context.Context, cwds []string) (Sources, error) {
	if err := ctx.Err(); err != nil {
		return Sources{}, err
	}
	if h.root == "" {
		return Sources{}, nil
	}
	evidence, warnings, err := DiscoverEvidence(ctx, cwds, []Harness{h})
	if err != nil {
		return Sources{}, err
	}
	out := Sources{Warnings: warnings}
	for _, session := range evidence {
		out.Files = append(out.Files, session.File)
	}
	return out, nil
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
	s, _, err := readQoderEvidenceFor(ctx, path, length, expected, nil)
	return s, err
}

type qoderOwnership uint8

const (
	qoderOwnershipUnresolved qoderOwnership = iota
	qoderOwnershipUnrelated
	qoderOwnershipRelevant
)

// Exclusion requires positive proof across the complete sampled prefix. A
// malformed or unreadable row may conceal workspace evidence; an earlier or
// later foreign cwd cannot repair that uncertainty. Relevant evidence is sticky.
func readQoderEvidenceFor(ctx context.Context, path string, length int64, expected os.FileInfo, requested map[string]bool) (SessionEvidence, qoderOwnership, error) {
	ownership := qoderOwnershipUnresolved
	s := SessionEvidence{Harness: "qoder", File: path}
	f, err := os.Open(path)
	if err != nil {
		return s, ownership, err
	}
	defer f.Close()
	info, err := f.Stat()
	if err != nil {
		return s, ownership, err
	}
	if !info.Mode().IsRegular() {
		return s, ownership, fmt.Errorf("nonregular qoder source")
	}
	if expected != nil && (!os.SameFile(expected, info) || info.Size() < length || info.Size() == length && !info.ModTime().Equal(expected.ModTime())) {
		return s, ownership, fmt.Errorf("qoder source changed before metadata scan")
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
	var parseErr error
	workspaceSeen, invalidWorkspace := false, false
	for scanner.Scan() {
		if err := ctx.Err(); err != nil {
			return s, ownership, err
		}
		if strings.TrimSpace(scanner.Text()) == "" {
			continue
		}
		records++
		if records > qoderRecordLimit {
			return s, ownership, fmt.Errorf("qoder metadata record limit exceeded")
		}
		rec, err := qodertranscript.ParseRecord(scanner.Bytes())
		if err != nil {
			// Keep scanning within the same bounds to discover relevant evidence,
			// without pretending the unparsed row could not contain another workspace.
			if parseErr == nil {
				parseErr = err
			}
			continue
		}
		// Workspace ambiguity must not erase the known usage session. Sticky state
		// retains contradictions; final Metadata decides actual admission.
		if rec.Type == "workspace-directories" {
			workspaceSeen = true
			if len(rec.Directories) == 0 {
				invalidWorkspace = true
			}
			for _, cwd := range rec.Directories {
				if cwd == "" || !filepath.IsAbs(cwd) {
					invalidWorkspace = true
				}
				if requested[normalizedPath(cwd)] {
					ownership = qoderOwnershipRelevant
				}
			}
		}
		_ = state.Step(rec)
		// Preserve partial proof for error classification, never successful admission.
		if id, e := state.Identity(); e == nil {
			s.ID = id
		}
		if m, _ := state.Metadata(); m.CWD != "" {
			s.CWD, s.Started = m.CWD, m.StartedAt
		}
	}
	if err := scanner.Err(); err != nil {
		return s, ownership, fmt.Errorf("qoder metadata scan incomplete")
	}
	if err := ctx.Err(); err != nil {
		return s, ownership, err
	}
	if counted.n != length {
		return s, ownership, fmt.Errorf("qoder source truncated during metadata scan")
	}
	after, err := f.Stat()
	if err != nil {
		return s, ownership, err
	}
	if after.Size() < length || after.Size() == info.Size() && !after.ModTime().Equal(info.ModTime()) {
		return s, ownership, fmt.Errorf("qoder source changed during metadata scan")
	}
	// A held descriptor survives rename/replacement. Exclusion also requires
	// the discovered path to retain that descriptor's identity at the boundary.
	afterPath, err := os.Stat(path)
	if err != nil {
		return s, ownership, err
	}
	if !os.SameFile(info, afterPath) || afterPath.Size() < length || afterPath.Size() == info.Size() && !afterPath.ModTime().Equal(info.ModTime()) {
		return s, ownership, fmt.Errorf("qoder source path changed during metadata scan")
	}
	s.Digest = fmt.Sprintf("%x", hash.Sum(nil))
	if parseErr != nil {
		return s, ownership, parseErr
	}
	if invalidWorkspace {
		return s, ownership, qodertranscript.ErrIncomplete
	}
	if workspaceSeen && ownership != qoderOwnershipRelevant {
		ownership = qoderOwnershipUnrelated
	}
	id, idErr := state.Identity()
	if idErr != nil {
		return s, ownership, idErr
	}
	s.ID = id
	m, err := state.Metadata()
	if err == nil {
		s.CWD, s.Started = m.CWD, m.StartedAt
	}
	return s, ownership, err
}

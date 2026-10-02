package gitx

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"io"
	"path"
	"strconv"
	"strings"
)

// TreeFile is an immutable snapshot value; callers own Content.
type TreeFile struct {
	Path, Mode, OID string
	Content         []byte
}

// Snapshot fetches once and pins subsequent reads to one commit, not a moving ref.
func (t *TrunkFile) Snapshot() (*TrunkView, error) {
	// #290: right after a presence probe, an unchanged tracker is read at the
	// tip the probe saw — exactly what a fetch would yield — without fetching.
	// The probe's tip is used once; anything else fetches as before.
	if tip := t.probedTip; tip != "" {
		t.probedTip = ""
		if local, err := t.resolveOpt(t.trackingRef()); err == nil && local == tip {
			return t.ViewOf(tip), nil
		}
	}
	ref, err := t.refreshTip()
	if err != nil {
		return nil, err
	}
	return t.ViewOf(ref), nil
}

// Files reads a subtree with one listing and one batch request, regardless of
// cardinality. Unlike archive, cat-file cannot apply export-ignore/subst filters.
func (v *TrunkView) Files(prefix string) ([]TreeFile, error) {
	if prefix != "" && (path.IsAbs(prefix) || path.Clean(prefix) != prefix || prefix == "." || strings.HasPrefix(prefix, "../") || strings.ContainsAny(prefix, "\\\x00")) {
		return nil, errors.New("snapshot prefix must be a canonical repository-relative directory")
	}
	args := []string{"--literal-pathspecs", "ls-tree", "-r", "-z", "--full-tree", v.ref, "--"}
	if prefix != "" {
		args = append(args, prefix+"/")
	}
	out, diag, err := runGitBoundedInputContext(v.tf.operationContext(), v.tf.dir, nil, nil, snapshotOutputLimit, args...)
	if err != nil {
		return nil, fmt.Errorf("snapshot tree: %w\n%s", err, diag)
	}
	files, err := parseSnapshotTree(out)
	if err != nil || len(files) == 0 {
		return files, err
	}
	var input strings.Builder
	for _, f := range files {
		input.WriteString(f.OID)
		input.WriteByte('\n')
	}
	ctx := v.tf.operationContext()
	out, diag, err = runGitBoundedInputContext(ctx, v.tf.dir, nil, strings.NewReader(input.String()), snapshotOutputLimit, "cat-file", "--batch", "--buffer")
	if ctx.Err() != nil {
		return nil, ctx.Err()
	}
	if err != nil {
		return nil, fmt.Errorf("snapshot blobs: %w\n%s", err, diag)
	}
	return parseSnapshotBlobs(out, files)
}

func parseSnapshotTree(raw []byte) ([]TreeFile, error) {
	if len(raw) == 0 {
		return nil, nil
	}
	if raw[len(raw)-1] != 0 {
		return nil, errors.New("truncated snapshot tree")
	}
	rows := bytes.Split(raw[:len(raw)-1], []byte{0})
	if len(rows) > SnapshotEntryLimit {
		return nil, errors.New("snapshot exceeds 10001 entries")
	}
	files := make([]TreeFile, 0, len(rows))
	seen := map[string]bool{}
	for _, row := range rows {
		header, name, ok := bytes.Cut(row, []byte{'\t'})
		fields := strings.Fields(string(header))
		p := string(name)
		if !ok || len(fields) != 3 || fields[1] != "blob" || p == "" || path.IsAbs(p) || path.Clean(p) != p || strings.HasPrefix(p, "../") || seen[p] {
			return nil, errors.New("invalid or duplicate snapshot tree entry")
		}
		oid, err := parseObjectID([]byte(fields[2]))
		if err != nil {
			return nil, err
		}
		seen[p] = true
		files = append(files, TreeFile{Path: p, Mode: fields[0], OID: oid})
	}
	return files, nil
}

func parseSnapshotBlobs(raw []byte, files []TreeFile) ([]TreeFile, error) {
	r := bufio.NewReader(bytes.NewReader(raw))
	result := append([]TreeFile(nil), files...)
	for i, f := range result {
		header, err := r.ReadString('\n')
		if err != nil {
			return nil, fmt.Errorf("truncated blob header: %w", err)
		}
		parts := strings.Fields(header)
		if len(parts) != 3 || parts[0] != f.OID || parts[1] != "blob" {
			return nil, errors.New("unexpected snapshot blob identity")
		}
		size, err := strconv.ParseInt(parts[2], 10, 64)
		if err != nil || size < 0 || size > SnapshotBlobLimit {
			return nil, errors.New("invalid snapshot blob size (limit 1 MiB)")
		}
		content := make([]byte, int(size))
		if _, err := io.ReadFull(r, content); err != nil {
			return nil, fmt.Errorf("truncated snapshot blob: %w", err)
		}
		if delim, err := r.ReadByte(); err != nil || delim != '\n' {
			return nil, errors.New("missing snapshot blob delimiter")
		}
		result[i].Content = content
	}
	if _, err := r.ReadByte(); err != io.EOF {
		return nil, errors.New("trailing snapshot blob output")
	}
	return result, nil
}

// ReadBlob resolves an immutable mirror baseline through the same command context.
func (v *TrunkView) ReadBlob(oid string) ([]byte, error) {
	if _, err := parseObjectID([]byte(oid)); err != nil {
		return nil, err
	}
	out, diag, err := runGitBoundedInputContext(v.tf.operationContext(), v.tf.dir, nil, nil, SnapshotBlobLimit, "cat-file", "blob", oid)
	if err != nil {
		return nil, fmt.Errorf("read snapshot blob: %w\n%s", err, diag)
	}
	return out, nil
}

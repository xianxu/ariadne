package gitx

import (
	"bytes"
	"context"
	"fmt"
	"strings"
)

// WriteBlob stores exact bytes (no filters) in the object database.
func WriteBlob(ctx context.Context, dir string, content []byte) (string, error) {
	if len(content) > SnapshotBlobLimit {
		return "", fmt.Errorf("%w: blob", ErrOutputLimit)
	}
	out, diag, err := runGitBoundedInputContext(ctx, dir, nil, bytes.NewReader(content), diagnosticOutputLimit, "hash-object", "-w", "--no-filters", "--stdin")
	if err != nil {
		return "", fmt.Errorf("write blob: %w: %s", err, diag)
	}
	return parseObjectID(out)
}

// ReadBlob reads one exact blob. Missing and unreadable are both errors.
func ReadBlob(ctx context.Context, dir, oid string) ([]byte, error) {
	if !fullObjectID(oid) {
		return nil, fmt.Errorf("blob %q is not a full object ID", oid)
	}
	out, diag, err := runGitBoundedInputContext(ctx, dir, nil, nil, SnapshotBlobLimit, "cat-file", "blob", oid)
	if err != nil {
		return nil, fmt.Errorf("read blob %s: %w: %s", oid, err, diag)
	}
	return out, nil
}

// ObjectFormat reports the repository hash algorithm (sha1 or sha256).
func ObjectFormat(ctx context.Context, dir string) (string, error) {
	out, diag, err := runGitBoundedInputContext(ctx, dir, nil, nil, diagnosticOutputLimit, "rev-parse", "--show-object-format")
	if err != nil {
		return "", fmt.Errorf("object format: %w: %s", err, diag)
	}
	format := strings.TrimSpace(string(out))
	if format != "sha1" && format != "sha256" {
		return "", fmt.Errorf("unsupported object format %q", format)
	}
	return format, nil
}

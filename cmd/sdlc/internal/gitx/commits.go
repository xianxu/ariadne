package gitx

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"
)

const (
	commitBatchLimit       = 128
	commitObjectLimit      = 1 << 20
	commitBatchOutputLimit = 32 << 20
)

// CommitObject contains the raw bytes of one immutable commit. Callers own Content.
type CommitObject struct {
	OID     string
	Content []byte
}

// ReadCommitObjects reads a bounded batch of full canonical object IDs in request
// order, ignoring replacement refs. Every requested object must be a commit;
// errors return no partial results. Content slices share the owned batch buffer.
func ReadCommitObjects(ctx context.Context, dir string, oids []string) ([]CommitObject, error) {
	if len(oids) == 0 {
		return nil, nil
	}
	if len(oids) > commitBatchLimit {
		return nil, errors.New("commit batch exceeds 128 objects")
	}
	var input strings.Builder
	for _, oid := range oids {
		parsed, err := parseObjectID([]byte(oid))
		if err != nil || parsed != oid || oid != strings.ToLower(oid) {
			return nil, fmt.Errorf("invalid canonical Git object ID %q", oid)
		}
		input.WriteString(oid)
		input.WriteByte('\n')
	}
	out, diag, err := runGitBoundedInputContext(ctx, dir, nil, strings.NewReader(input.String()), commitBatchOutputLimit, "--no-replace-objects", "cat-file", "--batch", "--buffer")
	if err != nil {
		return nil, fmt.Errorf("read commit objects: %w\n%s", err, diag)
	}
	return parseCommitObjects(out, oids)
}

func parseCommitObjects(raw []byte, oids []string) ([]CommitObject, error) {
	if len(raw) > commitBatchOutputLimit {
		return nil, ErrOutputLimit
	}
	if len(oids) > commitBatchLimit {
		return nil, errors.New("commit batch exceeds 128 objects")
	}
	var result []CommitObject
	for _, oid := range oids {
		header, rest, ok := bytes.Cut(raw, []byte{'\n'})
		if !ok {
			return nil, errors.New("truncated commit header")
		}
		identity, fields, ok := bytes.Cut(header, []byte{' '})
		kind, sizeText, hasSize := bytes.Cut(fields, []byte{' '})
		if !ok || !hasSize || string(identity) != oid || string(kind) != "commit" {
			return nil, errors.New("unexpected commit object identity or type")
		}
		for _, digit := range sizeText {
			if digit < '0' || digit > '9' {
				return nil, errors.New("invalid commit object size")
			}
		}
		size, err := strconv.ParseUint(string(sizeText), 10, 64)
		if err != nil || size > commitObjectLimit {
			return nil, errors.New("invalid commit object size (limit 1 MiB)")
		}
		if size >= uint64(len(rest)) {
			return nil, errors.New("truncated commit object")
		}
		n := int(size)
		if rest[n] != '\n' {
			return nil, errors.New("missing commit object delimiter")
		}
		result = append(result, CommitObject{OID: oid, Content: rest[:n:n]})
		raw = rest[n+1:]
	}
	if len(raw) != 0 {
		return nil, errors.New("trailing commit object output")
	}
	return result, nil
}

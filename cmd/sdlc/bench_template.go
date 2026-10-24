package main

import (
	"archive/tar"
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"syscall"

	"github.com/xianxu/ariadne/cmd/sdlc/internal/gitx"
)

const benchTemplateArchiveLimit = 80 << 20

// writeBenchTemplate publishes a complete immutable snapshot without replacing
// an existing destination. OIDs are deliberately omitted: bytes and modes own
// this exported template, independently of the source object database.
func writeBenchTemplate(ctx context.Context, filename string, files []gitx.TreeFile) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if err := validateBenchTemplate(files); err != nil {
		return err
	}
	ordered := append([]gitx.TreeFile(nil), files...)
	sort.Slice(ordered, func(i, j int) bool { return ordered[i].Path < ordered[j].Path })
	f, err := os.CreateTemp(filepath.Dir(filename), ".bench-template-*")
	if err != nil {
		return err
	}
	defer os.Remove(f.Name())
	defer f.Close()
	w := tar.NewWriter(&benchTemplateWriter{ctx: ctx, writer: f, remaining: benchTemplateArchiveLimit})
	for _, entry := range ordered {
		if err := ctx.Err(); err != nil {
			return err
		}
		h := &tar.Header{Name: entry.Path, Mode: 0644, Size: int64(len(entry.Content)), Typeflag: tar.TypeReg, Format: tar.FormatPAX}
		if entry.Mode == "100755" {
			h.Mode = 0755
		}
		if entry.Mode == "120000" {
			h.Typeflag = tar.TypeSymlink
			h.Mode = 0777
			h.Linkname = string(entry.Content)
			h.Size = 0
		}
		if err := w.WriteHeader(h); err != nil {
			return err
		}
		if h.Typeflag == tar.TypeReg {
			if _, err := w.Write(entry.Content); err != nil {
				return err
			}
		}
	}
	if err := w.Close(); err != nil {
		return err
	}
	if err := f.Sync(); err != nil {
		return err
	}
	if err := f.Close(); err != nil {
		return err
	}
	if err := ctx.Err(); err != nil {
		return err
	}
	// A hard link is an atomic, exclusive publication; unlike Rename it cannot
	// overwrite another writer's template or follow a destination symlink.
	return os.Link(f.Name(), filename)
}

type benchTemplateWriter struct {
	ctx       context.Context
	writer    io.Writer
	remaining int
}

func (w *benchTemplateWriter) Write(p []byte) (int, error) {
	if err := w.ctx.Err(); err != nil {
		return 0, err
	}
	if len(p) > w.remaining {
		return 0, errors.New("bench template archive exceeds size bound")
	}
	n, err := w.writer.Write(p)
	w.remaining -= n
	return n, err
}

func readBenchTemplate(ctx context.Context, filename string) ([]gitx.TreeFile, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	f, err := os.OpenFile(filename, os.O_RDONLY|syscall.O_NOFOLLOW|syscall.O_NONBLOCK, 0)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	info, err := f.Stat()
	if err != nil {
		return nil, err
	}
	if !info.Mode().IsRegular() || info.Size() < 1024 || info.Size() > benchTemplateArchiveLimit || info.Size()%512 != 0 {
		return nil, errors.New("invalid or oversized bench template archive")
	}
	raw, err := io.ReadAll(io.LimitReader(f, benchTemplateArchiveLimit+1))
	if err != nil {
		return nil, err
	}
	if len(raw) > benchTemplateArchiveLimit || len(raw) < 1024 || !bytes.Equal(raw[len(raw)-1024:], make([]byte, 1024)) {
		return nil, errors.New("truncated or oversized bench template archive")
	}
	input := bytes.NewReader(raw)
	r := tar.NewReader(input)
	var files []gitx.TreeFile
	total := 0
	for {
		if err := ctx.Err(); err != nil {
			return nil, err
		}
		h, err := r.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("read bench template: %w", err)
		}
		if len(files) >= 10000 || h.Size < 0 || h.Size > 1<<20 {
			return nil, errors.New("bench template entry count or size exceeds bound")
		}
		for key := range h.PAXRecords {
			if key != "path" && key != "linkpath" {
				return nil, fmt.Errorf("unsupported template metadata %q", key)
			}
		}
		entry := gitx.TreeFile{Path: h.Name}
		switch h.Typeflag {
		case tar.TypeReg, tar.TypeRegA:
			switch h.Mode {
			case 0644:
				entry.Mode = "100644"
			case 0755:
				entry.Mode = "100755"
			default:
				return nil, errors.New("unsupported template file mode")
			}
			total += int(h.Size)
			if total > 64<<20 {
				return nil, errors.New("bench template exceeds 64 MiB tree budget")
			}
			entry.Content, err = io.ReadAll(r)
			if err != nil {
				return nil, err
			}
		case tar.TypeSymlink:
			if h.Size != 0 || h.Mode != 0777 {
				return nil, errors.New("invalid template symlink metadata")
			}
			entry.Mode = "120000"
			entry.Content = []byte(h.Linkname)
			total += len(entry.Content)
		default:
			return nil, errors.New("unsupported bench template entry kind")
		}
		files = append(files, entry)
	}
	// tar.Reader stops at the first terminator; nonzero appended archives or
	// payload must not disappear behind that terminator.
	rest, err := io.ReadAll(input)
	if err != nil {
		return nil, err
	}
	for _, b := range rest {
		if b != 0 {
			return nil, errors.New("data after template archive terminator")
		}
	}
	if err := validateBenchTemplate(files); err != nil {
		return nil, err
	}
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	return files, nil
}

func validateBenchTemplate(files []gitx.TreeFile) error {
	if len(files) > 10000 {
		return errors.New("bench template exceeds 10000 entries")
	}
	for _, f := range files {
		if len(f.Content) > 1<<20 {
			return fmt.Errorf("bench template file exceeds 1 MiB: %s", f.Path)
		}
	}
	return validateBenchTree(files)
}

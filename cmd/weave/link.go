package main

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/xianxu/ariadne/cmd/weave/internal/acquire"
	"github.com/xianxu/ariadne/cmd/weave/internal/weavefs"
	"github.com/xianxu/ariadne/pkg/layergraph"
)

func linkRepository(ctx context.Context, root, input string, out io.Writer) error {
	path, source := input, ""
	if strings.Contains(input, "://") || strings.HasPrefix(input, "github.com/") || strings.HasPrefix(input, "git@") {
		src, err := acquire.NormalizeSource(input)
		if err != nil {
			return err
		}
		path = filepath.Join("..", src.Name)
		source = src.URL
	}
	dir := path
	if !filepath.IsAbs(dir) {
		dir = filepath.Join(root, dir)
	}
	if source == "" {
		if info, err := os.Stat(dir); err != nil || !info.IsDir() {
			return fmt.Errorf("link: local base %s is not a directory", dir)
		}
		c := exec.CommandContext(ctx, "git", "-C", dir, "config", "--get", "remote.origin.url")
		raw, err := c.Output()
		if err == nil {
			src, err := acquire.NormalizeSource(strings.TrimSpace(string(raw)))
			if err != nil {
				return err
			}
			source = src.URL
		} else if ctx.Err() != nil {
			return ctx.Err()
		}
	}
	if err := acquire.Ensure(ctx, dir, source, true); err != nil {
		return err
	}
	return recordLink(weavefs.OSFS{}, root, path, source, out)
}

func recordLink(fs weavefs.FS, root, path, source string, out io.Writer) error {
	// Whitespace-delimited deps cannot represent paths containing whitespace.
	row := "substrate " + path
	if source != "" {
		row += " " + source
	}
	if len(strings.Fields(path)) != 1 || strings.ContainsAny(path, "#\r\n") {
		return fmt.Errorf("link: dependency path cannot contain whitespace or #: %q", path)
	}
	if _, err := layergraph.ParseDeps(row); err != nil {
		return err
	}
	deps := filepath.Join(root, "construct", "deps")
	content, err := fs.ReadFile(deps)
	if err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("read %s: %w", deps, err)
	}
	lines := strings.Split(string(content), "\n")
	found := false
	changed := false
	for i, line := range lines {
		body, comment, hasComment := strings.Cut(line, "#")
		fields := strings.Fields(body)
		if len(fields) < 2 || fields[0] != "substrate" || fields[1] != path {
			continue
		}
		found = true
		if len(fields) > 2 && source != "" {
			previous, e1 := acquire.NormalizeSource(fields[2])
			next, e2 := acquire.NormalizeSource(source)
			if e1 != nil || e2 != nil || previous.Identity != next.Identity {
				return fmt.Errorf("link: conflicting source already recorded for %s", path)
			}
		} else if len(fields) == 2 && source != "" {
			lines[i] = row
			if hasComment {
				lines[i] += " #" + comment
			}
			changed = true
		}
	}
	if !found {
		next := string(content)
		if next != "" && !strings.HasSuffix(next, "\n") {
			next += "\n"
		}
		next += row + "\n"
		lines = []string{next}
		changed = true
	}
	if changed {
		if err := fs.MkdirAll(filepath.Dir(deps)); err != nil {
			return err
		}
		if err := fs.WriteFile(deps, []byte(strings.Join(lines, "\n"))); err != nil {
			return err
		}
		fmt.Fprintf(out, "weave: declared substrate %s in construct/deps\n", path)
	} else {
		fmt.Fprintf(out, "weave: substrate %s already present in construct/deps\n", path)
	}
	return ensureBaseManifest(fs, root, path, out)
}

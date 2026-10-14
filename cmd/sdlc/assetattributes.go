package main

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os/exec"
	"strings"
	"time"

	"github.com/xianxu/ariadne/cmd/sdlc/internal/gitx"
)

func verifyAssetAttributes(ctx context.Context, root, destination string) error {
	return verifyAssetAttributesAt(ctx, root, destination, "")
}

// Assets are immutable byte evidence. Conservatively refuse every configured
// conversion, including text=auto and smudge-only filters, rather than execute
// a filter and assume it is deterministic. Exact Git writes bypass filters;
// worktree checkout/staging cannot, so both current and incoming attributes
// need this check before an operation materializes or stages asset paths.
func verifyAssetAttributesAt(ctx context.Context, root, destination, ref string) error {
	ctx, cancel := context.WithTimeout(commandContext(ctx), 10*time.Second)
	defer cancel()
	args := []string{"-C", root, "check-attr", "-z"}
	if ref != "" {
		args = append(args, "--source="+ref)
	}
	args = append(args, "text", "filter", "eol", "working-tree-encoding", "ident", "--", destination)
	raw, _, err := gitx.RunGitContext(ctx, args...)
	if err != nil {
		return fmt.Errorf("cannot verify Git attributes for %s: %w; repair Git attribute lookup before retrying", destination, err)
	}
	fields := bytes.Split(raw, []byte{0})
	if len(fields) != 16 || len(fields[len(fields)-1]) != 0 {
		return fmt.Errorf("invalid Git attribute response for %s: repair Git attribute lookup before retrying", destination)
	}
	textUnset := false
	refuse := func(attribute string) error {
		return fmt.Errorf("Git %s can convert asset %s: disable content conversion for this path in .gitattributes (use -text -filter -ident -working-tree-encoding !eol), then retry", attribute, destination)
	}
	for i := 0; i < 15; i += 3 {
		if string(fields[i]) != destination {
			return fmt.Errorf("Git attributes named another path instead of %s: repair Git attribute lookup before retrying", destination)
		}
		attr, value := string(fields[i+1]), string(fields[i+2])
		if attr == "text" && value == "unset" {
			textUnset = true
		}
		if value != "unset" && value != "unspecified" {
			return refuse(attr)
		}
	}
	if !textUnset {
		out, _, err := gitx.RunGitContext(ctx, "-C", root, "config", "--get", "core.autocrlf")
		var exit *exec.ExitError
		if err != nil {
			if errors.As(err, &exit) && exit.ExitCode() == 1 {
				return nil
			} // absent differs from a bare boolean (true)
			return fmt.Errorf("cannot verify Git newline conversion for %s: %w; repair core.autocrlf before retrying", destination, err)
		}
		switch strings.ToLower(strings.TrimSpace(string(out))) {
		case "false", "no", "off", "0":
		default:
			return refuse("core.autocrlf")
		}
	}
	return nil
}

package main

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"github.com/xianxu/ariadne/cmd/sdlc/internal/gitx"
)

const maxTrunkProseBytes = 1 << 20

// Test seam for the state immediately after a confirmed remote push.
var trunkPublishAfterPush = func() error { return nil }

func newTrunkCmd() *cobra.Command {
	var selected string
	cmd := &cobra.Command{Use: "trunk", Short: "Publish selected non-issue workshop prose to main"}
	pub := markMutatingCommand(&cobra.Command{
		Use: "publish --path P", Short: "Publish one dirty workshop prose file to main without an issue-branch commit",
		Long: "Publish one dirty Markdown file under workshop/parley, pensive, targets or projects, or workshop/lessons.md, to main. Issue details use sdlc issue publish; plans, atlas, code and settings travel with reviewed issue branches. A main edit to the selected path refuses. A retry after an uncertain push reconciles the same local bytes. The local file remains untouched; the command prints exact-path cleanup guidance.",
		Args: cobra.NoArgs,
		RunE: func(c *cobra.Command, _ []string) error {
			guardSpineRepo(c.ErrOrStderr())
			return runTrunkPublish(c.Context(), c.OutOrStdout(), c.ErrOrStderr(), selected)
		},
	})
	pub.Flags().StringVar(&selected, "path", "", "one repository-relative non-issue workshop Markdown path")
	cmd.AddCommand(pub)
	return cmd
}

// trunkProsePath is lexical; filesystem type and symlink checks are separate.
func trunkProsePath(p string) error {
	if p == "" || path.Clean(p) != p || strings.HasPrefix(p, "/") || strings.Contains(p, "\\") || !strings.HasSuffix(p, ".md") {
		return fmt.Errorf("ineligible trunk prose path %q", p)
	}
	for _, r := range p {
		if r < 0x20 || r == 0x7f {
			return fmt.Errorf("ineligible trunk prose path %q: control character", p)
		}
	}
	if p == "workshop/lessons.md" {
		return nil
	}
	for _, root := range []string{"workshop/parley/", "workshop/pensive/", "workshop/targets/", "workshop/projects/"} {
		if strings.HasPrefix(p, root) && len(p) > len(root) {
			return nil
		}
	}
	return fmt.Errorf("ineligible trunk prose path %q: use issue publish for details, or the issue branch for plans, atlas, code and settings", p)
}

// Presence matters: an absent file and an empty file are different states.
func prosePublishDecision(local []byte, localOK bool, base []byte, baseOK bool, main []byte, mainOK bool) publishVerdict {
	if !localOK {
		return publishNothing
	}
	if mainOK && bytes.Equal(local, main) {
		return publishNothing
	}
	if baseOK == mainOK && (!baseOK || bytes.Equal(base, main)) {
		return publishWrite
	}
	return publishRefuse
}

func proseAt(view *gitx.TrunkView, p string) ([]byte, bool, error) {
	ok, err := view.Exists(p)
	if err != nil || !ok {
		return nil, ok, err
	}
	b, err := view.Read(p)
	return b, true, err
}

func readLocalProse(root, p string) ([]byte, bool, error) {
	cur := root
	parts := strings.Split(p, "/")
	for i, part := range parts {
		cur = filepath.Join(cur, part)
		info, err := os.Lstat(cur)
		if errors.Is(err, os.ErrNotExist) {
			return nil, false, nil
		}
		if err != nil {
			return nil, false, err
		}
		if info.Mode()&os.ModeSymlink != 0 {
			return nil, false, fmt.Errorf("%s is a symlink", cur)
		}
		if i == len(parts)-1 {
			if !info.Mode().IsRegular() {
				return nil, false, fmt.Errorf("%s is not a regular file", cur)
			}
			if info.Size() > maxTrunkProseBytes {
				return nil, false, fmt.Errorf("%s exceeds the 1 MiB prose limit", p)
			}
		} else if !info.IsDir() {
			return nil, false, fmt.Errorf("%s is not a directory", cur)
		}
	}
	b, err := os.ReadFile(cur)
	if err != nil {
		return nil, false, err
	}
	if len(b) > maxTrunkProseBytes {
		return nil, false, fmt.Errorf("%s exceeds the 1 MiB prose limit", p)
	}
	return b, true, nil
}

func gitForProse(ctx context.Context, root string, args ...string) ([]byte, error) {
	out, diag, err := gitx.RunGitContext(ctx, append([]string{"-C", root}, args...)...)
	if err != nil {
		return nil, fmt.Errorf("git %s: %w: %s", strings.Join(args, " "), err, strings.TrimSpace(string(diag)))
	}
	return bytes.TrimSpace(out), nil
}

func stagedProse(ctx context.Context, root, p string) (bool, error) {
	_, _, err := gitx.RunGitContext(ctx, "-C", root, "diff", "--cached", "--quiet", "--", ":(literal)"+p)
	if err == nil {
		return false, nil
	}
	var exit *exec.ExitError
	if errors.As(err, &exit) && exit.ExitCode() == 1 {
		return true, nil
	}
	return false, fmt.Errorf("check staged path %s: %w", p, err)
}

func runTrunkPublish(ctx context.Context, stdout, stderr io.Writer, p string) error {
	if err := trunkProsePath(p); err != nil {
		return err
	}
	root, err := gitx.RepoTopLevel()
	if err != nil {
		return err
	}
	local, localOK, err := readLocalProse(root, p)
	if err != nil {
		return err
	}
	staged, err := stagedProse(ctx, root, p)
	if err != nil {
		return err
	}
	if staged {
		return fmt.Errorf("%s is staged; unstage it before trunk publication", p)
	}
	tf, err := gitx.NewTrunkFileContext(ctx, root, "origin", "main")
	if err != nil {
		return err
	}
	tf.BoundRemote(2 * time.Minute)
	view, err := tf.Snapshot()
	if err != nil {
		return err
	}
	head, err := gitForProse(ctx, root, "rev-parse", "HEAD")
	if err != nil {
		return err
	}
	baseRef, err := gitForProse(ctx, root, "merge-base", string(head), view.Ref())
	if err != nil {
		return err
	}
	base, baseOK, err := proseAt(tf.ViewOf(string(baseRef)), p)
	if err != nil {
		return err
	}
	headBytes, headOK, err := proseAt(tf.ViewOf(string(head)), p)
	if err != nil {
		return err
	}
	if headOK != baseOK || (headOK && !bytes.Equal(headBytes, base)) {
		return fmt.Errorf("%s is already committed on this branch; publish only uncommitted prose", p)
	}
	if !localOK {
		if headOK {
			return fmt.Errorf("%s is deleted locally; deletion is not supported by trunk publish", p)
		}
		fmt.Fprintf(stdout, "%s: already clean\n", p)
		return nil
	}
	if headOK && bytes.Equal(local, headBytes) {
		fmt.Fprintf(stdout, "%s: already clean\n", p)
		return nil
	}
	if err := verifyAssetAttributes(ctx, root, p); err != nil {
		return err
	}
	msg := "trunk: publish " + p
	err = tf.UpdateMany(msg, func(v *gitx.TrunkView) (gitx.TrunkWrite, error) {
		current, present, err := proseAt(v, p)
		if err != nil {
			return gitx.TrunkWrite{}, err
		}
		if prosePublishDecision(local, localOK, base, baseOK, current, present) == publishRefuse {
			return gitx.TrunkWrite{}, fmt.Errorf("%s changed on main since this branch's base; reconcile it there before retrying", p)
		}
		if err := verifyAssetAttributesAt(ctx, root, p, v.Ref()); err != nil {
			return gitx.TrunkWrite{}, err
		}
		return gitx.TrunkWrite{Write: map[string][]byte{p: local}, ExactBytes: true}, nil
	})
	if err != nil {
		return err
	}
	if err := trunkPublishAfterPush(); err != nil {
		return fmt.Errorf("%s is on main and the local copy is untouched: %w; rerun trunk publish --path %s", p, err, p)
	}
	fmt.Fprintf(stderr, "%s is on main; local copy retained\n", p)
	current, present, err := readLocalProse(root, p)
	if err != nil || !present || !bytes.Equal(current, local) {
		fmt.Fprintf(stderr, "%s changed while publishing; preserve and reconcile the local path before cleanup\n", p)
	} else if headOK {
		fmt.Fprintf(stderr, "After merging main into this branch, run %s; or run it before a merge that would otherwise conflict (this discards the local copy).\n", trunkRestoreCommand(p))
	} else {
		fmt.Fprintf(stderr, "Before merging main into this branch, run %s to remove the published untracked copy; git restore cannot remove an untracked file.\n", trunkCleanCommand(p))
	}
	fmt.Fprintln(stdout, p)
	return nil
}

func trunkRestoreCommand(p string) string {
	return "git restore -- " + trunkLiteralArg(p)
}

func trunkCleanCommand(p string) string {
	return "git clean -f -- " + trunkLiteralArg(p)
}

func trunkLiteralArg(p string) string {
	// The Git pathspec and the shell argument must both be literal. A selected
	// filename may itself contain wildcard or shell metacharacters.
	return "':(literal)" + strings.ReplaceAll(p, "'", "'\\''") + "'"
}

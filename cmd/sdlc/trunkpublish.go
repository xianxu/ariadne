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

func newTrunkCmd() *cobra.Command {
	var selected string
	cmd := &cobra.Command{Use: "trunk", Short: "Publish selected non-issue workshop prose to main"}
	pub := markMutatingCommand(&cobra.Command{
		Use: "publish --path P", Short: "Publish one dirty workshop prose file to main without an issue-branch commit",
		Long: "Publish one dirty Markdown file under workshop/parley, pensive, targets or projects, or workshop/lessons.md, to main. Issue details use sdlc issue publish; plans, atlas, code and settings travel with reviewed issue branches. A main edit to the selected path refuses. A retry after an uncertain push reconciles the same local bytes.",
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
	if err := finishPublishedProse(ctx, root, p, local, headOK, string(head)); err != nil {
		return fmt.Errorf("%s is on main, but the local copy was not cleaned: %w; rerun trunk publish --path %s", p, err, p)
	}
	fmt.Fprintf(stderr, "%s published to main\n", p)
	fmt.Fprintln(stdout, p)
	return nil
}

// The seam permits a real-Git test to interrupt after the remote push but
// before local cleanup; production always uses the same finisher.
var finishPublishedProse = finishTrunkPublish

// This ordering seam lets real filesystem tests edit the moved source between
// the atomic rename and the validation that decides whether it may be removed.
var trunkProseAfterMove = func(string, string) {}

func restoreMovedProse(original, backup string) error {
	// Link is exclusive: another writer's newly-created original must never be
	// replaced by the backup. The link keeps writes through old descriptors
	// visible at the restored name until the backup name is removed.
	if err := os.Link(backup, original); err != nil {
		return fmt.Errorf("local copy preserved at %s; restoring its name: %w", backup, err)
	}
	if err := os.Remove(backup); err != nil {
		return fmt.Errorf("local copy restored at %s; removing backup %s: %w", original, backup, err)
	}
	return nil
}

func finishTrunkPublish(ctx context.Context, root, p string, published []byte, tracked bool, pinnedHead string) error {
	head, err := gitForProse(ctx, root, "rev-parse", "HEAD")
	if err != nil {
		return err
	}
	if string(head) != pinnedHead {
		return fmt.Errorf("HEAD moved during publication; local bytes left untouched")
	}
	current, present, err := readLocalProse(root, p)
	if err != nil {
		return err
	}
	if !present {
		return nil // already cleaned after a lost response
	}
	if !bytes.Equal(current, published) {
		return fmt.Errorf("local bytes changed during publication; left untouched")
	}
	staged, err := stagedProse(ctx, root, p)
	if err != nil {
		return err
	}
	if staged {
		return fmt.Errorf("index changed during publication; left untouched")
	}
	abs := filepath.Join(root, filepath.FromSlash(p))
	placeholder, err := os.CreateTemp(filepath.Dir(abs), ".sdlc-trunk-publish-*")
	if err != nil {
		return err
	}
	backup := placeholder.Name()
	if err := placeholder.Close(); err != nil {
		os.Remove(backup)
		return err
	}
	if err := os.Rename(abs, backup); err != nil {
		os.Remove(backup)
		return err
	}
	trunkProseAfterMove(abs, backup)
	// Every check after the move has an explicit recovery path. A writer that
	// replaced the original name makes Link fail without overwriting it.
	recover := func(cause error) error {
		if restoreErr := restoreMovedProse(abs, backup); restoreErr != nil {
			return fmt.Errorf("%w; %v", cause, restoreErr)
		}
		return cause
	}
	moved, infoErr := os.Lstat(backup)
	if infoErr != nil || !moved.Mode().IsRegular() {
		return recover(fmt.Errorf("moved local copy changed type during cleanup: %v", infoErr))
	}
	movedBytes, err := os.ReadFile(backup)
	if err != nil || !bytes.Equal(movedBytes, published) {
		return recover(fmt.Errorf("local bytes changed during cleanup; left untouched: %v", err))
	}
	head, err = gitForProse(ctx, root, "rev-parse", "HEAD")
	if err != nil || string(head) != pinnedHead {
		return recover(fmt.Errorf("HEAD moved during cleanup; local copy left untouched: %v", err))
	}
	staged, err = stagedProse(ctx, root, p)
	if err != nil || staged {
		return recover(fmt.Errorf("index changed during cleanup; local copy left untouched: %v", err))
	}
	if !tracked {
		// The rename itself cleaned the selected path. A replacement at the
		// original name belongs to a concurrent writer and is never removed.
		if _, err := os.Lstat(abs); err == nil {
			return fmt.Errorf("another writer claimed %s; published copy retained at %s", p, backup)
		} else if !errors.Is(err, os.ErrNotExist) {
			return fmt.Errorf("check local path %s: %w; published copy retained at %s", p, err, backup)
		}
		fresh, readErr := os.ReadFile(backup)
		if readErr != nil || !bytes.Equal(fresh, published) {
			return fmt.Errorf("local bytes changed during cleanup; copy retained at %s", backup)
		}
		return os.Remove(backup)
	}
	baseBytes, _, err := gitx.RunGitContext(ctx, "-C", root, "show", pinnedHead+":"+p)
	if err != nil {
		return recover(fmt.Errorf("read tracked base for %s: %w", p, err))
	}
	replacement, err := os.OpenFile(abs, os.O_WRONLY|os.O_CREATE|os.O_EXCL, moved.Mode().Perm())
	if err != nil {
		return fmt.Errorf("another writer claimed %s; local copy retained at %s: %w", p, backup, err)
	}
	_, writeErr := replacement.Write(baseBytes)
	closeErr := replacement.Close()
	if writeErr != nil || closeErr != nil {
		return fmt.Errorf("restoring %s failed; published copy retained at %s: write %v, close %v", p, backup, writeErr, closeErr)
	}
	fresh, err := os.ReadFile(backup)
	if err != nil || !bytes.Equal(fresh, published) {
		return fmt.Errorf("local bytes changed during cleanup; copy retained at %s", backup)
	}
	return os.Remove(backup)
}

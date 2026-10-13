package main

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
	"github.com/xianxu/ariadne/cmd/sdlc/internal/gitx"
	"github.com/xianxu/ariadne/cmd/sdlc/internal/issue"
	"github.com/xianxu/ariadne/cmd/sdlc/internal/issueassets"
)

type issueAttachFlags struct {
	Issue                    int
	File, Caption, IssuesDir string
}

func newIssueAttachCmd() *cobra.Command {
	f := issueAttachFlags{}
	cmd := markMutatingCommand(&cobra.Command{Use: "attach --issue N --file PATH --caption TEXT", Short: "Attach bounded evidence under the repository asset policy", Args: cobra.NoArgs, RunE: func(cmd *cobra.Command, _ []string) error {
		guardSpineRepo(cmd.ErrOrStderr())
		return runIssueAttach(cmd.Context(), cmd.OutOrStdout(), f)
	}})
	cmd.Flags().IntVar(&f.Issue, "issue", 0, "issue ID")
	cmd.Flags().StringVar(&f.File, "file", "", "regular evidence file (at most 1 MiB)")
	cmd.Flags().StringVar(&f.Caption, "caption", "", "what this evidence demonstrates")
	cmd.Flags().StringVar(&f.IssuesDir, "issues-dir", envOr("WF_ISSUES_DIR", "workshop/issues"), "issue documents directory")
	return cmd
}

func runIssueAttach(ctx context.Context, out io.Writer, f issueAttachFlags) error {
	if f.Issue <= 0 || f.File == "" || strings.TrimSpace(f.Caption) == "" {
		return fmt.Errorf("issue, file and caption are required; pass --issue N --file PATH --caption TEXT")
	}
	root, err := gitx.RepoTopLevel()
	if err != nil {
		return fmt.Errorf("not in a repository; run attachment from the issue repository")
	}
	policy, err := resolveAssetPolicy(ctx, root)
	if err != nil {
		return err
	}
	if err = requireAssetPolicy(policy); err != nil {
		return err
	}
	issuePath, err := locateIssueFile(f.IssuesDir, f.Issue)
	if err != nil {
		return fmt.Errorf("%w; create or restore the issue before attaching", err)
	}
	detailRel, err := filepath.Rel(root, issuePath)
	if err != nil {
		return err
	}
	detailRel = filepath.ToSlash(detailRel)
	if err = refuseAttachmentTemps(root, path.Dir(detailRel)); err != nil {
		return err
	}
	original, err := readAssetRegular(root, filepath.FromSlash(detailRel), 1<<20)
	if err != nil {
		return err
	}
	fm, body, err := issue.Parse(string(original))
	if err != nil {
		return fmt.Errorf("invalid issue: %w; repair its frontmatter", err)
	}
	id := fmt.Sprintf("%06d", f.Issue)
	if declared, ok := issue.GetField(fm, "id"); ok && strings.Trim(declared, "\"'") != id {
		return fmt.Errorf("issue %s declares id %s; repair its frontmatter to match the filename", id, declared)
	}
	if tracked, err := repositoryTracked(ctx, root); err != nil {
		return err
	} else if tracked {
		env, err := openTrackerAt(ctx, root)
		if err != nil {
			return err
		}
		snap, err := env.repo.Snapshot()
		if err != nil {
			return err
		}
		card, err := snap.Require(id)
		if err != nil {
			return err
		}
		if err = requireOwnedToPublish(env, card); err != nil {
			return fmt.Errorf("%w; ask that owner to attach the evidence", err)
		}
		if _, err = refreshMirror(env, id, original); err != nil {
			return err
		}
	}
	sourceAbs, err := filepath.Abs(f.File)
	if err != nil {
		return err
	}
	// macOS spells the system temp root through /var, an OS-owned alias. Resolve
	// only that trusted prefix, then reject every user-controlled symlink beneath
	// it (or beneath / for other explicit source paths).
	if tempAbs, e := filepath.Abs(os.TempDir()); e == nil {
		if within, e := filepath.Rel(tempAbs, sourceAbs); e == nil && within != ".." && !strings.HasPrefix(within, ".."+string(filepath.Separator)) {
			if realTemp, e := filepath.EvalSymlinks(tempAbs); e == nil {
				sourceAbs = filepath.Join(realTemp, within)
			}
		}
	}
	source, err := readAssetRegular(string(filepath.Separator), strings.TrimPrefix(sourceAbs, string(filepath.Separator)), issueassets.MaxFileBytes)
	if err != nil {
		return fmt.Errorf("read evidence: %w; select a readable regular file", err)
	}
	ext := path.Ext(sourceAbs)
	if ext == "" {
		ext = "bin"
	}
	name, err := issueassets.Name(source, ext)
	if err != nil {
		return err
	}
	rel := path.Join("assets", id, name)
	if err := verifyAssetAttributes(ctx, root, path.Join(path.Dir(detailRel), rel)); err != nil {
		return err
	}
	files, err := readAssetFiles(root, detailRel, id)
	if err != nil {
		return err
	}
	existing, exists := files[rel]
	if exists && !bytes.Equal(existing, source) {
		return fmt.Errorf("asset collision at %s; restore its original bytes", rel)
	}
	refs, err := issueassets.ParseRefs(id, []byte(body))
	if err != nil {
		return err
	}
	linked := false
	for _, ref := range refs {
		linked = linked || ref == rel
	}
	if linked && !exists {
		return fmt.Errorf("missing linked asset %s; restore that file before retrying attachment", rel)
	}
	nextBody := body
	if !linked {
		caption := strings.NewReplacer("\\", "\\\\", "[", "\\[", "]", "\\]", "\n", " ", "\r", " ", "<", "&lt;", ">", "&gt;").Replace(f.Caption)
		mark := ""
		if strings.HasPrefix(http.DetectContentType(source), "image/") {
			mark = "!"
		}
		link := mark + "[" + caption + "](" + rel + ")"
		if _, end, ok := issue.SectionByteBounds(body, "Assets", issue.UnterminatedIsFenced); ok {
			nextBody = body[:end] + "\n\n" + link + "\n" + body[end:]
		} else {
			nextBody = strings.TrimRight(body, "\n") + "\n\n## Assets\n\n" + link + "\n"
		}
	}
	files[rel] = source
	if _, err = issueassets.BuildBundle(id, []byte(nextBody), files); err != nil {
		return err
	}
	next := []byte(issue.Compose(fm, nextBody))
	if len(next) > 1<<20 {
		return fmt.Errorf("issue body would exceed 1 MiB; shorten the body before attaching")
	}
	if linked {
		fmt.Fprintf(out, "already attached %s\n", rel)
		return nil
	}
	state, err := issueassets.PrepareAttach(rel, original, source, exists)
	if err != nil {
		return err
	}
	store := fileAttachmentStore{root: root, detailRel: detailRel, id: id}
	if err = applyAttachment(store, state, next); err != nil {
		return err
	}
	fmt.Fprintf(out, "attached %s\nupdated %s\nCommit both paths on the issue branch, or run sdlc issue publish --issue %d.\n", path.Join(path.Dir(detailRel), rel), detailRel, f.Issue)
	return nil
}

func issueAssetHelp(cmd *cobra.Command, args []string) {
	root, err := gitx.RepoTopLevel()
	if err != nil {
		fmt.Fprintln(cmd.OutOrStdout(), "\nAsset policy: unresolved (outside a repository); attach from an issue repository.")
		return
	}
	d, err := resolveAssetPolicy(cmd.Context(), root)
	if err != nil {
		fmt.Fprintf(cmd.OutOrStdout(), "\nAsset policy: invalid — %v\n", err)
		return
	}
	mode := "text-only"
	if d.Allowed {
		mode = "allowed"
	}
	fmt.Fprintf(cmd.OutOrStdout(), "\nAsset policy: %s; source: %s\nOnly the operator changes %s.\n", mode, d.Source, assetPolicyPath)
}

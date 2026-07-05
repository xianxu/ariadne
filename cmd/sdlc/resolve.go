// resolve.go — `sdlc resolve` / `sdlc open`. Read-only artifact-reference
// resolver (ariadne#144): map a symbolic ref (ariadne#11, #15 M4, pair#84) to
// the current file path(s) of the issue and its id-keyed plan/review family,
// correct after archiving and across sibling repos.
//
// Two layers (ARCH-PURE):
//   - Pure core: parseRef (the single-source ref grammar parser) + classifyFamily
//     (family classification + ordering). string→struct, no IO, unit-tested.
//   - Thin IO shell: resolveRepoDir (sibling-repo scan) + familyFiles (3-dir glob,
//     directories from vocab.Discovery — nothing hardcoded), surfaced as two cobra
//     commands. Read-only ⟹ never tagged markMutatingCommand ⟹ takes no
//     .git/sdlc.lock, which is what lets parley shell to it on a keypress.
//
// The grammar is single-sourced HERE, as the parser: parley#160 and agents shell
// to `sdlc resolve` rather than re-encoding the grammar. helptext/resolve.md
// documents it for humans; parseRef is the authority.
package main

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"

	"github.com/spf13/cobra"

	"github.com/xianxu/ariadne/cmd/sdlc/internal/gitx"
	"github.com/xianxu/ariadne/pkg/vocab"
)

// ArtifactRef is a parsed symbolic artifact reference. Repo "" means the current
// repo. GitHub marks a GitHub-inbox ref (resolved to a label, not a local path —
// resolve stays read-only and offline).
type ArtifactRef struct {
	Repo      string
	ID        int
	Milestone string
	GitHub    bool
}

var milestoneRe = regexp.MustCompile(`^M\d+[a-z]?$`)

// parseRef parses the single-sourced ref grammar (see helptext/resolve.md). Pure:
// no filesystem, no repo knowledge — a bare `#id` yields Repo "" and the caller
// substitutes the current repo.
//
//	[repo]#id [Mx]        workshop ref   (repo attaches directly to '#')
//	[repo] gh#id          github ref     (a 'gh' token immediately before '#')
//	id is 1–6 digits (zero-padded to 6 at glob time); Mx is a milestone tag.
func parseRef(raw string) (ArtifactRef, error) {
	s := strings.TrimSpace(raw)
	if strings.Count(s, "#") != 1 {
		return ArtifactRef{}, fmt.Errorf("ref %q: expected exactly one '#'", raw)
	}
	i := strings.IndexByte(s, '#')
	left, right := strings.TrimSpace(s[:i]), strings.TrimSpace(s[i+1:])

	// right: id-digits [WS milestone]
	rf := strings.Fields(right)
	if len(rf) == 0 {
		return ArtifactRef{}, fmt.Errorf("ref %q: missing id", raw)
	}
	if len(rf[0]) < 1 || len(rf[0]) > 6 {
		return ArtifactRef{}, fmt.Errorf("ref %q: id must be 1–6 digits", raw)
	}
	id, err := strconv.Atoi(rf[0])
	if err != nil || id <= 0 {
		return ArtifactRef{}, fmt.Errorf("ref %q: bad id %q", raw, rf[0])
	}
	ref := ArtifactRef{ID: id}
	if len(rf) >= 2 {
		if !milestoneRe.MatchString(rf[1]) {
			return ArtifactRef{}, fmt.Errorf("ref %q: bad milestone %q", raw, rf[1])
		}
		ref.Milestone = rf[1]
	}
	if len(rf) > 2 {
		return ArtifactRef{}, fmt.Errorf("ref %q: trailing tokens after milestone", raw)
	}

	// left: [repo] ["gh" marker]
	switch {
	case left == "gh":
		ref.GitHub = true
	case strings.HasSuffix(left, " gh"):
		ref.GitHub = true
		ref.Repo = strings.TrimSpace(strings.TrimSuffix(left, " gh"))
	default:
		ref.Repo = left
	}
	if strings.ContainsAny(ref.Repo, " \t") {
		return ArtifactRef{}, fmt.Errorf("ref %q: malformed repo token %q", raw, ref.Repo)
	}
	return ref, nil
}

// artifactKind classifies a family member by role.
type artifactKind int

const (
	kindIssue artifactKind = iota
	kindPlan
	kindReview
)

func (k artifactKind) String() string {
	switch k {
	case kindPlan:
		return "plan"
	case kindReview:
		return "review"
	default:
		return "issue"
	}
}

// Artifact is one resolved family member. Milestone is set for reviews (from
// -mX-review.md); "" for the -close-review.md and for non-reviews.
type Artifact struct {
	Kind      artifactKind
	Path      string
	Milestone string
}

var reviewMilestoneRe = regexp.MustCompile(`-m(\d+[a-z]?)-review\.md$`)

// classifyFamily classifies id NNNNNN's matched paths by filename suffix and
// returns them ordered issue → plan → reviews (reviews by milestone, close
// last). Pure: no IO. Paths not matching the id prefix are dropped defensively
// (a glob can only over-match a shorter id if globs are ever widened).
func classifyFamily(id int, paths []string) []Artifact {
	prefix := fmt.Sprintf("%06d-", id)
	var issue, plan, reviews []Artifact
	for _, p := range paths {
		base := filepath.Base(p)
		if !strings.HasPrefix(base, prefix) {
			continue
		}
		switch {
		case strings.HasSuffix(base, "-plan.md"):
			plan = append(plan, Artifact{Kind: kindPlan, Path: p})
		case strings.HasSuffix(base, "-review.md"):
			ms := ""
			if m := reviewMilestoneRe.FindStringSubmatch(base); m != nil {
				ms = "M" + m[1]
			}
			reviews = append(reviews, Artifact{Kind: kindReview, Path: p, Milestone: ms})
		default:
			issue = append(issue, Artifact{Kind: kindIssue, Path: p})
		}
	}
	sort.Slice(reviews, func(i, j int) bool {
		// milestone-tagged before close ("" sorts last); then lexically. NOTE:
		// lexical means M10 would sort before M2 — fine for the realistic M1–M9
		// range; make it numeric only if milestones ever exceed 9.
		if (reviews[i].Milestone == "") != (reviews[j].Milestone == "") {
			return reviews[j].Milestone == ""
		}
		if reviews[i].Milestone != reviews[j].Milestone {
			return reviews[i].Milestone < reviews[j].Milestone
		}
		return reviews[i].Path < reviews[j].Path
	})
	out := append(issue, plan...)
	return append(out, reviews...)
}

// ── IO shell (thin; directories come from the model, repo dirs from disk) ──

// resolveRepoDir maps a ref's repo token to an absolute repo directory. Empty
// token → curRoot. Else scan curRoot's parent for a sibling: exact basename
// match wins; else a unique case-insensitive prefix match (so `parley` →
// `parley.nvim`); ambiguity or no match errors with the candidates. IO seam
// (reads the parent dir); curRoot is injected so the match logic is unit-testable.
func resolveRepoDir(ref ArtifactRef, curRoot string) (string, error) {
	if ref.Repo == "" {
		return curRoot, nil
	}
	parent := filepath.Dir(curRoot)
	entries, err := os.ReadDir(parent)
	if err != nil {
		return "", fmt.Errorf("read sibling dir %s: %w", parent, err)
	}
	var siblings []string
	for _, e := range entries {
		if e.IsDir() {
			siblings = append(siblings, e.Name())
		}
	}
	// exact basename match wins (so `brain` beats the `brain-family` prefix sibling)
	for _, s := range siblings {
		if s == ref.Repo {
			return filepath.Join(parent, s), nil
		}
	}
	// unique case-insensitive prefix match
	var pref []string
	low := strings.ToLower(ref.Repo)
	for _, s := range siblings {
		if strings.HasPrefix(strings.ToLower(s), low) {
			pref = append(pref, s)
		}
	}
	switch len(pref) {
	case 1:
		return filepath.Join(parent, pref[0]), nil
	case 0:
		return "", fmt.Errorf("no sibling repo matches %q under %s", ref.Repo, parent)
	default:
		sort.Strings(pref)
		return "", fmt.Errorf("ambiguous repo %q: matches %s", ref.Repo, strings.Join(pref, ", "))
	}
}

// familyFiles globs id NNNNNN's artifacts across the issue home, the plans home,
// and the archive — unioned and de-duped — so resolution is correct whether the
// family is active or (partially) archived. Directories come from the injected
// Discovery; nothing is hardcoded (ARCH-DRY).
func familyFiles(repoDir string, d vocab.Discovery, id int) ([]string, error) {
	pat := fmt.Sprintf("%06d-*.md", id)
	seen := map[string]bool{}
	var out []string
	for _, sub := range []string{d.Home, d.Plans, d.Archive} {
		matches, err := filepath.Glob(filepath.Join(repoDir, sub, pat))
		if err != nil {
			return nil, fmt.Errorf("glob %s: %w", sub, err)
		}
		for _, m := range matches {
			if !seen[m] {
				seen[m] = true
				out = append(out, m)
			}
		}
	}
	sort.Strings(out)
	return out, nil
}

// filterMilestone narrows the family to the review sidecar(s) for milestone ms.
// Returns the (possibly empty) hits; the caller turns an empty result for a
// present issue into a distinct "exists but has no <Mx> review sidecar" error.
func filterMilestone(fam []Artifact, ms string) []Artifact {
	var hits []Artifact
	for _, a := range fam {
		if a.Kind == kindReview && a.Milestone == ms {
			hits = append(hits, a)
		}
	}
	return hits
}

// currentRoot returns the injected root, else the git repo top-level. The single
// place the read-only IO of "where am I" happens (tests inject a temp root).
func currentRoot(injected string) (string, error) {
	if injected != "" {
		return injected, nil
	}
	return gitx.RepoTopLevel()
}

// resolveArtifacts is the read-only engine shared by runResolve and runOpen
// (ARCH-DRY): parse the ref → resolve the repo dir → glob the family → classify
// → narrow to a milestone when asked. `root` must be non-empty. GitHub refs
// return (nil, ref, nil) — the caller labels them (no local path; offline).
func resolveArtifacts(refStr, root string) ([]Artifact, ArtifactRef, error) {
	ref, err := parseRef(refStr)
	if err != nil {
		return nil, ArtifactRef{}, err
	}
	if ref.GitHub {
		return nil, ref, nil
	}
	repoDir, err := resolveRepoDir(ref, root)
	if err != nil {
		return nil, ref, err
	}
	files, err := familyFiles(repoDir, vocab.Issue().Discovery(), ref.ID)
	if err != nil {
		return nil, ref, err
	}
	fam := classifyFamily(ref.ID, files)
	// Distinguish "id not found at all" from "id found but this milestone has no
	// review sidecar" — clearer than a single generic not-found.
	if len(fam) == 0 {
		return nil, ref, fmt.Errorf("no artifact resolves for #%d (searched %s)", ref.ID, repoDir)
	}
	if ref.Milestone != "" {
		narrowed := filterMilestone(fam, ref.Milestone)
		if len(narrowed) == 0 {
			return nil, ref, fmt.Errorf("#%d exists but has no %s review sidecar in %s", ref.ID, ref.Milestone, repoDir)
		}
		fam = narrowed
	}
	return fam, ref, nil
}

// ── command surface (`sdlc resolve`) ──

// resolveResult is the --json schema (field names are the JSON keys).
type resolveResult struct {
	Ref       string        `json:"ref"`
	Repo      string        `json:"repo"`
	ID        int           `json:"id"`
	Milestone string        `json:"milestone,omitempty"`
	GitHub    bool          `json:"github,omitempty"`
	Files     []resolveFile `json:"files"`
}

type resolveFile struct {
	Kind      string `json:"kind"`
	Path      string `json:"path"`
	Milestone string `json:"milestone,omitempty"`
}

type resolveOpts struct {
	ref    string
	root   string // current repo root; "" ⇒ gitx.RepoTopLevel()
	asJSON bool
	out    io.Writer
}

func encodeJSON(w io.Writer, v any) error {
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	return enc.Encode(v)
}

// runResolve prints the resolved family paths (or --json). Read-only: takes no
// lock. A GitHub ref is labeled, not resolved to a file.
func runResolve(o resolveOpts) error {
	root, err := currentRoot(o.root)
	if err != nil {
		return err
	}
	fam, ref, err := resolveArtifacts(o.ref, root)
	if err != nil {
		return err
	}
	if ref.GitHub {
		who := ref.Repo
		if who == "" {
			who = filepath.Base(root)
		}
		if o.asJSON {
			return encodeJSON(o.out, resolveResult{Ref: o.ref, Repo: who, ID: ref.ID, GitHub: true})
		}
		fmt.Fprintf(o.out, "github:%s#%d\n", who, ref.ID)
		return nil
	}
	if o.asJSON {
		res := resolveResult{Ref: o.ref, Repo: ref.Repo, ID: ref.ID, Milestone: ref.Milestone}
		for _, a := range fam {
			res.Files = append(res.Files, resolveFile{Kind: a.Kind.String(), Path: a.Path, Milestone: a.Milestone})
		}
		return encodeJSON(o.out, res)
	}
	for _, a := range fam {
		fmt.Fprintln(o.out, a.Path)
	}
	return nil
}

// NewResolveCmd builds `sdlc resolve <ref>`. NOT tagged markMutatingCommand, so
// it never acquires .git/sdlc.lock — lock-free by construction (ariadne#144).
func NewResolveCmd() *cobra.Command {
	var asJSON bool
	cmd := &cobra.Command{
		Use:           "resolve <ref>",
		Args:          cobra.ExactArgs(1),
		SilenceUsage:  true,
		SilenceErrors: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runResolve(resolveOpts{ref: args[0], asJSON: asJSON, out: cmd.OutOrStdout()})
		},
	}
	cmd.Flags().BoolVar(&asJSON, "json", false, "emit the structured resolution as JSON")
	return cmd
}

// ── command surface (`sdlc open`) ──

// openExec is the injectable editor seam (tests swap it). Default execs $EDITOR.
var openExec = defaultOpenExec

func defaultOpenExec(editor, path string) error {
	c := exec.Command(editor, path)
	c.Stdin, c.Stdout, c.Stderr = os.Stdin, os.Stdout, os.Stderr
	return c.Run()
}

type openOpts struct {
	ref  string
	root string
	out  io.Writer
}

// runOpen resolves a ref and opens the PRIMARY artifact in $EDITOR: the Mx
// review when a milestone is given, else the issue (classifyFamily orders it
// first). GitHub refs are labeled, not opened. Shares resolveArtifacts with
// runResolve (ARCH-DRY).
func runOpen(o openOpts) error {
	root, err := currentRoot(o.root)
	if err != nil {
		return err
	}
	fam, ref, err := resolveArtifacts(o.ref, root)
	if err != nil {
		return err
	}
	if ref.GitHub {
		who := ref.Repo
		if who == "" {
			who = filepath.Base(root)
		}
		fmt.Fprintf(o.out, "github:%s#%d (not opened — github ref)\n", who, ref.ID)
		return nil
	}
	editor := os.Getenv("EDITOR")
	if editor == "" {
		editor = "vi"
	}
	return openExec(editor, fam[0].Path)
}

// NewOpenCmd builds `sdlc open <ref>`. Read-only (no lock) like resolve.
func NewOpenCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:           "open <ref>",
		Args:          cobra.ExactArgs(1),
		SilenceUsage:  true,
		SilenceErrors: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runOpen(openOpts{ref: args[0], out: cmd.OutOrStdout()})
		},
	}
	return cmd
}

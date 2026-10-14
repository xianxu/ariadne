package main

import (
	"bytes"
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"runtime/debug"
	"strings"
	"testing"
	"time"

	"github.com/xianxu/ariadne/cmd/sdlc/internal/issueassets"
	"github.com/xianxu/ariadne/cmd/sdlc/internal/testfix"
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/text"
)

// Bound the help callback independently: a recursive callback must fail this
// subprocess, not consume the package runner's stack or prevent its cleanup.
func TestAssetContractHelpChild(t *testing.T) {
	mode := os.Getenv("ARIADNE_ASSET_HELP_CHILD")
	if mode == "" {
		t.Skip("subprocess only")
	}
	debug.SetMaxStack(256 << 10)
	cmd := NewIssueCmd()
	if mode == "rooted" {
		cmd = buildRoot()
		cmd.SetArgs([]string{"issue", "attach", "--help"})
	} else {
		cmd.SetArgs([]string{"--help"})
	}
	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&out)
	if err := cmd.ExecuteContext(context.Background()); err != nil {
		t.Fatal(err)
	}
	t.Log(out.String())
}
func TestAssetContractHelpCompletesAndExplainsPolicy(t *testing.T) {
	root, _, _ := attachRepo(t, true)
	for _, mode := range []string{"standalone", "rooted"} {
		t.Run(mode, func(t *testing.T) {
			for _, invalid := range []bool{false, true} {
				raw := `{"version":1,"policy":"allowed"}`
				if invalid {
					raw = `{"version":1,"policy":"typo"}`
				}
				if err := os.WriteFile(filepath.Join(root, assetPolicyPath), []byte(raw), 0600); err != nil {
					t.Fatal(err)
				}
				ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
				cmd := exec.CommandContext(ctx, os.Args[0], "-test.run=^TestAssetContractHelpChild$", "-test.v")
				cmd.Dir = root
				cmd.Env = append(os.Environ(), "ARIADNE_ASSET_HELP_CHILD="+mode)
				output, err := cmd.CombinedOutput()
				cancel()
				if err != nil {
					if len(output) > 3000 {
						output = output[:3000]
					}
					t.Fatalf("help failed: %v\n%s", err, output)
				}
				expected := []string{"Asset policy: allowed; source: " + assetPolicyPath}
				if invalid {
					expected = []string{"Asset policy: invalid", "operator", "repair " + assetPolicyPath}
				}
				for _, s := range expected {
					if !strings.Contains(string(output), s) {
						t.Errorf("help omitted %q: %s", s, output)
					}
				}
			}
		})
	}
}
func assetContractSnapshot(t *testing.T, root, issue string) ([]byte, string, string) {
	t.Helper()
	body, err := os.ReadFile(issue)
	if err != nil {
		t.Fatal(err)
	}
	return body, git(t, root, "ls-files", "--stage"), git(t, root, "show-ref")
}
func TestAssetContractIdentityRefusalPreservesCheckout(t *testing.T) {
	root, issue, source := attachRepo(t, true)
	raw, err := os.ReadFile(issue)
	if err != nil {
		t.Fatal(err)
	}
	raw = bytes.Replace(raw, []byte("id: 000007"), []byte("id: 000008"), 1)
	if err = os.WriteFile(issue, raw, 0644); err != nil {
		t.Fatal(err)
	}
	before, index, refs := assetContractSnapshot(t, root, issue)
	err = runIssueAttach(context.Background(), &bytes.Buffer{}, issueAttachFlags{Issue: 7, File: source, Caption: "evidence", IssuesDir: "workshop/issues"})
	if err == nil {
		t.Error("accepted filename/frontmatter issue ID disagreement")
	} else if !strings.Contains(err.Error(), "000008") || !strings.Contains(err.Error(), "repair") {
		t.Errorf("identity refusal lacks identity and remedy: %v", err)
	}
	after, afterIndex, afterRefs := assetContractSnapshot(t, root, issue)
	if !bytes.Equal(before, after) || index != afterIndex || refs != afterRefs {
		t.Error("identity refusal changed body, index or refs")
	}
	if _, err = os.Stat(filepath.Join(root, "workshop/issues/assets")); !os.IsNotExist(err) {
		t.Error("identity refusal created assets")
	}
}
func TestAssetContractCaptionCannotAddMarkdownNodes(t *testing.T) {
	for _, caption := range []string{`a ](https://evil.invalid) ![injected`, "a\\](javascript:alert)\n\n## Injected", "<script>alert(1)</script> &quot; &#93;", "a `code` **bold** [bracket]"} {
		t.Run(caption, func(t *testing.T) {
			_, issue, source := attachRepo(t, true)
			err := runIssueAttach(context.Background(), &bytes.Buffer{}, issueAttachFlags{Issue: 7, File: source, Caption: caption, IssuesDir: "workshop/issues"})
			if err != nil {
				t.Fatal(err)
			}
			body, err := os.ReadFile(issue)
			if err != nil {
				t.Fatal(err)
			}
			refs, err := issueassets.ParseRefs("000007", body)
			if err != nil || len(refs) != 1 {
				t.Fatalf("caption corrupted index: %v %v", refs, err)
			}
			count := 0
			doc := goldmark.New().Parser().Parse(text.NewReader(body))
			err = ast.Walk(doc, func(n ast.Node, enter bool) (ast.WalkStatus, error) {
				if !enter {
					return ast.WalkContinue, nil
				}
				switch n := n.(type) {
				case *ast.Link:
					count++
					if string(n.Destination) != refs[0] {
						t.Errorf("caption created link %s", n.Destination)
					}
				case *ast.Image:
					count++
					if string(n.Destination) != refs[0] {
						t.Errorf("caption created image %s", n.Destination)
					}
				case *ast.RawHTML:
					t.Error("caption injected raw HTML")
				}
				return ast.WalkContinue, nil
			})
			if err != nil {
				t.Fatal(err)
			}
			if count != 1 {
				t.Fatalf("expected one evidence node, got %d", count)
			}
		})
	}
}
func TestAssetContractCustomRootAndSourceRefusals(t *testing.T) {
	for _, scenario := range []string{"custom-root", "missing", "directory", "too-large", "empty-caption"} {
		t.Run(scenario, func(t *testing.T) {
			root, issue, source := attachRepo(t, true)
			issuesDir := "workshop/issues"
			if scenario == "custom-root" {
				issuesDir = "evidence/issues"
				if err := os.MkdirAll(filepath.Join(root, issuesDir), 0755); err != nil {
					t.Fatal(err)
				}
				next := filepath.Join(root, issuesDir, filepath.Base(issue))
				if err := os.Rename(issue, next); err != nil {
					t.Fatal(err)
				}
				issue = next
			}
			if scenario == "missing" {
				source = filepath.Join(root, "missing.png")
			}
			if scenario == "directory" {
				source = filepath.Dir(source)
			}
			if scenario == "too-large" {
				if err := os.WriteFile(source, bytes.Repeat([]byte("x"), issueassets.MaxFileBytes+1), 0600); err != nil {
					t.Fatal(err)
				}
			}
			before, index, refs := assetContractSnapshot(t, root, issue)
			caption := "evidence"
			if scenario == "empty-caption" {
				caption = " \t "
			}
			err := runIssueAttach(context.Background(), &bytes.Buffer{}, issueAttachFlags{Issue: 7, File: source, Caption: caption, IssuesDir: issuesDir})
			if scenario == "custom-root" {
				if err != nil {
					t.Fatal(err)
				}
				raw, e := os.ReadFile(issue)
				if e != nil {
					t.Fatal(e)
				}
				links, e := issueassets.ParseRefs("000007", raw)
				if e != nil || len(links) != 1 {
					t.Fatalf("custom root link: %v %v", links, e)
				}
				if _, e = os.Stat(filepath.Join(root, issuesDir, links[0])); e != nil {
					t.Fatal(e)
				}
				return
			}
			if err == nil {
				t.Fatal("invalid source or caption accepted")
			}
			remedy := "select"
			if scenario == "too-large" {
				remedy = "reduce"
			}
			if scenario == "empty-caption" {
				remedy = "--caption"
			}
			if !strings.Contains(err.Error(), remedy) {
				t.Errorf("missing recovery action %q: %v", remedy, err)
			}
			after, afterIndex, afterRefs := assetContractSnapshot(t, root, issue)
			if !bytes.Equal(before, after) || index != afterIndex || refs != afterRefs {
				t.Error("refusal changed body, index or refs")
			}
			if _, e := os.Stat(filepath.Join(root, issuesDir, "assets")); !os.IsNotExist(e) {
				t.Error("refusal created assets")
			}
		})
	}
}

// Slot movement is a branch checkout, so evidence needs no separate copy step.
// Assert the actual bytes and relative link survive together, not just the name
// of the branch selected at the destination.
func TestAssetMoveCarriesCommittedBundle(t *testing.T) {
	roots, _ := moveFixture(t)
	source, destination := roots[1], roots[2]
	data := []byte("\x89PNG\r\n\x1a\nissue evidence\x00\xff")
	name, err := issueassets.Name(data, "png")
	if err != nil {
		t.Fatal(err)
	}
	rel := "assets/000001/" + name
	asset := filepath.ToSlash(filepath.Join(filepath.Dir(procedureIssue), rel))
	body, err := os.ReadFile(filepath.Join(source, procedureIssue))
	if err != nil {
		t.Fatal(err)
	}
	body = append(body, []byte("\n## Assets\n\n![saved dialog]("+rel+")\n")...)
	procedureWrite(t, source, procedureIssue, string(body))
	procedureWrite(t, source, asset, string(data))
	testfix.Git(t, source, "add", "--", procedureIssue, asset)
	testfix.Git(t, source, "commit", "-qm", "attach issue evidence")
	selected := procedureHead(t, source)
	rest := strings.TrimSpace(testfix.Capture(t, source, "rev-parse", "main-slot1"))
	if out, err := runMoveTest(t, source, ":2", false); err != nil {
		t.Fatalf("move: %v\n%s", err, out)
	}
	assertMoved(t, source, destination, "main-slot1", rest, selected)
	gotBody, err := os.ReadFile(filepath.Join(destination, procedureIssue))
	if err != nil || !bytes.Equal(gotBody, body) {
		t.Fatalf("destination body changed: %v", err)
	}
	links, err := issueassets.ParseRefs("000001", gotBody)
	if err != nil || len(links) != 1 || links[0] != rel {
		t.Fatalf("moved body links %v: %v", links, err)
	}
	got, err := os.ReadFile(filepath.Join(destination, filepath.Dir(procedureIssue), links[0]))
	if err != nil || !bytes.Equal(got, data) {
		t.Fatalf("destination evidence differs: %q %v", got, err)
	}
	if _, err := os.Stat(filepath.Join(source, asset)); !os.IsNotExist(err) {
		t.Fatalf("source rest retained branch evidence: %v", err)
	}
	for _, root := range []string{source, destination} {
		if status := strings.TrimSpace(testfix.Capture(t, root, "status", "--porcelain=v1", "--untracked-files=all")); status != "" {
			t.Errorf("move dirtied %s: %s", root, status)
		}
	}
}

func TestAssetAttachForeignOwnerRefusesBeforeSourceRead(t *testing.T) {
	cardPath, card, detailPath, detail := seededIssue(t, "000007", "evidence")
	r := newTrackerRepo(t, map[string]string{cardPath: card}, map[string]string{detailPath: detail, assetPolicyPath: `{"version":1,"policy":"allowed"}`})
	withClaimant(t, thisSlot)
	setOwner(t, r, "000007", cardPath, otherSlot)
	detailAbs := filepath.Join(r.root, detailPath)
	body, index, refs := assetContractSnapshot(t, r.root, detailAbs)
	cardBefore := r.card(cardPath)
	// A nonexistent source proves the ownership error wins before source IO.
	err := runIssueAttach(context.Background(), &bytes.Buffer{}, issueAttachFlags{Issue: 7, File: filepath.Join(r.root, "missing.png"), Caption: "evidence", IssuesDir: "workshop/issues"})
	if err == nil || !strings.Contains(err.Error(), refuseForeign) || !strings.Contains(err.Error(), "ask that owner") {
		t.Fatalf("foreign owner refusal lacks attribution/action: %v", err)
	}
	after, afterIndex, afterRefs := assetContractSnapshot(t, r.root, detailAbs)
	if !bytes.Equal(body, after) || index != afterIndex || refs != afterRefs {
		t.Fatal("ownership refusal changed details, index or refs")
	}
	if got := r.card(cardPath); got != cardBefore {
		t.Fatal("ownership refusal changed card")
	}
	if _, err := os.Stat(filepath.Join(r.root, "workshop/issues/assets")); !os.IsNotExist(err) {
		t.Fatalf("ownership refusal wrote evidence: %v", err)
	}
}

func TestAssetSlugRenameKeepsBundle(t *testing.T) {
	root, detail, source := attachRepo(t, true)
	flags := issueAttachFlags{Issue: 7, File: source, Caption: "evidence", IssuesDir: "workshop/issues"}
	if err := runIssueAttach(context.Background(), &bytes.Buffer{}, flags); err != nil {
		t.Fatal(err)
	}
	before, _ := os.ReadFile(detail)
	renamed := filepath.Join(filepath.Dir(detail), "000007-renamed.md")
	if err := os.Rename(detail, renamed); err != nil {
		t.Fatal(err)
	}
	if err := runIssueAttach(context.Background(), &bytes.Buffer{}, flags); err != nil {
		t.Fatal(err)
	}
	after, _ := os.ReadFile(renamed)
	if !bytes.Equal(before, after) {
		t.Fatal("slug rename changed the body index")
	}
	files, err := readAssetFiles(root, "workshop/issues/000007-renamed.md", "000007")
	if err != nil || len(files) != 1 {
		t.Fatalf("bundle after rename: %v, %v", files, err)
	}
}

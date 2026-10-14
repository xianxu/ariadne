package main

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/xianxu/ariadne/cmd/sdlc/internal/issueassets"
)

func attachRepo(t *testing.T, allowed bool) (string, string, string) {
	t.Helper()
	root := t.TempDir()
	gitInit(t, root, "https://github.com/example/assets.git")
	dir := filepath.Join(root, "workshop/issues")
	if err := os.MkdirAll(dir, 0755); err != nil {
		t.Fatal(err)
	}
	issue := filepath.Join(dir, "000007-example.md")
	if err := os.WriteFile(issue, []byte("---\nid: 000007\nstatus: open\n---\n\n# Example\n\n## Problem\nEvidence.\n"), 0644); err != nil {
		t.Fatal(err)
	}
	mode := "text-only"
	if allowed {
		mode = "allowed"
	}
	if err := os.WriteFile(filepath.Join(root, "workshop/asset-policy.json"), []byte(`{"version":1,"policy":"`+mode+`"}`), 0644); err != nil {
		t.Fatal(err)
	}
	source := filepath.Join(t.TempDir(), "evidence.png")
	if err := os.WriteFile(source, []byte("\x89PNG\r\n\x1a\nsmall evidence"), 0600); err != nil {
		t.Fatal(err)
	}
	git(t, root, "add", ".")
	git(t, root, "-c", "user.name=Test", "-c", "user.email=test@example.com", "commit", "-qm", "fixture")
	t.Chdir(root)
	return root, issue, source
}

func TestIssueAttachPolicyRefusesBeforeReadingSource(t *testing.T) {
	root, issue, _ := attachRepo(t, false)
	before, _ := os.ReadFile(issue)
	cmd := buildRoot()
	cmd.SetArgs([]string{"issue", "attach", "--issue", "7", "--file", "missing.png", "--caption", "evidence"})
	cmd.SetOut(&bytes.Buffer{})
	cmd.SetErr(&bytes.Buffer{})
	err := cmd.ExecuteContext(context.Background())
	if err == nil || !strings.Contains(err.Error(), "text-only") || !strings.Contains(err.Error(), "operator") {
		t.Fatalf("policy refusal: %v", err)
	}
	after, _ := os.ReadFile(issue)
	if !bytes.Equal(before, after) {
		t.Fatal("refusal changed body")
	}
	if _, err := os.Stat(filepath.Join(root, "workshop/issues/assets")); !os.IsNotExist(err) {
		t.Fatal("refusal created assets")
	}
}

func TestIssueAttachCopiesLinksAndRetryIsNoop(t *testing.T) {
	root, issue, source := attachRepo(t, true)
	run := func() error {
		cmd := buildRoot()
		cmd.SetArgs([]string{"issue", "attach", "--issue", "7", "--file", source, "--caption", "a [crop]"})
		cmd.SetOut(&bytes.Buffer{})
		cmd.SetErr(&bytes.Buffer{})
		return cmd.ExecuteContext(context.Background())
	}
	if err := run(); err != nil {
		t.Fatal(err)
	}
	first, _ := os.ReadFile(issue)
	if !strings.Contains(string(first), "assets/000007/") {
		t.Fatalf("missing link: %s", first)
	}
	files, err := os.ReadDir(filepath.Join(root, "workshop/issues/assets/000007"))
	if err != nil || len(files) != 1 {
		t.Fatalf("assets: %v %v", files, err)
	}
	got, err := os.ReadFile(filepath.Join(root, "workshop/issues/assets/000007", files[0].Name()))
	want, _ := os.ReadFile(source)
	if err != nil || !bytes.Equal(got, want) {
		t.Fatal("copied bytes differ")
	}
	if err := run(); err != nil {
		t.Fatal(err)
	}
	second, _ := os.ReadFile(issue)
	if !bytes.Equal(first, second) {
		t.Fatal("retry duplicated link")
	}
}

func TestIssueAttachMissingLinkedBytesNeverReportsSuccess(t *testing.T) {
	root, _, source := attachRepo(t, true)
	f := issueAttachFlags{Issue: 7, File: source, Caption: "evidence", IssuesDir: "workshop/issues"}
	if err := runIssueAttach(context.Background(), &bytes.Buffer{}, f); err != nil {
		t.Fatal(err)
	}
	entries, err := os.ReadDir(filepath.Join(root, "workshop/issues/assets/000007"))
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Remove(filepath.Join(root, "workshop/issues/assets/000007", entries[0].Name())); err != nil {
		t.Fatal(err)
	}
	err = runIssueAttach(context.Background(), &bytes.Buffer{}, f)
	if err == nil || !strings.Contains(err.Error(), "restore") {
		t.Fatalf("dangling link accepted: %v", err)
	}
}

func TestIssueAttachRejectsSymlinkedSourceAncestor(t *testing.T) {
	root, _, source := attachRepo(t, true)
	link := filepath.Join(root, "source-link")
	if err := os.Symlink(filepath.Dir(source), link); err != nil {
		t.Fatal(err)
	}
	err := runIssueAttach(context.Background(), &bytes.Buffer{}, issueAttachFlags{Issue: 7, File: filepath.Join(link, filepath.Base(source)), Caption: "evidence", IssuesDir: "workshop/issues"})
	if err == nil || !strings.Contains(err.Error(), "symlink") {
		t.Fatalf("symlinked ancestor accepted: %v", err)
	}
}

type failingAttachmentStore struct {
	fileAttachmentStore
	failAfterLink, failCleanup bool
}

func (s failingAttachmentStore) LinkBody(old, next []byte) error {
	if s.failAfterLink {
		if err := s.fileAttachmentStore.LinkBody(old, next); err != nil {
			return err
		}
	}
	return fmt.Errorf("injected body write outcome")
}
func (s failingAttachmentStore) RemoveOwned(rel string, data []byte) error {
	if s.failCleanup {
		return fmt.Errorf("injected cleanup failure")
	}
	return s.fileAttachmentStore.RemoveOwned(rel, data)
}
func TestIssueAttachFailureRecoveryUsesOwnedBytes(t *testing.T) {
	for _, scenario := range []string{"rollback", "existing", "retained", "lost-response"} {
		t.Run(scenario, func(t *testing.T) {
			root, issueFile, sourceFile := attachRepo(t, true)
			body, _ := os.ReadFile(issueFile)
			source, _ := os.ReadFile(sourceFile)
			name, err := issueassets.Name(source, "png")
			if err != nil {
				t.Fatal(err)
			}
			dest := "assets/000007/" + name
			store := fileAttachmentStore{root: root, detailRel: "workshop/issues/000007-example.md", id: "000007"}
			preexisting := scenario == "existing"
			if preexisting {
				if _, err := store.Install(dest, source); err != nil {
					t.Fatal(err)
				}
			}
			state, err := issueassets.PrepareAttach(dest, body, source, preexisting)
			if err != nil {
				t.Fatal(err)
			}
			next := append(bytes.Clone(body), []byte("\n![evidence]("+dest+")\n")...)
			err = applyAttachment(failingAttachmentStore{store, scenario == "lost-response", scenario == "retained"}, state, next)
			_, statErr := os.Stat(filepath.Join(root, "workshop/issues", dest))
			if scenario == "lost-response" {
				if err != nil || statErr != nil {
					t.Fatalf("confirmed success lost: %v %v", err, statErr)
				}
			} else {
				if err == nil {
					t.Fatal("failure swallowed")
				}
				if scenario == "rollback" && !os.IsNotExist(statErr) {
					t.Fatal("owned bytes not rolled back")
				}
				if scenario != "rollback" && statErr != nil {
					t.Fatal("retained/pre-existing bytes removed")
				}
				current, _ := os.ReadFile(issueFile)
				if !bytes.Equal(current, body) {
					t.Fatal("failure changed body")
				}
			}
		})
	}
}

func TestIssueAttachPreflightRecoveryAndBodyLimit(t *testing.T) {
	for _, kind := range []string{"body-temp", "body-limit"} {
		t.Run(kind, func(t *testing.T) {
			root, detail, source := attachRepo(t, true)
			if kind == "body-temp" {
				if err := os.WriteFile(filepath.Join(filepath.Dir(detail), ".sdlc-attach-body-interrupted"), []byte("unowned"), 0644); err != nil {
					t.Fatal(err)
				}
			} else {
				raw, _ := os.ReadFile(detail)
				raw = append(raw, bytes.Repeat([]byte("x"), (1<<20)-len(raw)-1)...)
				if err := os.WriteFile(detail, raw, 0644); err != nil {
					t.Fatal(err)
				}
			}
			before, _ := os.ReadFile(detail)
			err := runIssueAttach(context.Background(), &bytes.Buffer{}, issueAttachFlags{Issue: 7, File: source, Caption: "evidence", IssuesDir: "workshop/issues"})
			if err == nil {
				t.Fatal("preflight accepted unsafe recovery or oversized resulting body")
			}
			after, _ := os.ReadFile(detail)
			if !bytes.Equal(before, after) {
				t.Fatal("refusal changed body")
			}
			if _, err := os.Stat(filepath.Join(root, "workshop/issues/assets")); !os.IsNotExist(err) {
				t.Fatal("refusal installed bytes")
			}
		})
	}
}

func TestIssueAttachRefusesConvertingAttributes(t *testing.T) {
	root, detail, source := attachRepo(t, true)
	if err := os.WriteFile(filepath.Join(root, ".gitattributes"), []byte("workshop/issues/assets/** text\n"), 0644); err != nil {
		t.Fatal(err)
	}
	before, _ := os.ReadFile(detail)
	err := runIssueAttach(context.Background(), &bytes.Buffer{}, issueAttachFlags{Issue: 7, File: source, Caption: "evidence", IssuesDir: "workshop/issues"})
	if err == nil || !strings.Contains(err.Error(), ".gitattributes") {
		t.Fatalf("converting attributes accepted or missing remedy: %v", err)
	}
	after, _ := os.ReadFile(detail)
	if !bytes.Equal(before, after) {
		t.Fatal("refusal changed body")
	}
	if _, err := os.Stat(filepath.Join(root, "workshop/issues/assets")); !os.IsNotExist(err) {
		t.Fatal("refusal installed bytes")
	}
}

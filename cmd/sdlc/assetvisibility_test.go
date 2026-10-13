package main

import (
	"bytes"
	"context"
	"github.com/xianxu/ariadne/cmd/sdlc/internal/gitx"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestAssetVisibilityHostResponseFailsClosed(t *testing.T) {
	for _, tc := range []struct{ name, raw, want string }{
		{"private", `{"visibility":"PRIVATE"}`, "PRIVATE"},
		{"public", `{"visibility":"PUBLIC"}`, "PUBLIC"},
		{"duplicate", `{"visibility":"PUBLIC","visibility":"PRIVATE"}`, ""},
		{"unknown", `{"visibility":"OTHER"}`, ""},
		{"extra", `{"visibility":"PRIVATE","unexpected":true}`, ""},
		{"trailing", `{"visibility":"PRIVATE"} {}`, ""},
		{"oversize", strings.Repeat(" ", 4096) + `{"visibility":"PRIVATE"}`, ""},
	} {
		t.Run(tc.name, func(t *testing.T) {
			response := filepath.Join(t.TempDir(), "response")
			if err := os.WriteFile(response, []byte(tc.raw), 0600); err != nil {
				t.Fatal(err)
			}
			t.Setenv("ASSET_VISIBILITY_RESPONSE", response)
			fakeLandingGH(t, "cat \"$ASSET_VISIBILITY_RESPONSE\"\n")
			got, err := (realGH{}).RepoVisibility(context.Background(), "https://github.com/example/repo")
			if tc.want == "" {
				if err == nil {
					t.Fatalf("accepted malformed host visibility: %q", got)
				}
			} else if err != nil || got != tc.want {
				t.Fatalf("got %q, %v", got, err)
			}
		})
	}
}

// The CLI's unqualified owner/repository form follows GH_HOST. The public
// destination and the private enterprise repository deliberately share a slug.
func fakeVisibilityHosts(t *testing.T) string {
	t.Helper()
	calls := filepath.Join(t.TempDir(), "visibility-calls")
	t.Setenv("ASSET_VISIBILITY_CALLS", calls)
	t.Setenv("GH_HOST", "private.enterprise.invalid")
	fakeLandingGH(t, `
printf '%s\n' "$3" >> "$ASSET_VISIBILITY_CALLS"
case "$3" in
 https://github.com/example/assets) printf '{"visibility":"PUBLIC"}\n' ;;
 example/assets) if [ "$GH_HOST" = private.enterprise.invalid ]; then printf '{"visibility":"PRIVATE"}\n'; else printf '{"visibility":"PUBLIC"}\n'; fi ;;
 *) exit 2 ;;
esac
`)
	return calls
}

func TestAssetVisibilityRejectsIncompleteHostIdentity(t *testing.T) {
	calls := fakeVisibilityHosts(t)
	for _, repo := range []string{"example/assets", "github.com/example/assets", "https://github.com/example/assets?x=1", "https://github.com@example.invalid/example/assets"} {
		if got, err := (realGH{}).RepoVisibility(context.Background(), repo); err == nil {
			t.Errorf("ambiguous repository %q accepted as %s", repo, got)
		}
	}
	if raw, err := os.ReadFile(calls); err == nil {
		t.Fatalf("invalid identities reached gh: %s", raw)
	} else if !os.IsNotExist(err) {
		t.Fatal(err)
	}
}

func TestAssetVisibilityHostIdentityAcrossAdmission(t *testing.T) {
	for _, admission := range []string{"attach", "publication-local-default", "publication-destination-default"} {
		t.Run(admission, func(t *testing.T) {
			root, detail, source := attachRepo(t, true)
			if err := os.Remove(filepath.Join(root, assetPolicyPath)); err != nil {
				t.Fatal(err)
			}
			git(t, root, "add", "-u")
			git(t, root, "-c", "user.name=Test", "-c", "user.email=test@example.com", "commit", "-qm", "host policy default")
			git(t, root, "config", "branch.main.remote", "origin")
			git(t, root, "config", "branch.main.merge", "refs/heads/main")
			calls := fakeVisibilityHosts(t)
			previous := ghClient
			ghClient = realGH{}
			t.Cleanup(func() { ghClient = previous })
			if admission == "publication-destination-default" {
				if err := os.WriteFile(filepath.Join(root, assetPolicyPath), []byte(`{"version":1,"policy":"allowed"}`), 0600); err != nil {
					t.Fatal(err)
				}
			}
			before, index, refs := assetContractSnapshot(t, root, detail)
			var err error
			if admission == "attach" {
				err = runIssueAttach(context.Background(), &bytes.Buffer{}, issueAttachFlags{Issue: 7, File: source, Caption: "evidence", IssuesDir: "workshop/issues"})
			} else {
				tf, e := gitx.NewTrunkFileContext(context.Background(), root, "origin", "main")
				if e != nil {
					t.Fatal(e)
				}
				pinned := tf.ViewOf(strings.TrimSpace(git(t, root, "rev-parse", "HEAD")))
				err = requireAssetPublicationPolicy(context.Background(), root, pinned)
			}
			if err == nil || !strings.Contains(err.Error(), "text-only") || !strings.Contains(err.Error(), "operator") {
				t.Errorf("private GH_HOST authorized public destination: %v", err)
			}
			raw, e := os.ReadFile(calls)
			if e != nil {
				t.Fatal(e)
			}
			if strings.TrimSpace(string(raw)) != "https://github.com/example/assets" {
				t.Errorf("host lost across admission seam: %q", raw)
			}
			after, afterIndex, afterRefs := assetContractSnapshot(t, root, detail)
			if !bytes.Equal(before, after) || index != afterIndex || refs != afterRefs {
				t.Error("refusal mutated body,index or refs")
			}
			if _, e := os.Stat(filepath.Join(root, "workshop/issues/assets")); !os.IsNotExist(e) {
				t.Error("refusal installed asset bytes")
			}
		})
	}
}

func TestAssetVisibilityFullURLAllowsRepeatedDotsInName(t *testing.T) {
	fakeLandingGH(t, "printf '{\"visibility\":\"PUBLIC\"}\\n'\n")
	got, err := (realGH{}).RepoVisibility(context.Background(), "https://github.com/example/repo..name")
	if err != nil || got != "PUBLIC" {
		t.Fatalf("valid full repository URL refused: %q %v", got, err)
	}
}

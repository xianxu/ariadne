package main

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func benchContextFixture(t *testing.T) benchPublicContext {
	t.Helper()
	base := filepath.Join(t.TempDir(), "public")
	c := benchPublicContext{Version: 1, RunID: strings.Repeat("a", 32), WorkRoot: filepath.Join(base, "work"), HubRoot: filepath.Join(base, "hub"), OriginPath: filepath.Join(base, "origin.git")}
	c.OriginIdentity = "file:" + c.OriginPath
	for _, p := range []string{c.WorkRoot, c.HubRoot, c.OriginPath} {
		if err := os.MkdirAll(p, 0700); err != nil {
			t.Fatal(err)
		}
	}
	prepareTestGit(t, c.WorkRoot, "init", "-q", "-b", "main")
	prepareTestGit(t, c.OriginPath, "init", "-q", "--bare", "-b", "main")
	prepareTestGit(t, c.WorkRoot, "remote", "add", "origin", c.OriginPath)
	prepareTestGit(t, c.WorkRoot, "config", "branch.main.remote", "origin")
	prepareTestGit(t, c.WorkRoot, "config", "branch.main.merge", "refs/heads/main")
	for _, root := range []string{c.WorkRoot, c.HubRoot} {
		writeBenchContextFixture(t, root, c)
	}
	return c
}

func writeBenchContextFixture(t *testing.T, root string, c benchPublicContext) {
	t.Helper()
	raw, err := json.Marshal(c)
	if err != nil {
		t.Fatal(err)
	}
	prepareTestWrite(t, root, benchPublicContextPath, string(raw), 0600)
}

func TestBenchContextRootOriginAndPublicationGuards(t *testing.T) {
	c := benchContextFixture(t)
	for _, root := range []string{c.WorkRoot, c.HubRoot} {
		got, err := loadBenchContext(context.Background(), root)
		if err != nil || got == nil || got.RunID != c.RunID {
			t.Fatalf("valid context: %+v %v", got, err)
		}
	}
	for _, args := range [][]string{{"claim"}, {"start-plan"}, {"change-code"}, {"close"}, {"local-check", "run"}, {"local-check", "republish"}, {"actual"}, {"bench", "status"}} {
		if err := guardBenchCommand(&c, args); err != nil {
			t.Fatalf("normal gate %v: %v", args, err)
		}
	}
	for _, args := range [][]string{{"merge"}, {"pr"}, {"push"}, {"stage"}, {"trunk", "publish"}, {"issue", "publish"}, {"local-check", "resume"}, {"local-check", "reclassify"}} {
		if err := guardBenchCommand(&c, args); err == nil {
			t.Errorf("unsafe command permitted: %v", args)
		}
	}
	prepareTestGit(t, c.WorkRoot, "remote", "set-url", "origin", "https://github.com/example/production.git")
	if _, err := loadBenchContext(context.Background(), c.WorkRoot); err == nil {
		t.Fatal("retargeted replay accepted")
	}
}

func TestBenchContextRefusesSubstitutionAndMalformedAuthority(t *testing.T) {
	for _, kind := range []string{"duplicate", "case-alias", "null", "symlink-file", "symlink-dir", "extra-remote", "alternates", "origin-alternates", "ambient-git-dir", "different-root"} {
		t.Run(kind, func(t *testing.T) {
			c := benchContextFixture(t)
			p := filepath.Join(c.WorkRoot, benchPublicContextPath)
			switch kind {
			case "duplicate":
				raw, _ := os.ReadFile(p)
				raw = append(raw[:len(raw)-1], []byte(`,"version":1}`)...)
				if err := os.WriteFile(p, raw, 0600); err != nil {
					t.Fatal(err)
				}
			case "case-alias":
				raw, _ := os.ReadFile(p)
				raw = []byte(strings.Replace(string(raw), `"version"`, `"VERSION"`, 1))
				if err := os.WriteFile(p, raw, 0600); err != nil {
					t.Fatal(err)
				}
			case "null":
				if err := os.WriteFile(p, []byte("null"), 0600); err != nil {
					t.Fatal(err)
				}
			case "symlink-file":
				target := filepath.Join(t.TempDir(), "ctx")
				raw, _ := os.ReadFile(p)
				if err := os.WriteFile(target, raw, 0600); err != nil {
					t.Fatal(err)
				}
				if err := os.Remove(p); err != nil {
					t.Fatal(err)
				}
				if err := os.Symlink(target, p); err != nil {
					t.Fatal(err)
				}
			case "symlink-dir":
				target := filepath.Join(t.TempDir(), "ctxdir")
				if err := os.Rename(filepath.Dir(p), target); err != nil {
					t.Fatal(err)
				}
				if err := os.Symlink(target, filepath.Dir(p)); err != nil {
					t.Fatal(err)
				}
			case "extra-remote":
				prepareTestGit(t, c.WorkRoot, "remote", "add", "source", "https://github.com/example/source.git")
			case "alternates":
				prepareTestWrite(t, c.WorkRoot, ".git/objects/info/alternates", t.TempDir()+"\n", 0600)
			case "origin-alternates":
				prepareTestWrite(t, c.OriginPath, "objects/info/alternates", t.TempDir()+"\n", 0600)
			case "ambient-git-dir":
				t.Setenv("GIT_DIR", filepath.Join(c.WorkRoot, ".git"))
			case "different-root":
				c.WorkRoot = c.HubRoot
				writeBenchContextFixture(t, filepath.Dir(filepath.Dir(p)), c)
			}
			if _, err := loadBenchContext(context.Background(), filepath.Dir(filepath.Dir(p))); err == nil {
				t.Fatal("invalid bench authority accepted")
			}
		})
	}
	root := t.TempDir()
	got, err := loadBenchContext(context.Background(), root)
	if err != nil || got != nil {
		t.Fatalf("ordinary repo changed: %+v %v", got, err)
	}
}

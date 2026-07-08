package plan

import (
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"testing"
	"time"

	"github.com/xianxu/ariadne/cmd/weave/internal/weavefs"
)

// Apply executes []Action against an FS, idempotently, ported from setup.sh's
// create_symlink / create_scaffold / WriteFile behaviors. It is the only
// mutating code; we test it against a real t.TempDir()-rooted OSFS (the seam
// exercised end-to-end, no mocks — ARCH: faithful over mocked).

func TestApplyWriteFile(t *testing.T) {
	root := t.TempDir()
	if err := Apply(weavefs.OSFS{}, root, []Action{
		WriteFile{Path: "AGENTS.md", Content: "BASE\n\nLOCAL\n"},
	}); err != nil {
		t.Fatalf("Apply: %v", err)
	}
	got, err := os.ReadFile(filepath.Join(root, "AGENTS.md"))
	if err != nil {
		t.Fatalf("read AGENTS.md: %v", err)
	}
	if string(got) != "BASE\n\nLOCAL\n" {
		t.Fatalf("AGENTS.md = %q, want %q", got, "BASE\n\nLOCAL\n")
	}
}

func TestApplyWriteFileCreatesParents(t *testing.T) {
	// WriteFile to a nested path creates parents (ensure_parent / mkdir -p).
	root := t.TempDir()
	if err := Apply(weavefs.OSFS{}, root, []Action{
		WriteFile{Path: "workshop/lessons.md", Content: ""},
	}); err != nil {
		t.Fatalf("Apply: %v", err)
	}
	if _, err := os.Stat(filepath.Join(root, "workshop", "lessons.md")); err != nil {
		t.Fatalf("nested file not created: %v", err)
	}
}

func TestApplyWriteFileDoesNotClobberThroughSymlink(t *testing.T) {
	// The #95 cutover hazard: until the first weave, a derivative's AGENTS.md is a
	// SYMLINK into its ancestor (nous/AGENTS.md → ../ariadne/AGENTS.md). The
	// composed AGENTS.md WriteFile lands at that slot — and os.WriteFile FOLLOWS a
	// symlink, so a naive write would clobber the ANCESTOR's source through the
	// link. applyWriteFile must remove the symlink first and write a fresh REGULAR
	// file, leaving the ancestor untouched.
	root := t.TempDir()
	upstream := t.TempDir()
	victim := filepath.Join(upstream, "AGENTS.md")
	if err := os.WriteFile(victim, []byte("ANCESTOR SOURCE — DO NOT CLOBBER"), 0o644); err != nil {
		t.Fatalf("seed victim: %v", err)
	}
	// Place a symlink at the slot, pointing at the upstream victim (the pre-cutover
	// AGENTS.md → ancestor shape).
	dst := filepath.Join(root, "AGENTS.md")
	rel, err := filepath.Rel(root, victim)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(rel, dst); err != nil {
		t.Fatalf("seed symlink: %v", err)
	}

	if err := Apply(weavefs.OSFS{}, root, []Action{
		WriteFile{Path: "AGENTS.md", Content: "COMPOSED LEAF CONTENT\n"},
	}); err != nil {
		t.Fatalf("Apply: %v", err)
	}

	// The ancestor source must be byte-for-byte untouched.
	if got, err := os.ReadFile(victim); err != nil {
		t.Fatalf("read victim: %v", err)
	} else if string(got) != "ANCESTOR SOURCE — DO NOT CLOBBER" {
		t.Fatalf("ancestor clobbered through symlink: %q", got)
	}
	// The slot is now a REGULAR file holding the composed content (not a symlink).
	if fi, err := os.Lstat(dst); err != nil {
		t.Fatalf("lstat dst: %v", err)
	} else if fi.Mode()&os.ModeSymlink != 0 {
		t.Fatalf("dst is still a symlink after WriteFile, want a regular file")
	}
	if got, err := os.ReadFile(dst); err != nil {
		t.Fatalf("read dst: %v", err)
	} else if string(got) != "COMPOSED LEAF CONTENT\n" {
		t.Fatalf("dst content = %q, want composed content", got)
	}
}

func TestApplyTouchCreatesWhenMissing(t *testing.T) {
	// Touch creates an empty file (with parents) when none exists.
	root := t.TempDir()
	if err := Apply(weavefs.OSFS{}, root, []Action{
		Touch{Path: "workshop/lessons.md"},
	}); err != nil {
		t.Fatalf("Apply: %v", err)
	}
	fi, err := os.Stat(filepath.Join(root, "workshop", "lessons.md"))
	if err != nil {
		t.Fatalf("touched file not created: %v", err)
	}
	if fi.Size() != 0 {
		t.Fatalf("touched file size = %d, want 0", fi.Size())
	}
}

func TestApplyTouchDoesNotClobberExisting(t *testing.T) {
	// The golden-diff finding: Touch must NOT overwrite an existing,
	// content-bearing file (workshop/lessons.md accumulates lessons over time).
	root := t.TempDir()
	path := filepath.Join(root, "workshop", "lessons.md")
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte("ACCUMULATED LESSONS"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := Apply(weavefs.OSFS{}, root, []Action{
		Touch{Path: "workshop/lessons.md"},
	}); err != nil {
		t.Fatalf("Apply: %v", err)
	}
	got, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read: %v", err)
	}
	if string(got) != "ACCUMULATED LESSONS" {
		t.Fatalf("Touch clobbered existing content: got %q, want ACCUMULATED LESSONS", got)
	}
}

func TestApplyMkdir(t *testing.T) {
	root := t.TempDir()
	if err := Apply(weavefs.OSFS{}, root, []Action{
		Mkdir{Path: ".claude/skills"},
	}); err != nil {
		t.Fatalf("Apply: %v", err)
	}
	info, err := os.Stat(filepath.Join(root, ".claude", "skills"))
	if err != nil || !info.IsDir() {
		t.Fatalf("scaffold dir not created: err=%v info=%v", err, info)
	}
}

// Seed is the content-tracking real-file copy ported from setup.sh's
// create_seed: created on first run, refreshed when it drifts from upstream,
// a silent no-op when identical, non-fatal when the source is missing. These
// tests restore the coverage the retired construct/scripts/test/seed-refresh.test.sh
// had (the bash extracted create_seed verbatim; here we exercise plan.applySeed
// against a real t.TempDir-rooted OSFS — the seam end-to-end, no mocks). The
// upstream source lives in a separate temp dir, mirroring an out-of-repo layer
// (Seed.Src is absolute; Seed.Dst is repo-relative, joined to root by Apply).
//
// One mapping note vs the bash: create_seed used `cp -p` (mode-preserving) and
// PRINTED seeded/updated; weavefs.FS.WriteFile writes 0o644 and Apply is silent.
// These assert CONTENT + idempotency; the executable-bit preservation (the
// load-bearing half of `cp -p` — a non-peer invokes `./bootstrap.sh` directly)
// has its own test, TestApplySeedPreservesExecutableBit.
func TestApplySeed(t *testing.T) {
	upstream := t.TempDir()
	src := filepath.Join(upstream, "bootstrap.sh")
	v1 := "#!/usr/bin/env bash\necho v1\n"
	if err := os.WriteFile(src, []byte(v1), 0o755); err != nil {
		t.Fatal(err)
	}
	seed := Seed{Src: src, Dst: "bootstrap.sh"}

	// 1. Absent target → seeded: the target is created with the source's content.
	//    (create_seed: dst absent ⇒ cp; verb=seeded.)
	root := t.TempDir()
	dst := filepath.Join(root, "bootstrap.sh")
	if err := Apply(weavefs.OSFS{}, root, []Action{seed}); err != nil {
		t.Fatalf("Apply (create): %v", err)
	}
	if got := mustRead(t, dst); got != v1 {
		t.Fatalf("after create: dst = %q, want %q", got, v1)
	}

	// 2. Re-run, identical → no-op: the target is untouched (the cmp -s guard).
	//    We detect "untouched" by stamping a known-old mtime first, then asserting
	//    it survives — a rewrite would bump it to ~now. Using a fixed sentinel
	//    mtime (not a before/after diff) sidesteps any filesystem mtime-resolution
	//    race.
	oldTime := time.Now().Add(-time.Hour).Truncate(time.Second)
	if err := os.Chtimes(dst, oldTime, oldTime); err != nil {
		t.Fatal(err)
	}
	if err := Apply(weavefs.OSFS{}, root, []Action{seed}); err != nil {
		t.Fatalf("Apply (no-op): %v", err)
	}
	if after := mustModTime(t, dst); !after.Equal(oldTime) {
		t.Fatalf("identical re-run rewrote the file (mtime %v != stamped %v); create_seed must no-op", after, oldTime)
	}
	if got := mustRead(t, dst); got != v1 {
		t.Fatalf("after no-op: dst = %q, want unchanged %q", got, v1)
	}

	// 3. Upstream drifts → updated: the target converges to the new source bytes
	//    (THE regression create_seed's #45 change fixed — a derivative stranded on
	//    a stale entrypoint catches up).
	v2 := "#!/usr/bin/env bash\necho v2-transitive\n"
	if err := os.WriteFile(src, []byte(v2), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := Apply(weavefs.OSFS{}, root, []Action{seed}); err != nil {
		t.Fatalf("Apply (refresh): %v", err)
	}
	if got := mustRead(t, dst); got != v2 {
		t.Fatalf("after refresh: dst = %q, want converged %q", got, v2)
	}

	// 4. Re-run after the update, identical again → no-op.
	if err := os.Chtimes(dst, oldTime, oldTime); err != nil {
		t.Fatal(err)
	}
	if err := Apply(weavefs.OSFS{}, root, []Action{seed}); err != nil {
		t.Fatalf("Apply (post-update no-op): %v", err)
	}
	if after := mustModTime(t, dst); !after.Equal(oldTime) {
		t.Fatalf("post-update identical re-run rewrote the file (mtime %v != stamped %v)", after, oldTime)
	}
}

func TestApplySeedMissingSourceNonFatal(t *testing.T) {
	// Missing source → non-fatal skip (create_seed's `[[ ! -f "$src" ]]` warn +
	// return 0). Apply must NOT error, and must leave any existing target intact.
	root := t.TempDir()
	dst := filepath.Join(root, "bootstrap.sh")
	if err := os.WriteFile(dst, []byte("PRE-EXISTING\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	seed := Seed{Src: filepath.Join(t.TempDir(), "nonexistent.sh"), Dst: "bootstrap.sh"}
	if err := Apply(weavefs.OSFS{}, root, []Action{seed}); err != nil {
		t.Fatalf("Apply with missing source must be non-fatal, got: %v", err)
	}
	if got := mustRead(t, dst); got != "PRE-EXISTING\n" {
		t.Fatalf("missing source clobbered the target: got %q, want PRE-EXISTING", got)
	}
}

func TestApplySeedPreservesExecutableBit(t *testing.T) {
	// The non-peer-bootstrap bug: a seeded bootstrap.sh is invoked as
	// `./bootstrap.sh` (directly), so it MUST be executable. create_seed's `cp -p`
	// preserved the source mode; applySeed observes the source mode in the IO seam
	// and chmods the seeded file to match its exec bits.

	// Executable source → seeded file is executable.
	t.Run("executable source", func(t *testing.T) {
		upstream := t.TempDir()
		src := filepath.Join(upstream, "bootstrap.sh")
		if err := os.WriteFile(src, []byte("#!/usr/bin/env bash\necho hi\n"), 0o755); err != nil {
			t.Fatal(err)
		}
		root := t.TempDir()
		dst := filepath.Join(root, "bootstrap.sh")
		if err := Apply(weavefs.OSFS{}, root, []Action{Seed{Src: src, Dst: "bootstrap.sh"}}); err != nil {
			t.Fatalf("Apply: %v", err)
		}
		fi, err := os.Stat(dst)
		if err != nil {
			t.Fatalf("stat: %v", err)
		}
		if fi.Mode().Perm()&0o111 == 0 {
			t.Fatalf("seeded file mode = %v, want executable (some +x bit set)", fi.Mode().Perm())
		}
	})

	// Non-executable source → seeded file is 0o644 (the WriteFile default; no
	// spurious +x).
	t.Run("non-executable source", func(t *testing.T) {
		upstream := t.TempDir()
		src := filepath.Join(upstream, "data.txt")
		if err := os.WriteFile(src, []byte("plain content\n"), 0o644); err != nil {
			t.Fatal(err)
		}
		root := t.TempDir()
		dst := filepath.Join(root, "data.txt")
		if err := Apply(weavefs.OSFS{}, root, []Action{Seed{Src: src, Dst: "data.txt"}}); err != nil {
			t.Fatalf("Apply: %v", err)
		}
		fi, err := os.Stat(dst)
		if err != nil {
			t.Fatalf("stat: %v", err)
		}
		if got := fi.Mode().Perm(); got != 0o644 {
			t.Fatalf("seeded file mode = %v, want 0o644", got)
		}
	})

	// Convergence: a content-identical dst seeded WITHOUT the exec bit (by an
	// older mode-blind weave) gains +x on a re-weave when the source is +x — the
	// chmod runs even on the content no-op path.
	t.Run("converges stale-mode dst", func(t *testing.T) {
		upstream := t.TempDir()
		src := filepath.Join(upstream, "bootstrap.sh")
		body := []byte("#!/usr/bin/env bash\necho hi\n")
		if err := os.WriteFile(src, body, 0o755); err != nil {
			t.Fatal(err)
		}
		root := t.TempDir()
		dst := filepath.Join(root, "bootstrap.sh")
		// Pre-seed identical content but 0o644 (the stale-mode state).
		if err := os.WriteFile(dst, body, 0o644); err != nil {
			t.Fatal(err)
		}
		if err := Apply(weavefs.OSFS{}, root, []Action{Seed{Src: src, Dst: "bootstrap.sh"}}); err != nil {
			t.Fatalf("Apply: %v", err)
		}
		fi, err := os.Stat(dst)
		if err != nil {
			t.Fatalf("stat: %v", err)
		}
		if fi.Mode().Perm()&0o111 == 0 {
			t.Fatalf("re-weave did not converge stale-mode dst: mode = %v, want +x", fi.Mode().Perm())
		}
	})
}

// mustRead reads path or fails the test.
func mustRead(t *testing.T, path string) string {
	t.Helper()
	b, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read %s: %v", path, err)
	}
	return string(b)
}

// mustModTime stats path and returns its mtime, or fails the test.
func mustModTime(t *testing.T, path string) time.Time {
	t.Helper()
	fi, err := os.Stat(path)
	if err != nil {
		t.Fatalf("stat %s: %v", path, err)
	}
	return fi.ModTime()
}

// MergeSettings is the IO half of the settings cascade: Apply reads Source
// (settings.ariadne.json) + the sibling settings.local.json off disk, runs the
// pure settingsx.Merge, and writes Target (settings.json). Ported from
// merge-settings.sh: LOCAL_FILE is <dir(target)>/settings.local.json, absent ⇒
// base-with-meta-stripped. We assert on PARSED JSON (semantic equality).

func TestApplyMergeSettingsLocalAbsent(t *testing.T) {
	root := t.TempDir()
	base := `{
		"$comment": "doc",
		"$merge_keys": ["permissions.allow"],
		"permissions": {"allow": ["A", "B"]}
	}`
	mustWrite(t, filepath.Join(root, ".claude", "settings.ariadne.json"), base)

	if err := Apply(weavefs.OSFS{}, root, []Action{
		MergeSettings{Sources: []string{filepath.Join(root, ".claude", "settings.ariadne.json")}, Target: ".claude/settings.json"},
	}); err != nil {
		t.Fatalf("Apply: %v", err)
	}
	got := readJSON(t, filepath.Join(root, ".claude", "settings.json"))
	want := map[string]any{
		"permissions": map[string]any{"allow": []any{"A", "B"}},
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("merged (local-absent):\n got=%#v\nwant=%#v", got, want)
	}
}

func TestApplyMergeSettingsWithLocal(t *testing.T) {
	root := t.TempDir()
	base := `{
		"$merge_keys": ["permissions.allow"],
		"permissions": {"allow": ["A", "B"]},
		"scalar": 1
	}`
	local := `{
		"$remove": {"permissions.allow": ["A"]},
		"permissions": {"allow": ["C"]},
		"scalar": 2
	}`
	mustWrite(t, filepath.Join(root, ".claude", "settings.ariadne.json"), base)
	// LOCAL_FILE = <dir(target)>/settings.local.json (sibling of the target).
	mustWrite(t, filepath.Join(root, ".claude", "settings.local.json"), local)

	if err := Apply(weavefs.OSFS{}, root, []Action{
		MergeSettings{Sources: []string{filepath.Join(root, ".claude", "settings.ariadne.json")}, Target: ".claude/settings.json"},
	}); err != nil {
		t.Fatalf("Apply: %v", err)
	}
	got := readJSON(t, filepath.Join(root, ".claude", "settings.json"))
	want := map[string]any{
		"permissions": map[string]any{
			// A removed before union; B kept; C appended. scalar overridden.
			"allow": []any{"B", "C"},
		},
		"scalar": float64(2),
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("merged (with local):\n got=%#v\nwant=%#v", got, want)
	}
}

func TestApplyMergeSettingsMultipleSourcesWithLocal(t *testing.T) {
	root := t.TempDir()
	base := `{
		"$merge_keys": ["permissions.allow"],
		"permissions": {"allow": ["A"]},
		"scalar": "base"
	}`
	mid := `{
		"permissions": {"allow": ["B"]},
		"scalar": "mid"
	}`
	local := `{
		"$remove": {"permissions.allow": ["A"]},
		"permissions": {"allow": ["C"]},
		"scalar": "local"
	}`
	basePath := filepath.Join(root, "base", "settings.json")
	midPath := filepath.Join(root, "mid", "settings.json")
	mustWrite(t, basePath, base)
	mustWrite(t, midPath, mid)
	mustWrite(t, filepath.Join(root, ".claude", "settings.local.json"), local)

	if err := Apply(weavefs.OSFS{}, root, []Action{
		MergeSettings{Sources: []string{basePath, midPath}, Target: ".claude/settings.json"},
	}); err != nil {
		t.Fatalf("Apply: %v", err)
	}
	got := readJSON(t, filepath.Join(root, ".claude", "settings.json"))
	want := map[string]any{
		"permissions": map[string]any{"allow": []any{"B", "C"}},
		"scalar":      "local",
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("merged (multi-source with local):\n got=%#v\nwant=%#v", got, want)
	}
}

func TestApplyMergeSettingsMissingBaseErrors(t *testing.T) {
	// merge-settings.sh errors when the base file is absent; Apply must surface it.
	root := t.TempDir()
	err := Apply(weavefs.OSFS{}, root, []Action{
		MergeSettings{Sources: []string{filepath.Join(root, ".claude", "settings.ariadne.json")}, Target: ".claude/settings.json"},
	})
	if err == nil {
		t.Fatal("Apply: expected error for missing base, got nil")
	}
}

// mustWrite writes content to path (creating parents) or fails the test.
func mustWrite(t *testing.T, path, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}

// readJSON reads + parses the file at path into a map or fails the test.
func readJSON(t *testing.T, path string) map[string]any {
	t.Helper()
	b, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read %s: %v", path, err)
	}
	var m map[string]any
	if err := json.Unmarshal(b, &m); err != nil {
		t.Fatalf("parse %s: %v\n%s", path, err, b)
	}
	return m
}

func TestApplySymlinkRelativeToTarget(t *testing.T) {
	// Symlink{Src absolute upstream, Dst repo-relative} → a RELATIVE symlink
	// from the target dir to the upstream file (ports create_symlink's
	// rel_path(src, dirname(dst))). We point Src at a real upstream file.
	root := t.TempDir()
	upstream := t.TempDir()
	srcAbs := filepath.Join(upstream, "AGENTS.md")
	if err := os.WriteFile(srcAbs, []byte("UPSTREAM"), 0o644); err != nil {
		t.Fatalf("seed upstream: %v", err)
	}
	if err := Apply(weavefs.OSFS{}, root, []Action{
		Symlink{Src: srcAbs, Dst: "AGENTS.md"},
	}); err != nil {
		t.Fatalf("Apply: %v", err)
	}
	dst := filepath.Join(root, "AGENTS.md")
	target, err := os.Readlink(dst)
	if err != nil {
		t.Fatalf("readlink: %v", err)
	}
	// link target must be RELATIVE (not absolute) — survives a repo move.
	if filepath.IsAbs(target) {
		t.Fatalf("symlink target %q is absolute, want relative", target)
	}
	// and must resolve to the upstream content.
	got, err := os.ReadFile(dst)
	if err != nil {
		t.Fatalf("read through symlink: %v", err)
	}
	if string(got) != "UPSTREAM" {
		t.Fatalf("symlink resolves to %q, want UPSTREAM", got)
	}
}

func TestApplySymlinkReplacesExistingSymlink(t *testing.T) {
	// Idempotency: a re-run with a different upstream replaces a stale symlink
	// (ports create_symlink's [[ -L ]] → rm + relink). Also: re-applying the
	// SAME link is a no-op that still leaves a correct link.
	root := t.TempDir()
	upstream := t.TempDir()
	srcA := filepath.Join(upstream, "a.md")
	srcB := filepath.Join(upstream, "b.md")
	if err := os.WriteFile(srcA, []byte("A"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(srcB, []byte("B"), 0o644); err != nil {
		t.Fatal(err)
	}
	dst := filepath.Join(root, "doc.md")

	if err := Apply(weavefs.OSFS{}, root, []Action{Symlink{Src: srcA, Dst: "doc.md"}}); err != nil {
		t.Fatalf("Apply A: %v", err)
	}
	// Re-apply pointing elsewhere — must replace, not error.
	if err := Apply(weavefs.OSFS{}, root, []Action{Symlink{Src: srcB, Dst: "doc.md"}}); err != nil {
		t.Fatalf("Apply B: %v", err)
	}
	got, err := os.ReadFile(dst)
	if err != nil {
		t.Fatalf("read: %v", err)
	}
	if string(got) != "B" {
		t.Fatalf("symlink resolves to %q after replace, want B", got)
	}
}

func TestApplySymlinkReplacesExistingRegularFile(t *testing.T) {
	// A regular file/dir occupying the slot is removed and relinked
	// (create_symlink's [[ -e ]] → rm -rf).
	root := t.TempDir()
	upstream := t.TempDir()
	src := filepath.Join(upstream, "x.md")
	if err := os.WriteFile(src, []byte("UP"), 0o644); err != nil {
		t.Fatal(err)
	}
	dst := filepath.Join(root, "x.md")
	if err := os.WriteFile(dst, []byte("STALE REGULAR FILE"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := Apply(weavefs.OSFS{}, root, []Action{Symlink{Src: src, Dst: "x.md"}}); err != nil {
		t.Fatalf("Apply: %v", err)
	}
	if fi, err := os.Lstat(dst); err != nil || fi.Mode()&os.ModeSymlink == 0 {
		t.Fatalf("dst is not a symlink after relink: err=%v mode=%v", err, fi.Mode())
	}
	got, err := os.ReadFile(dst)
	if err != nil {
		t.Fatalf("read: %v", err)
	}
	if string(got) != "UP" {
		t.Fatalf("relinked content = %q, want UP", got)
	}
}

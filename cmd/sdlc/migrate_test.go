package main

import (
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/xianxu/ariadne/cmd/sdlc/internal/testfix"
)

// TestRewriteRefs pins the #179 rewrite semantics: three rules (bare →
// source-qualified; dest-qualified → bare; everything else passes through),
// fence/span awareness, and the parseRef candidate filter. src repo "src",
// dst repo "dst".
func TestRewriteRefs(t *testing.T) {
	cases := []struct {
		name        string
		in          string
		want        string
		wantRewrite int // len(rewrites)
		wantSkip    int // len(skipped)
	}{
		{"bare qualifies", "see #12 for detail\n", "see src#12 for detail\n", 1, 0},
		{"dest-qualified normalizes", "tracked in dst#5.\n", "tracked in #5.\n", 1, 0},
		{"self-qualified passes through", "src#12 stays\n", "src#12 stays\n", 0, 0},
		{"third-repo passes through", "see third#9\n", "see third#9\n", 0, 0},
		{"gh ref reported not rewritten", "inbox gh#4 item\n", "inbox gh#4 item\n", 0, 1},
		{"qualified gh reported not rewritten", "see ariadne gh#4\n", "see ariadne gh#4\n", 0, 1},
		{"fenced block untouched", "pre #12\n```\nref #99\n```\npost\n", "pre src#12\n```\nref #99\n```\npost\n", 1, 0},
		{"unterminated fence untouched", "pre\n```\nbroken #99", "pre\n```\nbroken #99", 0, 0},
		{"single-ref span rewritten", "fixed in `#12` today\n", "fixed in `src#12` today\n", 1, 0},
		{"single-ref span with milestone rewritten", "see `dst#5 M2` there\n", "see `#5 M2` there\n", 1, 0},
		{"multi-token span skipped + reported", "run `git log --grep \"^#15\"` now\n", "run `git log --grep \"^#15\"` now\n", 0, 1},
		{"mixed-ref span skipped + reported", "pair `nous#41 #11` case\n", "pair `nous#41 #11` case\n", 0, 1},
		{"milestone form in prose", "closed #15 M4 already\n", "closed src#15 M4 already\n", 1, 0},
		{"heading no match", "## Log\n", "## Log\n", 0, 0},
		{"six-digit id", "see #000175\n", "see src#000175\n", 1, 0},
		{"seven digits no match", "hex #1234567 blob\n", "hex #1234567 blob\n", 0, 0},
		{"id zero skipped + reported", "weird #0 token\n", "weird #0 token\n", 0, 1},
		{"hex-alike is scanned and rewritten", "color #123456 pops\n", "color src#123456 pops\n", 1, 0},
		{"punctuation forms", "(#12) and #12, done\n", "(src#12) and src#12, done\n", 2, 0},
		{"multiple refs one line", "#1 then dst#5 then third#9\n", "src#1 then #5 then third#9\n", 2, 0},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			out, rewrites, skipped := rewriteRefs(tc.in, "src", "dst")
			if out != tc.want {
				t.Errorf("out:\n got %q\nwant %q", out, tc.want)
			}
			if len(rewrites) != tc.wantRewrite {
				t.Errorf("rewrites = %d (%v), want %d", len(rewrites), rewrites, tc.wantRewrite)
			}
			if len(skipped) != tc.wantSkip {
				t.Errorf("skipped = %d (%v), want %d", len(skipped), skipped, tc.wantSkip)
			}
		})
	}
}

// TestRewriteRefs_LineNumbers pins that the rewrite report carries the
// 1-indexed line of each rewrite (the operator-review surface).
func TestRewriteRefs_LineNumbers(t *testing.T) {
	in := "line one\nsee #12 here\n```\n#99\n```\nand dst#5 last\n"
	_, rewrites, _ := rewriteRefs(in, "src", "dst")
	if len(rewrites) != 2 {
		t.Fatalf("want 2 rewrites, got %v", rewrites)
	}
	if rewrites[0].Line != 2 || rewrites[0].Old != "#12" || rewrites[0].New != "src#12" {
		t.Errorf("first rewrite = %+v, want line 2 #12→src#12", rewrites[0])
	}
	if rewrites[1].Line != 6 || rewrites[1].Old != "dst#5" || rewrites[1].New != "#5" {
		t.Errorf("second rewrite = %+v, want line 6 dst#5→#5", rewrites[1])
	}
}

// TestRefScan_GrammarRoundTrip: every NEW form a rewrite produces must parse
// under parseRef — true by construction (candidates are parseRef-filtered),
// and this test pins the construction.
func TestRefScan_GrammarRoundTrip(t *testing.T) {
	in := "#12 dst#5 src#12 third#9 `#7` #15 M4 #000175 (#3)\n"
	_, rewrites, _ := rewriteRefs(in, "src", "dst")
	if len(rewrites) == 0 {
		t.Fatal("fixture produced no rewrites")
	}
	for _, r := range rewrites {
		if _, err := parseRef(strings.TrimSpace(r.New)); err != nil {
			t.Errorf("rewritten form %q does not parse: %v", r.New, err)
		}
	}
}

// migrateRepos builds a parent dir holding two sibling git repos: src/ (with
// issue 000012 and the migratable data/project/p.md) and dst/ (with issue
// 000005). Everything is committed (the tracked-clean guard demands it).
// Chdir's into src; restores on cleanup. Returns the two absolute roots.
func migrateRepos(t *testing.T) (srcRoot, dstRoot string) {
	t.Helper()
	parent := t.TempDir()
	cwd, _ := os.Getwd()
	t.Cleanup(func() { _ = os.Chdir(cwd) })

	mkRepo := func(name string) string {
		t.Helper()
		dir := testfix.Repo(t, testfix.At(parent, name))
		if err := os.MkdirAll(filepath.Join(dir, "workshop", "issues"), 0o755); err != nil {
			t.Fatal(err)
		}
		return dir
	}
	gitIn := func(dir string, args ...string) { testfix.Git(t, dir, args...) }
	write := func(dir, rel, content string) {
		t.Helper()
		p := filepath.Join(dir, rel)
		if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(p, []byte(content), 0o644); err != nil {
			t.Fatal(err)
		}
	}

	srcRoot = mkRepo("src")
	dstRoot = mkRepo("dst")

	write(srcRoot, "workshop/issues/000012-x.md",
		"---\nid: 000012\nstatus: working\n---\n# x\n")
	write(srcRoot, "data/project/p.md", migrateFixtureBody)
	gitIn(srcRoot, "add", ".")
	gitIn(srcRoot, "commit", "-q", "-m", "seed src")

	write(dstRoot, "workshop/issues/000005-y.md",
		"---\nid: 000005\nstatus: working\n---\n# y\n")
	gitIn(dstRoot, "add", ".")
	gitIn(dstRoot, "commit", "-q", "-m", "seed dst")

	if err := os.Chdir(srcRoot); err != nil {
		t.Fatal(err)
	}
	return srcRoot, dstRoot
}

const migrateFixtureBody = `# metis-v2 project

Work tracked in #12, proven in dst#5. Peer work: src#12 self-qualified.
Closed #12 M4 earlier. See also third#9 (not verified — untouched).

` + "```" + `
fenced #99 stays put
` + "```" + `

Styled ref ` + "`#12`" + ` and quoted command ` + "`git log --grep \"^#15\"`" + `.
`

// migrateFixtureWant is the fixture body after src→dst rewriting.
const migrateFixtureWant = `# metis-v2 project

Work tracked in src#12, proven in #5. Peer work: src#12 self-qualified.
Closed src#12 M4 earlier. See also third#9 (not verified — untouched).

` + "```" + `
fenced #99 stays put
` + "```" + `

Styled ref ` + "`src#12`" + ` and quoted command ` + "`git log --grep \"^#15\"`" + `.
`

func gitOut(t *testing.T, dir string, args ...string) string {
	t.Helper()
	return testfix.Git(t, dir, args...)
}

// #179 happy path: content rewritten at dest, source gone, one scoped commit
// per side, rewrite summary printed.
func TestRunMigrate_HappyPath(t *testing.T) {
	srcRoot, dstRoot := migrateRepos(t)
	var errBuf strings.Builder
	o := &migrateOpts{file: "data/project/p.md", destDir: "../dst", stderr: &errBuf}
	if err := runMigrate(o); err != nil {
		t.Fatalf("runMigrate: %v\nstderr:\n%s", err, errBuf.String())
	}
	got, err := os.ReadFile(filepath.Join(dstRoot, "data/project/p.md"))
	if err != nil {
		t.Fatalf("dest file not written: %v", err)
	}
	if string(got) != migrateFixtureWant {
		t.Errorf("dest content:\n got %q\nwant %q", got, migrateFixtureWant)
	}
	if _, err := os.Stat(filepath.Join(srcRoot, "data/project/p.md")); !os.IsNotExist(err) {
		t.Errorf("source file should be removed (stat err=%v)", err)
	}
	if s := gitOut(t, dstRoot, "log", "-1", "--pretty=%s"); !strings.Contains(s, "migrate: receive data/project/p.md from src") {
		t.Errorf("dst commit subject = %q", s)
	}
	if s := gitOut(t, srcRoot, "log", "-1", "--pretty=%s"); !strings.Contains(s, "migrate: move data/project/p.md to dst") {
		t.Errorf("src commit subject = %q", s)
	}
	for _, repo := range []string{srcRoot, dstRoot} {
		if st := gitOut(t, repo, "show", "--stat", "--pretty=", "HEAD"); strings.Count(st, "|") != 1 {
			t.Errorf("commit in %s touches more than the migrated file:\n%s", repo, st)
		}
	}
	// Summary + skipped report on stderr.
	for _, want := range []string{"#12 → src#12", "dst#5 → #5", `git log --grep "^#15"`} {
		if !strings.Contains(errBuf.String(), want) {
			t.Errorf("stderr missing %q:\n%s", want, errBuf.String())
		}
	}
}

// #179 fail-closed: a dangling ref aborts BEFORE any write.
func TestRunMigrate_DanglingRefRefuses(t *testing.T) {
	srcRoot, dstRoot := migrateRepos(t)
	p := filepath.Join(srcRoot, "data/project/p.md")
	if err := os.WriteFile(p, []byte("see #77 nowhere\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	gitOut(t, srcRoot, "add", "--", "data/project/p.md")
	gitOut(t, srcRoot, "commit", "-q", "-m", "edit")

	msg, died := expectDie(t, func() {
		_ = runMigrate(&migrateOpts{file: "data/project/p.md", destDir: "../dst", stderr: io.Discard})
	})
	if !died {
		t.Fatal("dangling ref should refuse")
	}
	if !strings.Contains(msg, "#77") {
		t.Errorf("refusal should name the dangling ref:\n%s", msg)
	}
	if _, err := os.Stat(p); err != nil {
		t.Errorf("source must be untouched on refusal: %v", err)
	}
	if _, err := os.Stat(filepath.Join(dstRoot, "data/project/p.md")); !os.IsNotExist(err) {
		t.Errorf("dest must not be written on refusal (stat err=%v)", err)
	}
}

// #179 guards: each refusal fires with nothing moved.
func TestRunMigrate_Guards(t *testing.T) {
	t.Run("dest is brain", func(t *testing.T) {
		srcRoot, dstRoot := migrateRepos(t)
		_ = srcRoot
		if err := os.MkdirAll(filepath.Join(dstRoot, ".brain"), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(dstRoot, ".brain", "config.md"), []byte("x"), 0o644); err != nil {
			t.Fatal(err)
		}
		msg, died := expectDie(t, func() {
			_ = runMigrate(&migrateOpts{file: "data/project/p.md", destDir: "../dst", stderr: io.Discard})
		})
		if !died || !strings.Contains(msg, "brain") || !strings.Contains(msg, "#171") {
			t.Errorf("brain-dest refusal wrong (died=%v):\n%s", died, msg)
		}
	})
	t.Run("dest == source repo", func(t *testing.T) {
		migrateRepos(t)
		msg, died := expectDie(t, func() {
			_ = runMigrate(&migrateOpts{file: "data/project/p.md", destDir: ".", stderr: io.Discard})
		})
		if !died || !strings.Contains(msg, "same repo") {
			t.Errorf("same-repo refusal wrong (died=%v):\n%s", died, msg)
		}
	})
	t.Run("source outside cwd repo", func(t *testing.T) {
		migrateRepos(t)
		msg, died := expectDie(t, func() {
			_ = runMigrate(&migrateOpts{file: "../dst/workshop/issues/000005-y.md", destDir: "../dst", stderr: io.Discard})
		})
		if !died || !strings.Contains(msg, "inside the current repo") {
			t.Errorf("outside-repo refusal wrong (died=%v):\n%s", died, msg)
		}
	})
	t.Run("issue-family refuses", func(t *testing.T) {
		migrateRepos(t)
		msg, died := expectDie(t, func() {
			_ = runMigrate(&migrateOpts{file: "workshop/issues/000012-x.md", destDir: "../dst", stderr: io.Discard})
		})
		if !died || !strings.Contains(msg, "renumber") {
			t.Errorf("issue-family refusal wrong (died=%v):\n%s", died, msg)
		}
	})
	t.Run("dirty dest refuses; --no-clean-check proceeds", func(t *testing.T) {
		srcRoot, dstRoot := migrateRepos(t)
		_ = srcRoot
		if err := os.WriteFile(filepath.Join(dstRoot, "wip.txt"), []byte("x"), 0o644); err != nil {
			t.Fatal(err)
		}
		msg, died := expectDie(t, func() {
			_ = runMigrate(&migrateOpts{file: "data/project/p.md", destDir: "../dst", stderr: io.Discard})
		})
		if !died || !strings.Contains(msg, "dirty") {
			t.Errorf("dirty-dest refusal wrong (died=%v):\n%s", died, msg)
		}
		var errBuf strings.Builder
		if err := runMigrate(&migrateOpts{file: "data/project/p.md", destDir: "../dst", noCleanCheck: true, stderr: &errBuf}); err != nil {
			t.Errorf("--no-clean-check should proceed: %v", err)
		}
	})
	t.Run("dest path exists refuses", func(t *testing.T) {
		srcRoot, dstRoot := migrateRepos(t)
		_ = srcRoot
		if err := os.MkdirAll(filepath.Join(dstRoot, "data/project"), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(dstRoot, "data/project/p.md"), []byte("occupied"), 0o644); err != nil {
			t.Fatal(err)
		}
		msg, died := expectDie(t, func() {
			_ = runMigrate(&migrateOpts{file: "data/project/p.md", destDir: "../dst", noCleanCheck: true, stderr: io.Discard})
		})
		if !died || !strings.Contains(msg, "already exists") {
			t.Errorf("dest-exists refusal wrong (died=%v):\n%s", died, msg)
		}
	})
	t.Run("modified source refuses", func(t *testing.T) {
		srcRoot, _ := migrateRepos(t)
		if err := os.WriteFile(filepath.Join(srcRoot, "data/project/p.md"), []byte("dirty edit\n"), 0o644); err != nil {
			t.Fatal(err)
		}
		msg, died := expectDie(t, func() {
			_ = runMigrate(&migrateOpts{file: "data/project/p.md", destDir: "../dst", stderr: io.Discard})
		})
		if !died || !strings.Contains(msg, "uncommitted") {
			t.Errorf("modified-source refusal wrong (died=%v):\n%s", died, msg)
		}
	})
	t.Run("non-sibling dest refuses at verification", func(t *testing.T) {
		srcRoot, _ := migrateRepos(t)
		// nested dest: src#12 cannot resolve from there (not a sibling of src)
		nested := filepath.Join(filepath.Dir(srcRoot), "deeper", "dst2")
		if err := os.MkdirAll(filepath.Join(nested, "workshop", "issues"), 0o755); err != nil {
			t.Fatal(err)
		}
		testfix.Repo(t, testfix.At(filepath.Dir(nested), filepath.Base(nested)))
		msg, died := expectDie(t, func() {
			_ = runMigrate(&migrateOpts{file: "data/project/p.md", destDir: nested, stderr: io.Discard})
		})
		if !died || !strings.Contains(msg, "src#12") {
			t.Errorf("non-sibling verification refusal wrong (died=%v):\n%s", died, msg)
		}
	})
}

// #179: --no-commit stages both sides and leaves committing to the operator.
func TestRunMigrate_NoCommit(t *testing.T) {
	srcRoot, dstRoot := migrateRepos(t)
	srcHead := strings.TrimSpace(gitOut(t, srcRoot, "rev-parse", "HEAD"))
	dstHead := strings.TrimSpace(gitOut(t, dstRoot, "rev-parse", "HEAD"))
	if err := runMigrate(&migrateOpts{file: "data/project/p.md", destDir: "../dst", noCommit: true, stderr: io.Discard}); err != nil {
		t.Fatalf("runMigrate --no-commit: %v", err)
	}
	if got := strings.TrimSpace(gitOut(t, dstRoot, "diff", "--cached", "--name-only")); got != "data/project/p.md" {
		t.Errorf("dst staged = %q, want the migrated file", got)
	}
	if got := strings.TrimSpace(gitOut(t, srcRoot, "diff", "--cached", "--name-only")); got != "data/project/p.md" {
		t.Errorf("src staged (removal) = %q", got)
	}
	if h := strings.TrimSpace(gitOut(t, srcRoot, "rev-parse", "HEAD")); h != srcHead {
		t.Error("src must have no new commit under --no-commit")
	}
	if h := strings.TrimSpace(gitOut(t, dstRoot, "rev-parse", "HEAD")); h != dstHead {
		t.Error("dst must have no new commit under --no-commit")
	}
}

// #179: inbound references (path-based) across sibling repos are reported
// with file:line — they don't survive a move, unlike issue refs.
func TestRunMigrate_InboundReport(t *testing.T) {
	srcRoot, _ := migrateRepos(t)
	// A third sibling repo referencing the artifact by path.
	sib := filepath.Join(filepath.Dir(srcRoot), "sib")
	if err := os.MkdirAll(sib, 0o755); err != nil {
		t.Fatal(err)
	}
	testfix.Repo(t, testfix.At(filepath.Dir(sib), filepath.Base(sib)))
	if err := os.WriteFile(filepath.Join(sib, "notes.md"), []byte("see src's data/project/p.md for the plan\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	gitOut(t, sib, "add", ".")
	gitOut(t, sib, "commit", "-q", "-m", "note")

	var errBuf strings.Builder
	if err := runMigrate(&migrateOpts{file: "data/project/p.md", destDir: "../dst", stderr: &errBuf}); err != nil {
		t.Fatalf("runMigrate: %v", err)
	}
	if !strings.Contains(errBuf.String(), "sib/notes.md:1") {
		t.Errorf("inbound report missing sib/notes.md:1:\n%s", errBuf.String())
	}
}

// #179 round-trip: canonicalization + idempotence. The first return may
// canonicalize a self-qualified ref (src#12 → #12 — a semantic no-op); the
// second round-trip must be byte-stable.
func TestRunMigrate_RoundTrip(t *testing.T) {
	srcRoot, dstRoot := migrateRepos(t)
	original := migrateFixtureBody

	roundTrip := func() {
		t.Helper()
		// src → dst (cwd is src)
		if err := os.Chdir(srcRoot); err != nil {
			t.Fatal(err)
		}
		if err := runMigrate(&migrateOpts{file: "data/project/p.md", destDir: "../dst", stderr: io.Discard}); err != nil {
			t.Fatalf("outbound: %v", err)
		}
		// dst → src (roles swap; cwd must be dst — the file's owner)
		if err := os.Chdir(dstRoot); err != nil {
			t.Fatal(err)
		}
		if err := runMigrate(&migrateOpts{file: "data/project/p.md", destDir: "../src", stderr: io.Discard}); err != nil {
			t.Fatalf("return: %v", err)
		}
	}

	roundTrip()
	once, err := os.ReadFile(filepath.Join(srcRoot, "data/project/p.md"))
	if err != nil {
		t.Fatal(err)
	}
	// The ONLY allowed delta vs the original: self-qualified src#12 → #12.
	wantOnce := strings.Replace(original, "Peer work: src#12 self-qualified.", "Peer work: #12 self-qualified.", 1)
	if string(once) != wantOnce {
		t.Errorf("first round-trip:\n got %q\nwant %q", once, wantOnce)
	}

	roundTrip()
	twice, err := os.ReadFile(filepath.Join(srcRoot, "data/project/p.md"))
	if err != nil {
		t.Fatal(err)
	}
	if string(twice) != string(once) {
		t.Errorf("second round-trip not byte-stable:\n got %q\nwant %q", twice, once)
	}
}

// #179: migrate's stderr lines must not collide with any gatesig classifier
// pattern (#172 friction attribution) — cheap insurance even though migrate
// isn't in the gate catalog (close_atlasskip_test.go precedent).
func TestMigrateLines_NoGatesigCollision(t *testing.T) {
	for _, line := range []string{
		"line 3: #12 → src#12",
		"not rewritten — line 9: `git log --grep \"^#15\"` — code span with ref-like content, not rewritten",
		"moved data/project/p.md → dst/data/project/p.md (both sides committed, scoped)",
		"--no-clean-check: proceeding into a dirty destination (staging is explicit-path)",
		"inbound-ref sweep: no references to the old path across sibling repos",
		"--no-commit: both sides staged; commit with:",
	} {
		assertNoGatesigCollision(t, line)
	}
}

// #179 regression (found by live dogfood): Go's os.Getwd prefers the $PWD
// env var — a LOGICAL, symlink-preserving path — while git rev-parse
// --show-toplevel returns the RESOLVED path. Under a symlinked cwd (macOS
// /tmp → /private/tmp, or any `ln -s`), filepath.Abs and the repo top then
// disagree on a prefix and the inside-repo guard misfires. runMigrate must
// EvalSymlinks both sides.
func TestRunMigrate_SymlinkedCwd(t *testing.T) {
	srcRoot, dstRoot := migrateRepos(t)
	parent := filepath.Dir(srcRoot)
	link := filepath.Join(parent, "link-parent")
	if err := os.Symlink(parent, link); err != nil {
		t.Skipf("symlink unavailable: %v", err)
	}
	// Simulate the shell: logical $PWD through the symlink, same physical dir.
	t.Setenv("PWD", filepath.Join(link, filepath.Base(srcRoot)))
	var errBuf strings.Builder
	if err := runMigrate(&migrateOpts{file: "data/project/p.md", destDir: "../dst", stderr: &errBuf}); err != nil {
		t.Fatalf("runMigrate under symlinked $PWD: %v\nstderr:\n%s", err, errBuf.String())
	}
	if _, err := os.Stat(filepath.Join(dstRoot, "data/project/p.md")); err != nil {
		t.Errorf("dest file not written: %v", err)
	}
}

// #179 close review I3: --dest-path must stay INSIDE the destination repo —
// a traversal value would write a stray file outside it before git add
// fails, breaking fail-closed. Plus the flag's happy path (previously only
// the "" default was covered).
func TestRunMigrate_DestPath(t *testing.T) {
	t.Run("relocates within the dest repo", func(t *testing.T) {
		srcRoot, dstRoot := migrateRepos(t)
		_ = srcRoot
		if err := runMigrate(&migrateOpts{file: "data/project/p.md", destDir: "../dst", destPath: "docs/moved/q.md", stderr: io.Discard}); err != nil {
			t.Fatalf("runMigrate --dest-path: %v", err)
		}
		if _, err := os.Stat(filepath.Join(dstRoot, "docs/moved/q.md")); err != nil {
			t.Errorf("dest-path file not written: %v", err)
		}
		if s := gitOut(t, dstRoot, "log", "-1", "--pretty=%s"); !strings.Contains(s, "migrate: receive docs/moved/q.md from src") {
			t.Errorf("dst commit subject = %q", s)
		}
	})
	t.Run("traversal refuses before any write", func(t *testing.T) {
		srcRoot, dstRoot := migrateRepos(t)
		msg, died := expectDie(t, func() {
			_ = runMigrate(&migrateOpts{file: "data/project/p.md", destDir: "../dst", destPath: "../evil.md", stderr: io.Discard})
		})
		if !died || !strings.Contains(msg, "escapes the destination repo") {
			t.Errorf("traversal refusal wrong (died=%v):\n%s", died, msg)
		}
		if _, err := os.Stat(filepath.Join(filepath.Dir(dstRoot), "evil.md")); !os.IsNotExist(err) {
			t.Errorf("stray file written outside the dest repo (stat err=%v)", err)
		}
		if _, err := os.Stat(filepath.Join(srcRoot, "data/project/p.md")); err != nil {
			t.Errorf("source must be untouched: %v", err)
		}
	})
}

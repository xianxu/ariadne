package gitx

import (
	"context"
	"fmt"
	"os"
	"strings"
	"testing"
)

// Measures the real ls-tree + cat-file subprocess boundary, independently of
// tracker YAML parsing and network fetch latency. Setup uses one shared blob to
// avoid measuring ten thousand fixture-writing subprocesses.
func BenchmarkSnapshotGitTenThousandFiles(b *testing.B) {
	b.Setenv("HOME", b.TempDir())
	b.Setenv("GIT_CONFIG_GLOBAL", os.DevNull)
	b.Setenv("GIT_CONFIG_NOSYSTEM", "1")
	root := b.TempDir()
	ctx := context.Background()
	run := func(input string, args ...string) string {
		out, diag, err := runGitInputContext(ctx, root, nil, strings.NewReader(input), args...)
		if err != nil {
			b.Fatalf("%v: %v %s", args, err, diag)
		}
		return strings.TrimSpace(string(out))
	}
	run("", "init", "-q")
	payload := strings.Repeat("x", 512)
	blob := run(payload, "hash-object", "-w", "--stdin")
	var entries strings.Builder
	for i := 1; i <= 10000; i++ {
		fmt.Fprintf(&entries, "100644 blob %s\t%06d-card.md\n", blob, i)
	}
	tree := run(entries.String(), "mktree")
	tf, err := NewTrunkFileContext(ctx, root, "unused", "issue-tracker")
	if err != nil {
		b.Fatal(err)
	}
	view := tf.ViewOf(tree)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		files, err := view.Files("")
		if err != nil || len(files) != 10000 || string(files[9999].Content) != payload {
			b.Fatalf("snapshot cardinality=%d error=%v", len(files), err)
		}
	}
	b.ReportMetric(2, "git-processes/op")
}

// The cutover guard's per-generation cost (#252): proving the marked root is
// in a tracker history of 10,000 card writes. One git process per new tracker
// generation; repeated reads of one generation are cached by the repository.
func BenchmarkHasRootTenThousandGenerations(b *testing.B) {
	b.Setenv("HOME", b.TempDir())
	b.Setenv("GIT_CONFIG_GLOBAL", os.DevNull)
	b.Setenv("GIT_CONFIG_NOSYSTEM", "1")
	root := b.TempDir()
	ctx := context.Background()
	run := func(input string, args ...string) string {
		out, diag, err := runGitInputContext(ctx, root, nil, strings.NewReader(input), args...)
		if err != nil {
			b.Fatalf("%v: %v %s", args, err, diag)
		}
		return strings.TrimSpace(string(out))
	}
	run("", "init", "-q")
	var stream strings.Builder
	for i := 1; i <= 10000; i++ {
		fmt.Fprintf(&stream, "commit refs/heads/issue-tracker\ncommitter t <t@t> %d +0000\ndata 1\nx\n", 1700000000+i)
		body := fmt.Sprintf("card %d\n", i)
		fmt.Fprintf(&stream, "M 100644 inline workshop/issue-cards/%06d-c.md\ndata %d\n%s\n", i%500, len(body), body)
	}
	run(stream.String(), "fast-import", "--quiet")
	first := run("", "rev-list", "--max-parents=0", "refs/heads/issue-tracker")
	tip := run("", "rev-parse", "refs/heads/issue-tracker")
	tf, err := NewTrunkFileContext(ctx, root, "unused", "issue-tracker")
	if err != nil {
		b.Fatal(err)
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if ok, err := tf.HasRoot(first, tip); err != nil || !ok {
			b.Fatalf("root not found: %v", err)
		}
	}
	b.ReportMetric(1, "git-processes/op")
}

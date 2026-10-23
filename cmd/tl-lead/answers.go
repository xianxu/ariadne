package main

import (
	"errors"
	"fmt"
	"io/fs"
	"path"
	"path/filepath"
	"strings"
	"time"

	"github.com/xianxu/ariadne/pkg/sidedoc"
)

// answerTarget checks that an answer link's target is a dated side doc of the
// session file (#455, #562): `<stem>.side/YYYY-MM/YYYY-MM-DD-NN-slug.md`
// (legacy dated flat paths also accepted), relative to the
// file, already written, with `type: side-doc` and a `parent:` that names the
// file back. Only such a doc shows in parley's attention finder under the
// session file. read reads a path relative to the file's directory.
func answerTarget(file, target string, read func(rel string) ([]byte, error)) error {
	side := path.Base(sidedoc.SideDir(filepath.Base(file)))
	want := fmt.Sprintf("%s/YYYY-MM/YYYY-MM-DD-NN-<slug>.md", side)
	clean := path.Clean(target)
	dir, name := path.Split(clean)
	validDir := dir == side+"/"
	if sidedoc.AnswerName(name) {
		// AnswerName owns the filename grammar; the month layout also
		// requires a real date and exactly its matching month directory.
		_, dateErr := time.Parse("2006-01-02", name[:10])
		validDir = validDir || (dateErr == nil && dir == side+"/"+name[:7]+"/")
	}
	if clean != target || !validDir || !sidedoc.AnswerName(name) {
		return fmt.Errorf("refused: TARGET %q is not an answer doc of this session file; write it as %s (tl-lead --help: Answer docs)", target, want)
	}
	b, err := read(clean)
	if errors.Is(err, fs.ErrNotExist) {
		return fmt.Errorf("refused: %s does not exist; write the answer doc first, then link it", clean)
	}
	if err != nil {
		return err
	}
	fields, ok := sidedoc.Fields(string(b))
	if !ok || fields["type"] != "side-doc" {
		return fmt.Errorf("refused: %s has no `type: side-doc` frontmatter; give it the side-doc frontmatter (tl-lead --help: Answer docs)", clean)
	}
	parent := strings.Trim(fields["parent"], `"`)
	if parent == "" || path.Join(dir, parent) != filepath.Base(file) {
		wantParent := sidedoc.RelPath(path.Dir(clean), filepath.Base(file))
		return fmt.Errorf("refused: %s's parent: is %q; it must name the session file, %s", clean, parent, wantParent)
	}
	return nil
}

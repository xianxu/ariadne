package issue

import (
	"fmt"
	"slices"
	"strings"

	"github.com/xianxu/ariadne/cmd/sdlc/internal/repoenum"
	"github.com/xianxu/ariadne/pkg/vocab"
	"go.yaml.in/yaml/v3"
)

// The single validators for the classification fields (#348): the card parser,
// the writers' check and the list filter all use these, so no site restates
// severity's values or area's syntax.

// EnumValid reports whether v is a value of the model's shared enum name.
func EnumValid(name, v string) bool { return slices.Contains(vocab.Issue().Enums[name], v) }

// RepoEnumValid reports whether v has the syntax of the model's repo enum name
// (membership in a repository's declaration is checked separately).
func RepoEnumValid(name, v string) bool {
	spec, ok := vocab.Issue().RepoEnums.Fields[name]
	return ok && repoenum.ValidValue(spec, v)
}

// CheckEnum refuses a value outside the model's shared enum name; "" passes.
func CheckEnum(name, v string) error {
	if v == "" || EnumValid(name, v) {
		return nil
	}
	return fmt.Errorf("%s %q is not one of %s", name, v, strings.Join(vocab.Issue().Enums[name], ", "))
}

// CheckSeverity refuses a severity outside the shared enum; "" passes.
func CheckSeverity(v string) error { return CheckEnum("severity", v) }

// CheckAreaSyntax refuses an area that is not a /-path of the model's segments.
func CheckAreaSyntax(v string) error {
	if RepoEnumValid("area", v) {
		return nil
	}
	return fmt.Errorf("area %q is not a /-path of lowercase segments (at most %d)", v, vocab.Issue().RepoEnums.Fields["area"].MaxDepth)
}

// FlowList renders values as a one-line YAML flow list whose items are all
// strings: a value whose plain form YAML would read as another type (2026,
// true, null, 1e5) is double-quoted, so what a writer emits is what the card
// shape check and every reader decode.
func FlowList(values []string) string {
	items := make([]string, len(values))
	for i, v := range values {
		items[i] = v
		var n yaml.Node
		if err := yaml.Unmarshal([]byte(v), &n); err != nil || len(n.Content) != 1 || n.Content[0].Tag != "!!str" || n.Content[0].Value != v {
			items[i] = fmt.Sprintf("%q", v)
		}
	}
	return "[" + strings.Join(items, ", ") + "]"
}

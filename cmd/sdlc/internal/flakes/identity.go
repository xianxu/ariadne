package flakes

import (
	"fmt"
	"strings"
	"unicode"
	"unicode/utf8"
)

// IdentityKind identifies the runner that owns a test identity.
type IdentityKind string

const (
	KindGo     IdentityKind = "go"
	KindScript IdentityKind = "script"
	KindLua    IdentityKind = "lua"
)

// Identity is a validated test identity or execution unit. Its components are
// immutable; constructors preserve opaque case labels exactly as supplied.
type Identity struct {
	kind      IdentityKind
	goName    Name
	path      string
	caseLabel string
}

func (id Identity) Kind() IdentityKind { return id.kind }
func (id Identity) Path() string       { return id.path }
func (id Identity) Case() string       { return id.caseLabel }

// Unit names the whole execution that can produce the identity's result.
func (id Identity) Unit() string {
	if id.kind == KindGo {
		return id.goName.Top()
	}
	if id.kind == "" {
		return ""
	}
	return string(id.kind) + ":" + id.path
}

func (id Identity) String() string {
	if id.kind == KindGo {
		return id.goName.String()
	}
	if id.caseLabel != "" {
		return id.Unit() + "#" + id.caseLabel
	}
	return id.Unit()
}

// Display is the label a runner prints when reporting the identity.
func (id Identity) Display() string {
	if id.kind == KindGo {
		return id.goName.Run()
	}
	return id.String()
}

func (id Identity) GoName() (Name, bool) { return id.goName, id.kind == KindGo }

// ParseIdentity accepts registrable Go tests, whole scripts and exact Lua cases.
// A bare Lua execution unit is deliberately not a registrable wildcard.
func ParseIdentity(s string) (Identity, error) {
	if strings.HasPrefix(s, "script:") {
		return ParseUnit(s)
	}
	if strings.HasPrefix(s, "lua:") {
		unit, label, found := strings.Cut(s, "#")
		if !found || label == "" || len(label) > 2048 || !validText(label) {
			return Identity{}, fmt.Errorf("%q: Lua identity requires a nonempty UTF-8 case label of at most 2048 bytes without controls", s)
		}
		id, err := ParseUnit(unit)
		if err != nil {
			return Identity{}, err
		}
		id.caseLabel = label
		return id, nil
	}
	if n, ok := ParseName(s); ok {
		return Identity{kind: KindGo, goName: n}, nil
	}
	return Identity{}, fmt.Errorf("%q is not a test identity: want <pkg>.TestName[/subtest], script:<path.sh>, or lua:<path.lua>#<case>", s)
}

// ParseUnit accepts a whole script or Lua file for execution, without a case.
func ParseUnit(s string) (Identity, error) {
	kind, path, found := strings.Cut(s, ":")
	var suffix string
	switch IdentityKind(kind) {
	case KindScript:
		suffix = ".sh"
	case KindLua:
		suffix = ".lua"
	default:
		return Identity{}, fmt.Errorf("%q: execution unit must start with script: or lua:", s)
	}
	if !found || len(path) == 0 || len(path) > 512 || !validText(path) || strings.ContainsAny(path, "\\#") || strings.HasPrefix(path, "-") || !strings.HasSuffix(path, suffix) {
		return Identity{}, fmt.Errorf("%q: want a clean relative UTF-8 %s path of at most 512 bytes", s, suffix)
	}
	for _, segment := range strings.Split(path, "/") {
		if segment == "" || segment == "." || segment == ".." {
			return Identity{}, fmt.Errorf("%q: path must be relative with no empty or dot segments", s)
		}
	}
	return Identity{kind: IdentityKind(kind), path: path}, nil
}

func validText(s string) bool {
	if !utf8.ValidString(s) {
		return false
	}
	for _, r := range s {
		if unicode.IsControl(r) {
			return false
		}
	}
	return true
}

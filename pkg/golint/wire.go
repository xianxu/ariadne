package golint

import (
	"go/ast"
	"go/token"
	"path"
	"strconv"
	"strings"
)

// CheckContext is the source location and one shared annotation index for a
// matcher. Package is the source directory relative to the module root.
type CheckContext struct {
	Path, Package, PackageName, Module string
	Index                              *WireTypeIndex
}

type wirePackage struct{ dir, name string }

// WireTypeIndex maps package directories to annotated named types.
type WireTypeIndex struct {
	marked                 map[wirePackage]map[string]bool
	declarations           map[*ast.TypeSpec]bool
	packageByDir           map[string]string
	fileVisits, typeVisits int // counted work, defended at N and 2N
}

func newWireTypeIndex(files map[string]*ast.File) *WireTypeIndex {
	index := &WireTypeIndex{marked: make(map[wirePackage]map[string]bool), declarations: map[*ast.TypeSpec]bool{}, packageByDir: map[string]string{}}
	ambiguous := map[string]bool{}
	for filename, f := range files {
		index.fileVisits++
		if ast.IsGenerated(f) {
			continue
		}
		pkg := path.Dir(filename)
		key := wirePackage{pkg, f.Name.Name}
		if index.marked[key] == nil {
			index.marked[key] = map[string]bool{}
		}
		if !strings.HasSuffix(f.Name.Name, "_test") {
			if existing, ok := index.packageByDir[pkg]; ok && existing != f.Name.Name {
				ambiguous[pkg] = true
			} else if !ambiguous[pkg] {
				index.packageByDir[pkg] = f.Name.Name
			}
		}
		marked, visited := markedWireTypes(f)
		index.typeVisits += visited
		for name, declaration := range marked {
			index.marked[key][name] = true
			index.declarations[declaration] = true
		}
	}
	for pkg := range ambiguous {
		delete(index.packageByDir, pkg)
	}
	return index
}

// markedWireTypes finds declarations explicitly documented as independently
// evolving JSON input. A declaration group documents its type only when it
// contains one type, so a comment cannot accidentally mark its siblings.
func markedWireTypes(f *ast.File) (map[string]*ast.TypeSpec, int) {
	marked := map[string]*ast.TypeSpec{}
	visited := 0
	for _, d := range f.Decls {
		group, ok := d.(*ast.GenDecl)
		if !ok || group.Tok != token.TYPE {
			continue
		}
		for _, spec := range group.Specs {
			t, ok := spec.(*ast.TypeSpec)
			if !ok {
				continue
			}
			visited++
			comments := t.Doc
			if comments == nil && len(group.Specs) == 1 {
				comments = group.Doc
			}
			if comments == nil {
				continue
			}
			for _, c := range comments.List {
				line := strings.TrimSpace(strings.TrimPrefix(c.Text, "//"))
				if reason, ok := strings.CutPrefix(line, "wire:"); ok && strings.TrimSpace(reason) != "" {
					marked[t.Name.Name] = t
				}
			}
		}
	}
	return marked, visited
}

func wireStrictDecode(f *ast.File, context *CheckContext) []ast.Node {
	jsonPkg := localName(f, "encoding/json")
	if jsonPkg == "" {
		return nil
	}
	if context == nil || context.Index == nil || len(context.Index.marked) == 0 {
		return nil
	}
	var out []ast.Node
	for _, d := range f.Decls {
		fn, ok := d.(*ast.FuncDecl)
		if !ok || fn.Body == nil {
			continue
		}
		// Only the function's direct statement list is an unconditional flow.
		// A flag set inside a branch cannot prove a later decode is strict.
		decoders := map[string]bool{} // value says DisallowUnknownFields ran
		variables := map[string]bool{}
		checkCall := func(e ast.Expr) {
			call, ok := e.(*ast.CallExpr)
			if !ok {
				return
			}
			sel, ok := call.Fun.(*ast.SelectorExpr)
			if !ok {
				return
			}
			id, ok := sel.X.(*ast.Ident)
			if !ok {
				return
			}
			strict, known := decoders[id.Name]
			if !known {
				return
			}
			switch sel.Sel.Name {
			case "DisallowUnknownFields":
				decoders[id.Name] = true
			case "Decode":
				if strict && len(call.Args) == 1 && wireTarget(call.Args[0], f, context, variables) {
					out = append(out, call)
				}
			}
		}
		for _, stmt := range fn.Body.List {
			// A decoder copied, passed, captured, or otherwise used beyond
			// these direct calls may be overwritten through an alias. Syntax
			// alone cannot prove it still has the strict setting afterward.
			discardEscapedDecoders(stmt, decoders)
			switch v := stmt.(type) {
			case *ast.DeclStmt:
				group, ok := v.Decl.(*ast.GenDecl)
				if !ok {
					continue
				}
				for _, spec := range group.Specs {
					value, ok := spec.(*ast.ValueSpec)
					if ok && len(value.Names) == 1 && len(value.Values) == 1 && isPackageCall(value.Values[0], jsonPkg, "NewDecoder") {
						decoders[value.Names[0].Name] = false
					}
					// An explicit interface type erases an initializer's wire
					// identity. Infer from a constructor only when no type was
					// declared and that initializer belongs to one name.
					if ok && (context.markedType(f, value.Type) || value.Type == nil && len(value.Names) == 1 && len(value.Values) == 1 && isWireConstructor(value.Values[0], f, context)) {
						for _, name := range value.Names {
							variables[name.Name] = true
						}
					}
				}
			case *ast.AssignStmt:
				if len(v.Rhs) == 1 {
					checkCall(v.Rhs[0])
				}
				for _, lhs := range v.Lhs {
					id := assignedRoot(lhs)
					if id == nil || id.Name == "_" {
						continue
					}
					// A write can replace either the target value or the
					// decoder's strict configuration. Establish a new binding
					// only from a simple, direct constructor assignment.
					delete(variables, id.Name)
					delete(decoders, id.Name)
					if _, ok := lhs.(*ast.Ident); !ok {
						continue
					}
					if len(v.Lhs) == 1 && len(v.Rhs) == 1 && v.Tok == token.DEFINE && isWireConstructor(v.Rhs[0], f, context) {
						variables[id.Name] = true
					}
					if len(v.Lhs) == 1 && len(v.Rhs) == 1 && isPackageCall(v.Rhs[0], jsonPkg, "NewDecoder") {
						decoders[id.Name] = false
					}
				}
			case *ast.ExprStmt:
				checkCall(v.X)
			case *ast.ReturnStmt:
				for _, result := range v.Results {
					checkCall(result)
				}
			case *ast.IfStmt:
				// An if-init executes before the condition; its branches do not
				// necessarily run. The common `if err := d.Decode(...)` fits.
				if init, ok := v.Init.(*ast.AssignStmt); ok && len(init.Rhs) == 1 {
					checkCall(init.Rhs[0])
				}
				invalidateAssigned(v, decoders, variables)
			default:
				invalidateAssigned(stmt, decoders, variables)
			}
		}
	}
	return out
}

// A branch may replace a known decoder or target on only one path. Drop both
// identities rather than guessing which value a later Decode uses.
func invalidateAssigned(node ast.Node, decoders, variables map[string]bool) {
	ast.Inspect(node, func(n ast.Node) bool {
		var left []ast.Expr
		switch v := n.(type) {
		case *ast.AssignStmt:
			left = v.Lhs
		case *ast.RangeStmt:
			left = []ast.Expr{v.Key, v.Value}
		}
		for _, lhs := range left {
			if id := assignedRoot(lhs); id != nil {
				delete(decoders, id.Name)
				delete(variables, id.Name)
			}
		}
		return true
	})
}

// Keep a decoder proof only while its local name is used as the receiver of
// the two modeled calls. Any other use can expose an alias; proving the alias
// graph would require type/flow analysis outside this syntax-only rule.
func discardEscapedDecoders(node ast.Node, decoders map[string]bool) {
	allowed := map[*ast.Ident]bool{}
	ast.Inspect(node, func(n ast.Node) bool {
		call, ok := n.(*ast.CallExpr)
		if !ok {
			return true
		}
		sel, ok := call.Fun.(*ast.SelectorExpr)
		if !ok || sel.Sel.Name != "Decode" && sel.Sel.Name != "DisallowUnknownFields" {
			return true
		}
		if id, ok := sel.X.(*ast.Ident); ok {
			allowed[id] = true
		}
		return true
	})
	escaped := map[string]bool{}
	ast.Inspect(node, func(n ast.Node) bool {
		if closure, ok := n.(*ast.FuncLit); ok {
			ast.Inspect(closure.Body, func(inner ast.Node) bool {
				if id, ok := inner.(*ast.Ident); ok {
					escaped[id.Name] = true
				}
				return true
			})
			return false
		}
		if id, ok := n.(*ast.Ident); ok && !allowed[id] {
			escaped[id.Name] = true
		}
		return true
	})
	for name := range escaped {
		delete(decoders, name)
	}
}

// A direct assignment or a write through a tracked pointer can replace the
// value whose provenance justified a finding. Other expressions have no
// single known local root and are outside this bounded matcher.
func assignedRoot(e ast.Expr) *ast.Ident {
	switch v := e.(type) {
	case *ast.Ident:
		return v
	case *ast.ParenExpr:
		return assignedRoot(v.X)
	case *ast.StarExpr:
		return assignedRoot(v.X)
	}
	return nil
}

func isWireConstructor(e ast.Expr, f *ast.File, context *CheckContext) bool {
	switch v := e.(type) {
	case *ast.CompositeLit:
		return context.markedType(f, v.Type)
	case *ast.UnaryExpr:
		return v.Op == token.AND && isWireConstructor(v.X, f, context)
	case *ast.CallExpr:
		id, ok := v.Fun.(*ast.Ident)
		return ok && id.Name == "new" && len(v.Args) == 1 && context.markedType(f, v.Args[0])
	}
	return false
}

func isPackageCall(e ast.Expr, pkg, name string) bool {
	call, ok := e.(*ast.CallExpr)
	if !ok {
		return false
	}
	sel, ok := call.Fun.(*ast.SelectorExpr)
	if !ok || sel.Sel.Name != name {
		return false
	}
	id, ok := sel.X.(*ast.Ident)
	// Imported package identifiers are unresolved by the parser. A local
	// variable shadowing that name has an Obj and is not encoding/json.
	return ok && id.Name == pkg && id.Obj == nil
}

func wireTarget(e ast.Expr, f *ast.File, context *CheckContext, variables map[string]bool) bool {
	if p, ok := e.(*ast.UnaryExpr); ok && p.Op == token.AND {
		e = p.X
	}
	switch v := e.(type) {
	case *ast.Ident:
		return variables[v.Name]
	case *ast.CompositeLit:
		return context.markedType(f, v.Type)
	}
	return false
}

func (context *CheckContext) markedType(f *ast.File, e ast.Expr) bool {
	if p, ok := e.(*ast.StarExpr); ok {
		e = p.X
	}
	switch t := e.(type) {
	case *ast.Ident:
		if t.Obj != nil && t.Obj.Kind == ast.Typ {
			if declaration, ok := t.Obj.Decl.(*ast.TypeSpec); ok {
				return context.Index.declarations[declaration]
			}
		}
		return context.Index.marked[wirePackage{context.Package, context.PackageName}][t.Name]
	case *ast.SelectorExpr:
		alias, ok := t.X.(*ast.Ident)
		if !ok || context.Module == "" {
			return false
		}
		for _, imp := range f.Imports {
			importPath, err := strconv.Unquote(imp.Path.Value)
			if err != nil || !strings.HasPrefix(importPath, context.Module+"/") {
				continue
			}
			dir := strings.TrimPrefix(importPath, context.Module+"/")
			name := context.Index.packageByDir[dir]
			if imp.Name != nil {
				name = imp.Name.Name
			}
			if name == alias.Name {
				return context.Index.marked[wirePackage{dir, context.Index.packageByDir[dir]}][t.Sel.Name]
			}
		}
	}
	return false
}

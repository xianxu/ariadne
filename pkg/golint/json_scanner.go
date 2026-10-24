package golint

import (
	"go/ast"
	"go/token"
	"strings"
)

type jsonScannerFunction struct {
	summary                *jsonScannerSummary
	parent                 *jsonScannerFunction
	token, more, recursive bool
	uses                   []jsonScannerUse
	node                   ast.Node
	name                   *ast.Ident
	receiver               *ast.Object
	calls                  []*ast.CallExpr
	assignments            []*ast.AssignStmt
	values                 []*ast.ValueSpec
	returns                []*ast.ReturnStmt
	maps                   bool
	stableSelf             bool
}
type jsonScannerUse struct {
	id     *ast.Ident
	parent ast.Node
}
type jsonScannerReference struct {
	owner *jsonScannerFunction
	call  *ast.CallExpr
}
type jsonScannerSummary struct {
	useOwners      map[*ast.Object]*jsonScannerFunction
	captured       map[*ast.Object]bool
	indexedWrites  map[*ast.Object]bool
	aliased        map[*ast.Object]bool
	bindingWrites  map[*ast.Object]int
	addressed      map[*ast.Object]bool
	roleNodeVisits int
	roleUseVisits  int
	functions      []*jsonScannerFunction
	declarations   map[*ast.Object]*jsonScannerFunction
	references     map[*ast.Object][]jsonScannerReference
	closures       map[*ast.FuncLit]*ast.Ident
	visits         int
	edgeVisits     int
}

type jsonScannerVisitor struct {
	summary *jsonScannerSummary
	owner   *jsonScannerFunction
	parent  ast.Node
}

// Each node belongs to one lexical function, including nested closures. A body
// is never re-walked for each scanner candidate or each potential role.
func (v jsonScannerVisitor) Visit(n ast.Node) ast.Visitor {
	if n == nil {
		return nil
	}
	s := v.summary
	s.visits++
	switch x := n.(type) {
	case *ast.AssignStmt:
		for _, lhs := range x.Lhs {
			if id := assignedRoot(lhs); id != nil && id.Obj != nil {
				s.bindingWrites[id.Obj]++
			} else if id := jsonBindingRoot(lhs); id != nil && id.Obj != nil {
				s.indexedWrites[id.Obj] = true
			}
		}
		for i, rhs := range x.Rhs {
			if id, ok := ast.Unparen(rhs).(*ast.Ident); ok && id.Obj != nil && i < len(x.Lhs) {
				if _, indexStore := ast.Unparen(x.Lhs[i]).(*ast.IndexExpr); !indexStore {
					s.aliased[id.Obj] = true
				}
			}
		}
		for i, rhs := range x.Rhs {
			if lit, ok := rhs.(*ast.FuncLit); ok && i < len(x.Lhs) {
				if id, ok := x.Lhs[i].(*ast.Ident); ok {
					s.closures[lit] = id
				}
			}
		}
	case *ast.FuncDecl:
		fn := &jsonScannerFunction{summary: s, node: x, name: x.Name, parent: v.owner}
		if x.Recv != nil && len(x.Recv.List) > 0 && len(x.Recv.List[0].Names) > 0 {
			fn.receiver = x.Recv.List[0].Names[0].Obj
		}
		s.functions = append(s.functions, fn)
		v.owner = fn
		if x.Recv == nil && x.Name.Obj != nil {
			s.declarations[x.Name.Obj] = fn
		}
	case *ast.FuncLit:
		fn := &jsonScannerFunction{summary: s, node: x, name: s.closures[x], parent: v.owner}
		s.functions = append(s.functions, fn)
		v.owner = fn
	case *ast.ValueSpec:
		for i, id := range x.Names {
			if i < len(x.Values) && id.Obj != nil {
				s.bindingWrites[id.Obj]++
				if source, ok := ast.Unparen(x.Values[i]).(*ast.Ident); ok && source.Obj != nil {
					s.aliased[source.Obj] = true
				}
			}
		}
	case *ast.IncDecStmt:
		if id := jsonBindingRoot(x.X); id != nil && id.Obj != nil {
			s.indexedWrites[id.Obj] = true
		}
	case *ast.UnaryExpr:
		if x.Op == token.AND {
			if id := jsonBindingRoot(x.X); id != nil && id.Obj != nil {
				s.addressed[id.Obj] = true
			}
		}
	}
	if id, ok := n.(*ast.Ident); ok && id.Obj != nil {
		if owner, seen := s.useOwners[id.Obj]; seen && owner != v.owner {
			s.captured[id.Obj] = true
		} else if !seen {
			s.useOwners[id.Obj] = v.owner
		}
	}
	if fn := v.owner; fn != nil {
		if id, ok := n.(*ast.Ident); ok && id.Obj != nil {
			fn.uses = append(fn.uses, jsonScannerUse{id: id, parent: v.parent})
		}
		switch x := n.(type) {
		case *ast.CallExpr:
			fn.calls = append(fn.calls, x)
		case *ast.AssignStmt:
			fn.assignments = append(fn.assignments, x)
		case *ast.ValueSpec:
			fn.values = append(fn.values, x)
		case *ast.ReturnStmt:
			fn.returns = append(fn.returns, x)
		case *ast.MapType:
			if key, ok := x.Key.(*ast.Ident); ok && key.Name == "string" {
				fn.maps = true
			}
		}
	}
	if id, ok := n.(*ast.Ident); ok && id.Obj != nil && id.Obj.Kind == ast.Fun {
		if decl, ok := id.Obj.Decl.(*ast.FuncDecl); !ok || decl.Name != id {
			var call *ast.CallExpr
			if p, ok := v.parent.(*ast.CallExpr); ok && p.Fun == id {
				call = p
			}
			s.references[id.Obj] = append(s.references[id.Obj], jsonScannerReference{owner: v.owner, call: call})
		}
	}
	v.parent = n
	return v
}

func summarizeJSONScanners(f *ast.File) *jsonScannerSummary {
	s := &jsonScannerSummary{useOwners: map[*ast.Object]*jsonScannerFunction{}, captured: map[*ast.Object]bool{}, indexedWrites: map[*ast.Object]bool{}, aliased: map[*ast.Object]bool{}, bindingWrites: map[*ast.Object]int{}, addressed: map[*ast.Object]bool{}, declarations: map[*ast.Object]*jsonScannerFunction{}, references: map[*ast.Object][]jsonScannerReference{}, closures: map[*ast.FuncLit]*ast.Ident{}}
	ast.Walk(jsonScannerVisitor{summary: s}, f)
	objects, receivers := map[*ast.Object]*jsonScannerFunction{}, map[*ast.Object]*jsonScannerFunction{}
	for _, fn := range s.functions {
		if fn.name != nil && fn.name.Obj != nil {
			fn.stableSelf = fn.name.Obj.Kind == ast.Fun || (s.bindingWrites[fn.name.Obj] == 1 && !s.addressed[fn.name.Obj])
			objects[fn.name.Obj] = fn
		}
		if fn.receiver != nil {
			receivers[fn.receiver] = fn
		}
	}
	for _, fn := range s.functions {
		for _, call := range fn.calls {
			var ancestor *jsonScannerFunction
			switch target := call.Fun.(type) {
			case *ast.SelectorExpr:
				fn.token = fn.token || target.Sel.Name == "Token" && len(call.Args) == 0
				fn.more = fn.more || target.Sel.Name == "More" && len(call.Args) == 0
				fn.recursive = fn.recursive || fn.name != nil && target.Sel.Name == fn.name.Name
				if recv, ok := target.X.(*ast.Ident); ok {
					if outer := receivers[recv.Obj]; outer != nil && target.Sel.Name == outer.name.Name {
						ancestor = outer
					}
				}
			case *ast.Ident:
				fn.recursive = fn.recursive || fn.name != nil && target.Name == fn.name.Name
				ancestor = objects[target.Obj]
			}
			if ancestor != nil && call.Pos() >= ancestor.node.Pos() && call.End() <= ancestor.node.End() {
				ancestor.recursive = true
			}
		}
	}
	// Propagate bounded feature bits, never copied call sets, through lexical
	// nesting. A nested closure cannot hide an actual call to its outer scanner.
	for i := len(s.functions) - 1; i >= 0; i-- {
		fn := s.functions[i]
		if parent := fn.parent; parent != nil {
			parent.token = parent.token || fn.token
			parent.more = parent.more || fn.more
			parent.maps = parent.maps || fn.maps
		}
	}
	return s
}

func recursiveJSONScanners(f *ast.File, context *CheckContext) []ast.Node {
	if f.Name.Name == "jsonstrict" {
		return nil
	} // existing owner boundary (#489)
	s := summarizeJSONScanners(f)
	oracles := jsonDifferentialOracleRoles(f, context, s)
	var out []ast.Node
	for _, fn := range s.functions {
		if fn.name == nil || !fn.maps {
			continue
		}

		if fn.token && fn.more && fn.recursive && !oracles[fn] && !jsonMaterializerRole(fn) {
			out = append(out, fn.node)
		}
	}
	return out
}

func jsonSelfCall(fn *jsonScannerFunction, e ast.Expr) bool {
	call, ok := e.(*ast.CallExpr)
	if !ok || fn.name == nil {
		return false
	}
	switch target := call.Fun.(type) {
	case *ast.Ident:
		return fn.stableSelf && target.Obj != nil && target.Obj == fn.name.Obj
	case *ast.SelectorExpr:
		receiver, ok := target.X.(*ast.Ident)
		return ok && fn.receiver != nil && receiver.Obj == fn.receiver && target.Sel.Name == fn.name.Name
	}
	return false
}

// Project mutation/address roots once. Writes to an element do not rebind the
// container, but do invalidate a scanner input's or child value's provenance.
func jsonBindingRoot(e ast.Expr) *ast.Ident {
	for {
		switch x := e.(type) {
		case *ast.Ident:
			return x
		case *ast.ParenExpr:
			e = x.X
		case *ast.StarExpr:
			e = x.X
		case *ast.IndexExpr:
			e = x.X
		case *ast.SelectorExpr:
			e = x.X
		default:
			return nil
		}
	}
}

func jsonStableBinding(fn *jsonScannerFunction, obj *ast.Object, writes int) bool {
	if obj == nil {
		return false
	}
	s := fn.summary
	return s.bindingWrites[obj] == writes && !s.addressed[obj] && !s.captured[obj]
}

func jsonStableInput(fn *jsonScannerFunction, obj *ast.Object) bool {
	if obj == nil {
		return false
	}
	writes := 0
	switch decl := obj.Decl.(type) {
	case *ast.AssignStmt:
		writes = 1
	case *ast.ValueSpec:
		if len(decl.Values) > 0 {
			writes = 1
		}
	}
	s := fn.summary
	return jsonStableBinding(fn, obj, writes) && !s.indexedWrites[obj] && !s.aliased[obj]
}

func jsonContainerType(e ast.Expr) bool {
	switch x := e.(type) {
	case *ast.MapType:
		key, ok := x.Key.(*ast.Ident)
		return ok && key.Name == "string"
	case *ast.ArrayType:
		return x.Len == nil
	}
	return false
}
func jsonContainerInit(e ast.Expr) bool {
	if lit, ok := e.(*ast.CompositeLit); ok {
		return jsonContainerType(lit.Type)
	}
	if call, ok := e.(*ast.CallExpr); ok {
		id, ok := call.Fun.(*ast.Ident)
		return ok && id.Name == "make" && id.Obj == nil && len(call.Args) > 0 && jsonContainerType(call.Args[0])
	}
	return false
}

// A returned container must actually store a recursive child value. Merely
// declaring a map, returning (any,error), or remembering seen keys is not a role.
func jsonMaterializerRole(fn *jsonScannerFunction) bool {
	containers, children, writes := map[*ast.Object]bool{}, map[*ast.Object]bool{}, map[*ast.Object]int{}
	initializers, appendWrites := map[*ast.Object]int{}, map[*ast.Object]int{}
	for _, value := range fn.values {
		for i, id := range value.Names {
			if i < len(value.Values) {
				containers[id.Obj] = jsonContainerInit(value.Values[i])
				if containers[id.Obj] {
					initializers[id.Obj]++
				}
				children[id.Obj] = jsonSelfCall(fn, value.Values[i])
				writes[id.Obj]++
			}
		}
	}
	for _, assign := range fn.assignments {
		for i, lhs := range assign.Lhs {
			id, ok := lhs.(*ast.Ident)
			if !ok || id.Obj == nil {
				continue
			}
			writes[id.Obj]++
			if i < len(assign.Rhs) {
				if jsonContainerInit(assign.Rhs[i]) {
					containers[id.Obj] = true
					initializers[id.Obj]++
				}
				children[id.Obj] = children[id.Obj] || jsonSelfCall(fn, assign.Rhs[i])
			}
		}
	}
	// Calls other than the value-storing builtin append cannot prove that the
	// recursive child is still the decoded value at storage time.
	for _, use := range fn.uses {
		if !children[use.id.Obj] {
			continue
		}
		switch parent := use.parent.(type) {
		case *ast.AssignStmt, *ast.ValueSpec: // declaration or storage; aliases are checked separately
		case *ast.CallExpr:
			builtin, isIdent := parent.Fun.(*ast.Ident)
			if !isIdent || builtin.Name != "append" || builtin.Obj != nil || len(parent.Args) == 0 || parent.Args[0] == use.id {
				children[use.id.Obj] = false
			}
		default:
			children[use.id.Obj] = false
		}
	}
	child := func(e ast.Expr) bool {
		id, ok := e.(*ast.Ident)
		return (ok && id.Obj != nil && children[id.Obj] && writes[id.Obj] == 1 && jsonStableBinding(fn, id.Obj, 1) && !fn.summary.indexedWrites[id.Obj] && !fn.summary.aliased[id.Obj]) || jsonSelfCall(fn, e)
	}
	stored := map[*ast.Object]bool{}
	appendCalls := map[*ast.CallExpr]*ast.Object{}
	for _, assign := range fn.assignments {
		for i, lhs := range assign.Lhs {
			if i >= len(assign.Rhs) {
				continue
			}
			if index, ok := lhs.(*ast.IndexExpr); ok {
				if id, ok := index.X.(*ast.Ident); ok && containers[id.Obj] && writes[id.Obj] == 1 && child(assign.Rhs[i]) {
					stored[id.Obj] = true
				}
			}
			id, ok := lhs.(*ast.Ident)
			call, isCall := assign.Rhs[i].(*ast.CallExpr)
			if !ok || !isCall || !containers[id.Obj] || len(call.Args) < 2 {
				continue
			}
			appendID, ok := call.Fun.(*ast.Ident)
			base, baseOK := call.Args[0].(*ast.Ident)
			if ok && appendID.Name == "append" && appendID.Obj == nil && baseOK && base.Obj == id.Obj {
				appendCalls[call] = id.Obj
				appendWrites[id.Obj]++
				if child(call.Args[1]) {
					stored[id.Obj] = true
				}
			}
		}
	}
	escaped := map[*ast.Object]bool{}
	for _, use := range fn.uses {
		if !containers[use.id.Obj] {
			continue
		}
		safe := false
		switch parent := use.parent.(type) {
		case *ast.AssignStmt:
			for _, lhs := range parent.Lhs {
				safe = safe || lhs == use.id
			}
		case *ast.ValueSpec:
			for _, name := range parent.Names {
				safe = safe || name == use.id
			}
		case *ast.IndexExpr:
			safe = parent.X == use.id
		case *ast.ReturnStmt:
			safe = true
		case *ast.CallExpr:
			safe = appendCalls[parent] == use.id.Obj && len(parent.Args) > 0 && parent.Args[0] == use.id
		}
		if !safe {
			escaped[use.id.Obj] = true
		}
	}
	for _, ret := range fn.returns {
		for _, e := range ret.Results {
			if id, ok := e.(*ast.Ident); ok && id.Obj != nil && !escaped[id.Obj] && stored[id.Obj] && initializers[id.Obj] == 1 && writes[id.Obj] == initializers[id.Obj]+appendWrites[id.Obj] && jsonStableBinding(fn, id.Obj, writes[id.Obj]) {
				return true
			}
		}
	}
	return false
}

func jsonFuzzParameter(f *ast.File, fn *ast.FuncDecl) *ast.Object {
	testing := localName(f, "testing")
	if testing == "" || !strings.HasPrefix(fn.Name.Name, "Fuzz") || fn.Type.Params == nil {
		return nil
	}
	for _, field := range fn.Type.Params.List {
		ptr, ok := field.Type.(*ast.StarExpr)
		if !ok {
			continue
		}
		selector, ok := ptr.X.(*ast.SelectorExpr)
		if !ok {
			continue
		}
		pkg, ok := selector.X.(*ast.Ident)
		if ok && pkg.Name == testing && selector.Sel.Name == "F" && len(field.Names) == 1 {
			return field.Names[0].Obj
		}
	}
	return nil
}

func jsonNilComparison(e ast.Expr) (*ast.Object, token.Token) {
	for {
		if paren, ok := e.(*ast.ParenExpr); ok {
			e = paren.X
		} else {
			break
		}
	}
	b, ok := e.(*ast.BinaryExpr)
	if !ok || (b.Op != token.EQL && b.Op != token.NEQ) {
		return nil, token.ILLEGAL
	}
	lhs, lok := b.X.(*ast.Ident)
	rhs, rok := b.Y.(*ast.Ident)
	if !lok || !rok {
		return nil, token.ILLEGAL
	}
	if rhs.Name == "nil" && rhs.Obj == nil {
		return lhs.Obj, b.Op
	}
	if lhs.Name == "nil" && lhs.Obj == nil {
		return rhs.Obj, b.Op
	}
	return nil, token.ILLEGAL
}

// Only named, unshadowed scanner calls on the same bound input establish the
// differential idiom. Function variables and aliases do not prove independence.
func jsonComparedScanners(a, b *ast.CallExpr) bool {
	if a == nil || b == nil || len(a.Args) != 1 || len(b.Args) != 1 {
		return false
	}
	x, xok := a.Fun.(*ast.Ident)
	y, yok := b.Fun.(*ast.Ident)
	if !xok || !yok || x.Name == y.Name || (x.Obj != nil && x.Obj.Kind != ast.Fun) || (y.Obj != nil && y.Obj.Kind != ast.Fun) {
		return false
	}
	left, lok := a.Args[0].(*ast.Ident)
	right, rok := b.Args[0].(*ast.Ident)
	return lok && rok && left.Obj != nil && left.Obj == right.Obj
}

// An input must be used only by compared calls. Merely preserving
// its binding is insufficient for a mutable byte slice passed to another
// helper. A diagnostic on the callback's actual *testing.T is observational.
func jsonOracleInputCalls(f *ast.File, fn *jsonScannerFunction, lit *ast.FuncLit) map[*ast.Object]int {
	diagnostics := map[*ast.Object]bool{}
	testingName := localName(f, "testing")
	if lit.Type.Params != nil {
		for _, field := range lit.Type.Params.List {
			ptr, ok := field.Type.(*ast.StarExpr)
			if !ok {
				continue
			}
			sel, ok := ptr.X.(*ast.SelectorExpr)
			if !ok {
				continue
			}
			pkg, ok := sel.X.(*ast.Ident)
			if ok && pkg.Name == testingName && pkg.Obj == nil && sel.Sel.Name == "T" {
				for _, name := range field.Names {
					diagnostics[name.Obj] = true
				}
			}
		}
	}
	counts, unknown := map[*ast.Object]int{}, map[*ast.Object]bool{}
	for _, use := range fn.uses {
		obj := use.id.Obj
		fn.summary.roleUseVisits++
		switch parent := use.parent.(type) {
		case *ast.Field: // declaration
		case *ast.CallExpr:
			if sel, ok := parent.Fun.(*ast.SelectorExpr); ok {
				if recv, ok := sel.X.(*ast.Ident); ok && diagnostics[recv.Obj] {
					switch sel.Sel.Name {
					case "Fatal", "Fatalf", "Error", "Errorf", "Log", "Logf":
						continue
					}
				}
			}
			// An identifier immediately owned by a CallExpr is either its
			// function or a direct argument; no per-use argument-list scan.
			if parent.Fun != use.id {
				counts[obj]++
			} else {
				unknown[obj] = true
			}
		default:
			unknown[obj] = true
		}
	}
	for obj := range unknown {
		counts[obj] = -1
	}
	return counts
}

func jsonDifferentialOracleRoles(f *ast.File, context *CheckContext, s *jsonScannerSummary) map[*jsonScannerFunction]bool {
	eligible := map[*jsonScannerFunction]bool{}
	if context == nil || !strings.HasSuffix(context.Path, "_test.go") {
		return eligible
	}
	callbacks := map[*ast.FuncLit]bool{}
	for _, fn := range s.functions {
		decl, ok := fn.node.(*ast.FuncDecl)
		if !ok {
			continue
		}
		parameter := jsonFuzzParameter(f, decl)
		if parameter == nil {
			continue
		}
		for _, call := range fn.calls {
			s.edgeVisits++
			sel, ok := call.Fun.(*ast.SelectorExpr)
			if !ok || sel.Sel.Name != "Fuzz" || len(call.Args) != 1 {
				continue
			}
			recv, ok := sel.X.(*ast.Ident)
			if !ok || recv.Obj != parameter {
				continue
			}
			if lit, ok := call.Args[0].(*ast.FuncLit); ok {
				callbacks[lit] = true
			}
		}
	}
	seeds := map[*ast.CallExpr]bool{}
	for _, fn := range s.functions {
		lit, ok := fn.node.(*ast.FuncLit)
		if !ok || !callbacks[lit] {
			continue
		}
		inputCalls := jsonOracleInputCalls(f, fn, lit)
		potential := map[*ast.CallExpr]*ast.Object{}
		results := map[*ast.Object]*ast.CallExpr{}
		counts := map[*ast.Object]int{}
		for _, assign := range fn.assignments {
			for i, lhs := range assign.Lhs {
				if id, ok := lhs.(*ast.Ident); ok && id.Obj != nil {
					counts[id.Obj]++
					if len(assign.Lhs) == 1 && len(assign.Rhs) == 1 && i == 0 {
						if call, ok := assign.Rhs[0].(*ast.CallExpr); ok {
							results[id.Obj] = call
						}
					}
				}
			}
		}
		ast.Inspect(lit.Body, func(n ast.Node) bool {
			if n != nil {
				s.roleNodeVisits++
			}
			if _, nested := n.(*ast.FuncLit); nested {
				return false
			}
			bin, ok := n.(*ast.BinaryExpr)
			if !ok || (bin.Op != token.EQL && bin.Op != token.NEQ) {
				return true
			}
			x, opX := jsonNilComparison(bin.X)
			y, opY := jsonNilComparison(bin.Y)
			if x != nil && y != nil && counts[x] == 1 && counts[y] == 1 && jsonStableBinding(fn, x, 1) && jsonStableBinding(fn, y, 1) && opX == opY && jsonComparedScanners(results[x], results[y]) {
				input := results[x].Args[0].(*ast.Ident)
				if !jsonStableInput(fn, input.Obj) {
					return true
				}
				potential[results[x]] = input.Obj
				potential[results[y]] = input.Obj
			}
			return true
		})
		comparedCalls := map[*ast.Object]int{}
		for _, input := range potential {
			comparedCalls[input]++
		}
		for call, input := range potential {
			if inputCalls[input] == comparedCalls[input] {
				seeds[call] = true
			}
		}
	}
	var pending []*jsonScannerFunction
	for call := range seeds {
		s.edgeVisits++
		if id, ok := call.Fun.(*ast.Ident); ok {
			if fn := s.declarations[id.Obj]; fn != nil {
				pending = append(pending, fn)
			}
		}
	}
	for len(pending) > 0 {
		s.edgeVisits++
		fn := pending[len(pending)-1]
		pending = pending[:len(pending)-1]
		if eligible[fn] {
			continue
		}
		eligible[fn] = true
		for _, call := range fn.calls {
			s.edgeVisits++
			if id, ok := call.Fun.(*ast.Ident); ok {
				if next := s.declarations[id.Obj]; next != nil {
					pending = append(pending, next)
				}
			}
		}
	}
	// Any outside call or value escape invalidates that helper and its callees.
	var invalid []*jsonScannerFunction
	for fn := range eligible {
		for _, ref := range s.references[fn.name.Obj] {
			s.edgeVisits++
			if ref.call == nil || (!eligible[ref.owner] && !seeds[ref.call]) {
				invalid = append(invalid, fn)
				break
			}
		}
	}
	for len(invalid) > 0 {
		s.edgeVisits++
		fn := invalid[len(invalid)-1]
		invalid = invalid[:len(invalid)-1]
		if !eligible[fn] {
			continue
		}
		delete(eligible, fn)
		for _, call := range fn.calls {
			s.edgeVisits++
			if id, ok := call.Fun.(*ast.Ident); ok {
				if next := s.declarations[id.Obj]; next != nil && eligible[next] {
					invalid = append(invalid, next)
				}
			}
		}
	}
	return eligible
}

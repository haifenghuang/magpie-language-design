package eval

import (
	"fmt"
	"io"
	"magpie/ast"
	"unicode"
)

func NewScope(p *Scope, w io.Writer) *Scope {
	s := make(map[string]Object)
	ss := make(map[string]*ast.StructStatement)
	ret := &Scope{store: s, parentScope: p, structStore: ss}
	if p == nil {
		ret.Writer = w
	} else {
		ret.Writer = p.Writer
	}

	return ret
}

type Scope struct {
	store       map[string]Object
	parentScope *Scope
	Writer      io.Writer

	structStore map[string]*ast.StructStatement
}

//Get all exported to 'anotherScope'
func (s *Scope) GetAllExported(anotherScope *Scope) {
	for key, value := range s.store {
		if unicode.IsUpper(rune(key[0])) { //only upppercase functions/variables are exported
			anotherScope.Set(key, value)
		}
	}

	for key, value := range s.structStore {
		if unicode.IsUpper(rune(key[0])) { //only upppercase struct name are exported
			anotherScope.SetStruct(value)
		}
	}
}

func (s *Scope) Get(name string) (Object, bool) {
	obj, ok := s.store[name]
	if !ok && s.parentScope != nil {
		obj, ok = s.parentScope.Get(name)
	}
	return obj, ok
}

// Get all the keys of the scope.
func (s *Scope) GetKeys() []string {
	keys := make([]string, 0, len(s.store))
	for k := range s.store {
		keys = append(keys, k)
	}
	return keys
}

func (s *Scope) DebugPrint(indent string) {

	for k, v := range s.store {
		fmt.Fprintf(s.Writer, "%s<%s> = <%s>  value.Type: %T\n", indent, k, v.Inspect(), v)
	}

	if s.parentScope != nil {
		fmt.Fprintf(s.Writer, "\n%sParentScope:\n", indent)
		s.parentScope.DebugPrint(indent + "  ")
	}

}

func (s *Scope) Set(name string, val Object) Object {
	s.store[name] = val
	return val
}

func (s *Scope) Del(name string) {
	delete(s.store, name)
}

func (s *Scope) GetStruct(name string) (*ast.StructStatement, bool) {
	obj, ok := s.structStore[name]
	if !ok && s.parentScope != nil {
		obj, ok = s.parentScope.GetStruct(name)
	}
	return obj, ok
}

func (s *Scope) SetStruct(structStmt *ast.StructStatement) *ast.StructStatement {
	s.structStore[structStmt.Name] = structStmt
	return structStmt
}

var GlobalScopes map[string]Object = make(map[string]Object)

func GetGlobalObj(name string) (Object, bool) {
	obj, ok := GlobalScopes[name]
	return obj, ok
}

func SetGlobalObj(name string, Obj Object) {
	GlobalScopes[name] = Obj
}

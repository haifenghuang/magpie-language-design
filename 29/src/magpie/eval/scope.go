package eval

import (
	"fmt"
	"io"
	"unicode"
)

func NewScope(p *Scope, w io.Writer) *Scope {
	s := make(map[string]Object)
	ret := &Scope{store: s, parentScope: p}
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
}

//Get all exported to 'anotherScope'
func (s *Scope) GetAllExported(anotherScope *Scope) {
	for key, value := range s.store {
		if unicode.IsUpper(rune(key[0])) { //only upppercase functions/variables are exported
			anotherScope.Set(key, value)
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

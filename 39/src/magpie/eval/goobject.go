package eval

import (
	"errors"
	"fmt"
	"reflect"
	"strings"
)

var (
	ERR_HASDOT           = errors.New("symbol contains '.'")
	ERR_VALUENOTFUNCTION = errors.New("symbol value not function")
)

// Wrapper for go object
type GoObject struct {
	obj   interface{}
	value reflect.Value
}

func (gobj *GoObject) iter() bool {
	kind := gobj.value.Kind()

	switch kind {
	case reflect.Slice, reflect.Array:
		return true
	default:
		return false
	}
}

func (gobj *GoObject) Inspect() string  { return fmt.Sprint(gobj.obj) }
func (gobj *GoObject) Type() ObjectType { return GO_OBJ }

func (gobj *GoObject) CallMethod(line string, scope *Scope, method string, args ...Object) Object {
	methodValue := gobj.value.MethodByName(method)
	if !methodValue.IsValid() {
		return newError(line, ERR_NOMETHOD, method, gobj.Type())
	}

	return callGoMethod(line, methodValue, args...)
}

func NewGoObject(obj interface{}) *GoObject {
	return &GoObject{obj: obj, value: reflect.ValueOf(obj)}
}

// wrapper for go functions
type GoFuncObject struct {
	name string
	typ  reflect.Type
	fn   interface{}
}

func (gfn *GoFuncObject) Inspect() string  { return gfn.name }
func (gfn *GoFuncObject) Type() ObjectType { return GFO_OBJ }

func (gfn *GoFuncObject) CallMethod(line string, scope *Scope, method string, args ...Object) Object {
	return callGoMethod(line, reflect.ValueOf(gfn.fn), args...)
}

func NewGoFuncObject(fname string, fn interface{}) *GoFuncObject {
	return &GoFuncObject{fname, reflect.TypeOf(fn), fn}
}

// Magpie language Object to go language Value.
func ObjectToGoValue(obj Object, typ reflect.Type) reflect.Value {
	var v reflect.Value
	switch obj := obj.(type) {
	case *Number:
		switch typ.Kind() {
		case reflect.Int:
			v = reflect.ValueOf(int(obj.Value))
		case reflect.Int8:
			v = reflect.ValueOf(int8(obj.Value))
		case reflect.Int16:
			v = reflect.ValueOf(int16(obj.Value))
		case reflect.Int32:
			v = reflect.ValueOf(int32(obj.Value))
		case reflect.Int64:
			v = reflect.ValueOf(int64(obj.Value))
		case reflect.Uint:
			v = reflect.ValueOf(uint(obj.Value))
		case reflect.Uint8:
			v = reflect.ValueOf(uint8(obj.Value))
		case reflect.Uint16:
			v = reflect.ValueOf(uint16(obj.Value))
		case reflect.Uint32:
			v = reflect.ValueOf(uint32(obj.Value))
		case reflect.Uint64:
			v = reflect.ValueOf(uint64(obj.Value))
		default:
			v = reflect.ValueOf(obj.Value)
		}
	case *String:
		v = reflect.ValueOf(obj.String)
	case *Boolean:
		v = reflect.ValueOf(obj.Bool)
	case *Nil:
		v = reflect.ValueOf(nil)
	case *GoObject:
		v = obj.value
	default:
		v = reflect.ValueOf(obj)
	}
	return v
}

// Go language Value to magpie language Object(take care of slice object value)
func goValueToObject(v interface{}) Object {
	val := reflect.ValueOf(v)
	kind := val.Kind()

	switch kind {
	case reflect.Slice, reflect.Array:
		ret := &Array{}
		for i := 0; i < val.Len(); i++ {
			ret.Members = append(ret.Members, goValueToObject(val.Index(i).Interface()))
		}
		return ret
	case reflect.String:
		return NewString(val.String())
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return NewNumber(float64(val.Int()))
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return NewNumber(float64(val.Uint()))
	case reflect.Float32, reflect.Float64:
		return NewNumber(val.Float())
	case reflect.Bool:
		if v.(bool) {
			return TRUE
		} else {
			return FALSE
		}
	case reflect.Invalid: //nil
		return NIL
	default:
		return NewGoObject(v)
	}
}

func callGoMethod(line string, methodVal reflect.Value, args ...Object) (ret Object) {
	defer func() {
		if r := recover(); r != nil {
			ret = newError(line, "error calling go method. %s", r)
		}
	}()

	methodType := methodVal.Type()
	//process arguments
	callArgs := []reflect.Value{}
	for i := 0; i < len(args); i++ {
		reqTyp := methodType.In(i)
		callArgs = append(callArgs, ObjectToGoValue(args[i], reqTyp))
	}

	retValues := methodVal.Call(callArgs) //call go method
	//handling return value
	var results []Object
	for _, retVal := range retValues {
		switch retVal.Kind() {
		case reflect.Invalid: //nil
			results = append(results, NIL)
		case reflect.Bool:
			results = append(results, NewBooleanObj(retVal.Bool()))
		case reflect.String:
			results = append(results, NewString(retVal.String()))
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			results = append(results, NewNumber(float64(retVal.Int())))
		case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
			results = append(results, NewNumber(float64(retVal.Uint())))
		case reflect.Float64, reflect.Float32:
			results = append(results, NewNumber(retVal.Float()))
		default:
			results = append(results, NewGoObject(retVal.Interface()))
		}
	}

	if len(results) == 1 { //one return value
		ret = results[0]
	} else if len(results) > 1 { //multiple return values
		ret = &Tuple{Members: results, IsMulti: true}
	} else { //no return value
		ret = NIL
	}

	return
}

func RegisterGoVars(name string, vars map[string]interface{}) error {
	for k, v := range vars {
		if strings.Contains(k, ".") {
			return ERR_HASDOT
		}
		SetGlobalObj(name+"."+k, NewGoObject(v))
	}

	return nil
}

func RegisterGoFunctions(name string, vars map[string]interface{}) error {
	hash := NewHash()
	for k, v := range vars {
		val := reflect.ValueOf(v)
		if val.Kind() != reflect.Func { //if v is not a function
			return ERR_VALUENOTFUNCTION
		}

		if strings.Contains(k, ".") {
			return ERR_HASDOT
		}

		key := NewString(k)
		hash.push("", key, NewGoFuncObject(k, v))
	}

	//Replace all '/' to '_'.
	newName := strings.Replace(name, "/", "_", -1)
	SetGlobalObj(newName, hash)

	return nil
}

package eval

import (
	"fmt"
	"magpie/ast"
	"math"
	"strings"
	"unicode/utf8"
)

func Eval(node ast.Node, scope *Scope) (val Object) {
	//fmt.Printf("node.Type=%T, node=<%s>, start=%d, end=%d\n", node, node.String(), node.Pos().Line, node.End().Line) //debugging
	switch node := node.(type) {
	case *ast.Program:
		return evalProgram(node, scope)
	case *ast.BlockStatement:
		return evalBlockStatement(node, scope)
	case *ast.ExpressionStatement:
		return Eval(node.Expression, scope)
	case *ast.NumberLiteral:
		return evalNumber(node, scope)
	case *ast.StringLiteral:
		return evalStringLiteral(node, scope)
	case *ast.FunctionLiteral:
		return evalFunctionLiteral(node, scope)
	case *ast.CallExpression:
		return evalCallExpression(node, scope)
	case *ast.MethodCallExpression:
		return evalMethodCallExpression(node, scope)
	case *ast.PrefixExpression:
		right := Eval(node.Right, scope)
		if isError(right) {
			return right
		}
		return evalPrefixExpression(node, right, scope)
	case *ast.InfixExpression:
		left := Eval(node.Left, scope)
		if isError(left) {
			return left
		}
		right := Eval(node.Right, scope)
		if isError(right) {
			return right
		}
		return evalInfixExpression(node, left, right, scope)
	case *ast.PostfixExpression:
		left := Eval(node.Left, scope)
		if left.Type() == ERROR_OBJ {
			return left
		}
		return evalPostfixExpression(node, left, scope)
	case *ast.BooleanLiteral:
		return nativeBoolToBooleanObject(node.Value)
	case *ast.NilLiteral:
		return NIL
	case *ast.ArrayLiteral:
		members := evalExpressions(node.Members, scope)
		if len(members) == 1 && isError(members[0]) {
			return members[0]
		}

		return &Array{Members: members}
	case *ast.IndexExpression:
		left := Eval(node.Left, scope)
		if isError(left) {
			return left
		}

		index := Eval(node.Index, scope)
		if isError(index) {
			return index
		}

		return evalIndexExpression(node, left, index)
	case *ast.HashLiteral:
		return evalHashLiteral(node, scope)
	case *ast.TupleLiteral:
		members := evalExpressions(node.Members, scope)
		if len(members) == 1 && isError(members[0]) {
			return members[0]
		}

		return &Tuple{Members: members}
	case *ast.LetStatement:
		return evalLetStatement(node, scope)
	case *ast.ReturnStatement:
		return evalReturnStatement(node, scope)
	case *ast.Identifier:
		return evalIdentifier(node, scope)
	case *ast.IfExpression:
		return evalIfExpression(node, scope)
	case *ast.AssignExpression:
		return evalAssignExpression(node, scope)
	case *ast.BreakExpression:
		return BREAK
	case *ast.ContinueExpression:
		return CONTINUE
	case *ast.CForLoop:
		return evalCForLoopExpression(node, scope)
	case *ast.ForEverLoop:
		return evalForEverLoopExpression(node, scope)
	case *ast.ForEachArrayLoop:
		return evalForEachArrayExpression(node, scope)
	case *ast.ForEachMapLoop:
		return evalForEachMapExpression(node, scope)
	case *ast.DoLoop:
		return evalDoLoopExpression(node, scope)
	case *ast.WhileLoop:
		return evalWhileLoopExpression(node, scope)
	}

	return nil
}

func evalProgram(program *ast.Program, scope *Scope) (results Object) {
	for _, stmt := range program.Statements {
		results = Eval(stmt, scope)
		if returnValue, ok := results.(*ReturnValue); ok {
			return returnValue.Value
		}
		if errObj, ok := results.(*Error); ok {
			return errObj
		}
	}

	if results == nil {
		return NIL
	}
	return results
}

func evalBlockStatement(block *ast.BlockStatement, scope *Scope) Object {
	var result Object = NIL
	for _, statement := range block.Statements {
		result = Eval(statement, scope)
		if result != nil {
			rt := result.Type()
			if rt == RETURN_VALUE_OBJ || rt == ERROR_OBJ {
				return result
			}
		}
		if _, ok := result.(*Break); ok {
			return result
		}
		if _, ok := result.(*Continue); ok {
			return result
		}
	}
	return result
}

func evalNumber(n *ast.NumberLiteral, scope *Scope) Object {
	return NewNumber(n.Value)
}

func evalStringLiteral(s *ast.StringLiteral, scope *Scope) Object {
	return NewString(s.Value)
}

func evalFunctionLiteral(fl *ast.FunctionLiteral, scope *Scope) Object {
	fn := &Function{Literal: fl, Scope: scope}
	if fl.Name != "" {
		scope.Set(fl.Name, fn)
	}
	return fn
}

func evalPrefixExpression(node *ast.PrefixExpression, right Object, scope *Scope) Object {
	switch node.Operator {
	case "+":
		return evalPlusPrefixOperatorExpression(node, right, scope)
	case "-":
		return evalMinusPrefixOperatorExpression(node, right, scope)
	case "!":
		return evalBangOperatorExpression(node, right, scope)
	default:
		return newError(node.Pos().Sline(), ERR_PREFIXOP, node.Operator, right.Type())
	}
}

func evalPlusPrefixOperatorExpression(node *ast.PrefixExpression, right Object, scope *Scope) Object {
	if right.Type() != NUMBER_OBJ {
		return newError(node.Pos().Sline(), ERR_PREFIXOP, node.Operator, right.Type())
	}
	return right
}

func evalMinusPrefixOperatorExpression(node *ast.PrefixExpression, right Object, scope *Scope) Object {
	if right.Type() != NUMBER_OBJ {
		return newError(node.Pos().Sline(), ERR_PREFIXOP, node.Operator, right.Type())
	}
	value := right.(*Number).Value
	return NewNumber(-value)
}

func evalBangOperatorExpression(node *ast.PrefixExpression, right Object, scope *Scope) Object {
	return nativeBoolToBooleanObject(!IsTrue(right))
}

func evalInfixExpression(node *ast.InfixExpression, left, right Object, scope *Scope) Object {
	operator := node.Operator
	switch {
	case operator == "&&":
		leftCond := objectToNativeBoolean(left)
		if !leftCond {
			return FALSE
		}

		rightCond := objectToNativeBoolean(right)
		return nativeBoolToBooleanObject(leftCond && rightCond)
	case operator == "||":
		leftCond := objectToNativeBoolean(left)
		if leftCond {
			return TRUE
		}

		rightCond := objectToNativeBoolean(right)
		return nativeBoolToBooleanObject(leftCond || rightCond)
	case left.Type() == NUMBER_OBJ && right.Type() == NUMBER_OBJ:
		return evalNumberInfixExpression(node, left, right, scope)
	case left.Type() == STRING_OBJ && right.Type() == STRING_OBJ:
		return evalStringInfixExpression(node, left, right, scope)
	case operator == "==":
		return nativeBoolToBooleanObject(left == right)
	case operator == "!=":
		return nativeBoolToBooleanObject(left != right)
	default:
		return newError(node.Pos().Sline(), ERR_INFIXOP, left.Type(), node.Operator, right.Type())
	}
}

func evalStringInfixExpression(node *ast.InfixExpression, left, right Object, scope *Scope) Object {
	leftVal := left.(*String).String
	rightVal := right.(*String).String

	switch node.Operator {
	case "+":
		return NewString(leftVal + rightVal)
	case "<":
		return nativeBoolToBooleanObject(leftVal < rightVal)
	case "<=":
		return nativeBoolToBooleanObject(leftVal <= rightVal)
	case ">":
		return nativeBoolToBooleanObject(leftVal > rightVal)
	case ">=":
		return nativeBoolToBooleanObject(leftVal >= rightVal)
	case "==":
		return nativeBoolToBooleanObject(leftVal == rightVal)
	case "!=":
		return nativeBoolToBooleanObject(leftVal != rightVal)
	default:
		return newError(node.Pos().Sline(), ERR_INFIXOP, left.Type(), node.Operator, right.Type())
	}
}

func evalNumberInfixExpression(node *ast.InfixExpression, left, right Object, scope *Scope) Object {
	leftVal := left.(*Number).Value
	rightVal := right.(*Number).Value

	switch node.Operator {
	case "+":
		return &Number{Value: leftVal + rightVal}
	case "-":
		return &Number{Value: leftVal - rightVal}
	case "*":
		return &Number{Value: leftVal * rightVal}
	case "/":
		if rightVal == 0 {
			return newError(node.Pos().Sline(), ERR_DIVIDEBYZERO)
		}
		return &Number{Value: leftVal / rightVal}
	case "%":
		v := math.Mod(leftVal, rightVal)
		return &Number{Value: v}
	case "**":
		return &Number{Value: math.Pow(leftVal, rightVal)}
	case "<":
		return nativeBoolToBooleanObject(leftVal < rightVal)
	case "<=":
		return nativeBoolToBooleanObject(leftVal <= rightVal)
	case ">":
		return nativeBoolToBooleanObject(leftVal > rightVal)
	case ">=":
		return nativeBoolToBooleanObject(leftVal >= rightVal)
	case "==":
		return nativeBoolToBooleanObject(leftVal == rightVal)
	case "!=":
		return nativeBoolToBooleanObject(leftVal != rightVal)
	default:
		return newError(node.Pos().Sline(), ERR_INFIXOP, left.Type(), node.Operator, right.Type())
	}
}

func evalPostfixExpression(node *ast.PostfixExpression, left Object, scope *Scope) Object {
	switch node.Operator {
	case "++":
		return evalIncrementPostfixExpression(node, left, scope)
	case "--":
		return evalDecrementPostfixExpression(node, left, scope)
	default:
		return newError(node.Pos().Sline(), ERR_POSTFIXOP, node.Operator, left.Type())
	}
}

func evalIncrementPostfixExpression(node *ast.PostfixExpression, left Object, scope *Scope) Object {
	switch left.Type() {
	case NUMBER_OBJ:
		leftObj := left.(*Number)
		returnVal := NewNumber(leftObj.Value)
		scope.Set(node.Left.String(), NewNumber(leftObj.Value+1))
		return returnVal
	default:
		return newError(node.Pos().Sline(), ERR_POSTFIXOP, node.Operator, left.Type())
	}
}

func evalDecrementPostfixExpression(node *ast.PostfixExpression, left Object, scope *Scope) Object {
	switch left.Type() {
	case NUMBER_OBJ:
		leftObj := left.(*Number)
		returnVal := NewNumber(leftObj.Value)
		scope.Set(node.Left.String(), NewNumber(leftObj.Value-1))
		return returnVal
	default:
		return newError(node.Pos().Sline(), ERR_POSTFIXOP, node.Operator, left.Type())
	}
}

func evalLetStatement(l *ast.LetStatement, scope *Scope) (val Object) {
	values := []Object{}
	valuesLen := 0
	for _, value := range l.Values {
		val := Eval(value, scope)
		if val.Type() == TUPLE_OBJ {
			tupleObj := val.(*Tuple)
			if tupleObj.IsMulti {
				valuesLen += len(tupleObj.Members)
				values = append(values, tupleObj.Members...)

			} else {
				valuesLen += 1
				values = append(values, tupleObj)
			}

		} else {
			valuesLen += 1
			values = append(values, val)
		}
	}

	for idx, item := range l.Names {
		if idx >= valuesLen { //There are more Names than Values
			if item.Token.Literal != "_" {
				val = NIL
				scope.Set(item.String(), val)
			}
		} else {
			if item.Token.Literal == "_" { // _: placeholder
				continue
			}
			val = values[idx]
			if val.Type() != ERROR_OBJ {
				scope.Set(item.String(), val)
			} else {
				return
			}
		}
	}

	return
}

func evalReturnStatement(r *ast.ReturnStatement, scope *Scope) Object {
	if r.ReturnValue == nil { //no return value, we default return `NIL` object
		return &ReturnValue{Value: NIL, Values: []Object{NIL}}
	}

	ret := &ReturnValue{}
	for _, value := range r.ReturnValues {
		ret.Values = append(ret.Values, Eval(value, scope))
	}

	// for old campatibility
	ret.Value = ret.Values[0]

	return ret
}

func evalIdentifier(node *ast.Identifier, scope *Scope) Object {
	if val, ok := scope.Get(node.Value); ok {
		return val
	}

	if builtin, ok := builtins[node.Value]; ok {
		return builtin
	}

	return newError(node.Pos().Sline(), ERR_UNKNOWNIDENT, node.Value)

}

func evalIfExpression(ie *ast.IfExpression, scope *Scope) Object {
	//eval "if/else-if" part
	for _, c := range ie.Conditions {
		condition := Eval(c.Cond, scope)
		if condition.Type() == ERROR_OBJ {
			return condition
		}

		if IsTrue(condition) {
			return evalBlockStatement(c.Body, scope)
		}
	}

	//eval "else" part
	if ie.Alternative != nil {
		return evalBlockStatement(ie.Alternative, scope)
	}

	return NIL
}

func evalExpressions(exps []ast.Expression, scope *Scope) []Object {
	var result []Object
	for _, e := range exps {
		evaluated := Eval(e, scope)
		if isError(evaluated) {
			return []Object{evaluated}
		}

		result = append(result, evaluated)
	}

	return result
}

func evalIndexExpression(node *ast.IndexExpression, left, index Object) Object {
	switch {
	case left.Type() == STRING_OBJ:
		return evalStringIndex(node.Pos().Sline(), left, index)
	case left.Type() == ARRAY_OBJ:
		return evalArrayIndexExpression(node.Pos().Sline(), left, index)
	case left.Type() == HASH_OBJ:
		return evalHashIndexExpression(node.Pos().Sline(), left, index)
	case left.Type() == TUPLE_OBJ:
		return evalTupleIndexExpression(node.Pos().Sline(), left, index)
	default:
		return newError(node.Pos().Sline(), ERR_NOINDEXABLE, left.Type())
	}
}

func evalStringIndex(line string, left, index Object) Object {
	str := left.(*String)

	idx := int64(index.(*Number).Value)
	max := int64(utf8.RuneCountInString(str.String)) - 1
	if idx < 0 || idx > max {
		return newError(line, ERR_INDEX, idx)
	}

	return NewString(string([]rune(str.String)[idx])) //support utf8,not very efficient
}

func evalArrayIndexExpression(line string, array, index Object) Object {
	arrayObject := array.(*Array)
	idx := int64(index.(*Number).Value)
	max := int64(len(arrayObject.Members) - 1)
	if idx < 0 || idx > max {
		return newError(line, ERR_INDEX, idx)
	}

	return arrayObject.Members[idx]
}

//Almost same as evalArrayIndexExpression
func evalTupleIndexExpression(line string, tuple, index Object) Object {
	tupleObject := tuple.(*Tuple)
	idx := int64(index.(*Number).Value)
	max := int64(len(tupleObject.Members) - 1)
	if idx < 0 || idx > max {
		return newError(line, ERR_INDEX, idx)
	}

	return tupleObject.Members[idx]
}

func evalHashIndexExpression(line string, hash, index Object) Object {
	hashObject := hash.(*Hash)
	key, ok := index.(Hashable)
	if !ok {
		return newError(line, ERR_KEY, index.Type())
	}

	pair, ok := hashObject.Pairs[key.HashKey()]
	if !ok {
		return NIL
	}

	return pair.Value
}

func evalHashLiteral(node *ast.HashLiteral, scope *Scope) Object {
	pairs := make(map[HashKey]HashPair)
	for keyNode, valueNode := range node.Pairs {
		key := Eval(keyNode, scope)
		if isError(key) {
			return key
		}

		hashKey, ok := key.(Hashable)
		if !ok {
			return newError(node.Pos().Sline(), ERR_KEY, key.Type())
		}

		value := Eval(valueNode, scope)
		if isError(value) {
			return value
		}

		hashed := hashKey.HashKey()
		pairs[hashed] = HashPair{Key: key, Value: value}
	}

	return &Hash{Pairs: pairs}
}

func evalMethodCallExpression(call *ast.MethodCallExpression, scope *Scope) Object {
	obj := Eval(call.Object, scope)
	if obj.Type() == ERROR_OBJ {
		return obj
	}

	if method, ok := call.Call.(*ast.CallExpression); ok {
		args := evalExpressions(method.Arguments, scope)
		if len(args) == 1 && isError(args[0]) {
			return args[0]
		}

		return obj.CallMethod(call.Call.Pos().Sline(), scope, method.Function.String(), args...)
	}

	return newError(call.Call.Pos().Sline(), ERR_NOMETHOD, call.String(), obj.Type())
}

func evalAssignExpression(a *ast.AssignExpression, scope *Scope) Object {
	val := Eval(a.Value, scope)
	if val.Type() == ERROR_OBJ {
		return val
	}

	var name string
	switch nodeType := a.Name.(type) {
	//a = 10
	case *ast.Identifier:
		name = nodeType.Value

	//arr[idx] = "xxx", here `a.Name` = arr[idx]
	case *ast.IndexExpression:
		switch nodeType.Left.(type) {
		case *ast.Identifier:
			name = nodeType.Left.(*ast.Identifier).Value //here, name = arr
		}
	case *ast.MethodCallExpression:
		name = nodeType.Object.String()
	}

	if a.Token.Literal == "=" {
		switch nodeType := a.Name.(type) {
		case *ast.Identifier: //e.g. a = "hello"
			scope.Set(nodeType.Value, val)
			return val
		}
	}

	// Check if the variable exists or not
	var left Object
	var ok bool
	if left, ok = scope.Get(name); !ok {
		return newError(a.Pos().Sline(), ERR_UNKNOWNIDENT, name)
	}

	switch left.Type() {
	case NUMBER_OBJ:
		return evalNumAssignExpression(a, name, left, scope, val)
	case STRING_OBJ:
		return evalStrAssignExpression(a, name, left, scope, val)
	case ARRAY_OBJ:
		return evalArrayAssignExpression(a, name, left, scope, val)
	case TUPLE_OBJ:
		return evalTupleAssignExpression(a, name, left, scope, val)
	case HASH_OBJ:
		return evalHashAssignExpression(a, name, left, scope, val)
	}

	return newError(a.Pos().Sline(), ERR_INFIXOP, left.Type(), a.Token.Literal, val.Type())
}

// num += num
// num -= num
// etc...
func evalNumAssignExpression(a *ast.AssignExpression, name string, left Object, scope *Scope, val Object) (ret Object) {
	switch a.Token.Literal {
	case "+=":
	case "-=":
	}
	return newError(a.Pos().Sline(), ERR_INFIXOP, left.Type(), a.Token.Literal, val.Type())
}

//str[idx] = item
//str += item
func evalStrAssignExpression(a *ast.AssignExpression, name string, left Object, scope *Scope, val Object) (ret Object) {
	leftVal := left.(*String).String

	switch a.Token.Literal {
	case "=":
		switch nodeType := a.Name.(type) {
		case *ast.IndexExpression: //str[idx] = xxx
			index := Eval(nodeType.Index, scope)
			if index == NIL {
				ret = NIL
				return
			}

			idx := int64(index.(*Number).Value)

			if idx < 0 || idx >= int64(len(leftVal)) {
				return newError(a.Pos().Sline(), ERR_INDEX, idx)
			}

			ret = NewString(leftVal[:idx] + val.Inspect() + leftVal[idx+1:])
			scope.Set(name, ret)
			return
		}
	}

	return newError(a.Pos().Sline(), ERR_INFIXOP, left.Type(), a.Token.Literal, val.Type())
}

//array[idx] = item
//array += item
func evalArrayAssignExpression(a *ast.AssignExpression, name string, left Object, scope *Scope, val Object) (ret Object) {
	leftVals := left.(*Array).Members

	switch a.Token.Literal {
	case "=":
		switch nodeType := a.Name.(type) {
		case *ast.IndexExpression: //arr[idx] = xxx
			index := Eval(nodeType.Index, scope)
			if index == NIL {
				ret = NIL
				return
			}

			idx := int64(index.(*Number).Value)
			if idx < 0 {
				return newError(a.Pos().Sline(), ERR_INDEX, idx)
			}

			if idx < int64(len(leftVals)) { //index is in range
				leftVals[idx] = val
				ret = &Array{Members: leftVals}
				scope.Set(name, ret)
				return
			} else { //index is out of range, we auto-expand the array
				for i := int64(len(leftVals)); i < idx; i++ {
					leftVals = append(leftVals, NIL)
				}

				leftVals = append(leftVals, val)
				ret = &Array{Members: leftVals}
				scope.Set(name, ret)
				return
			}
		}

		return newError(a.Pos().Sline(), ERR_INFIXOP, left.Type(), a.Token.Literal, val.Type())
	}

	return newError(a.Pos().Sline(), ERR_INFIXOP, left.Type(), a.Token.Literal, val.Type())
}

//tuple element can not be assigned
func evalTupleAssignExpression(a *ast.AssignExpression, name string, left Object, scope *Scope, val Object) (ret Object) {
	//Tuple is an immutable sequence of values
	if a.Token.Literal == "=" { //tuple[idx] = item
		str := fmt.Sprintf("%s[IDX]", TUPLE_OBJ)
		return newError(a.Pos().Sline(), ERR_INFIXOP, str, a.Token.Literal, val.Type())
	}
	return newError(a.Pos().Sline(), ERR_INFIXOP, left.Type(), a.Token.Literal, val.Type())
}

//hash[key] = value
func evalHashAssignExpression(a *ast.AssignExpression, name string, left Object, scope *Scope, val Object) (ret Object) {
	leftHash := left.(*Hash)

	switch a.Token.Literal {
	case "=":
		switch nodeType := a.Name.(type) {
		case *ast.IndexExpression: //hashObj[key] = val
			key := Eval(nodeType.Index, scope)
			leftHash.push(a.Pos().Sline(), key, val)
			return leftHash
		case *ast.Identifier: //hashObj.key = val
			key := strings.Split(a.Name.String(), ".")[1]
			keyObj := NewString(key)
			leftHash.push(a.Pos().Sline(), keyObj, val)
			return leftHash
		}
		return newError(a.Pos().Sline(), ERR_INFIXOP, left.Type(), a.Token.Literal, val.Type())
	}

	return newError(a.Pos().Sline(), ERR_INFIXOP, left.Type(), a.Token.Literal, val.Type())
}

//for (init; condition; updater) { block }
// returns the last expression value or NIL
func evalCForLoopExpression(fl *ast.CForLoop, scope *Scope) Object { //fl:For Loop
	if fl.Init != nil {
		init := Eval(fl.Init, scope)
		if init.Type() == ERROR_OBJ {
			return init
		}
	}

	var result Object = NIL
	for {
		//condition
		var condition Object = NIL
		if fl.Cond != nil {
			condition = Eval(fl.Cond, scope)
			if condition.Type() == ERROR_OBJ {
				return condition
			}
			if !IsTrue(condition) {
				break
			}
		}

		//body
		result = Eval(fl.Block, scope)
		if result.Type() == ERROR_OBJ {
			return result
		}

		if _, ok := result.(*Break); ok {
			break
		}
		if _, ok := result.(*Continue); ok {
			if fl.Update != nil {
				newVal := Eval(fl.Update, scope) //Before continue, we need to call 'Update'
				if newVal.Type() == ERROR_OBJ {
					return newVal
				}
			}

			continue
		}
		if v, ok := result.(*ReturnValue); ok {
			return v
		}

		if fl.Update != nil {
			newVal := Eval(fl.Update, scope)
			if newVal.Type() == ERROR_OBJ {
				return newVal
			}
		}
	}

	if result == nil || result.Type() == BREAK_OBJ || result.Type() == CONTINUE_OBJ {
		return NIL
	}

	return result
}

// for { block }
// returns the last expression value or NIL
func evalForEverLoopExpression(fel *ast.ForEverLoop, scope *Scope) Object {
	var e Object = NIL
	for {
		e = Eval(fel.Block, scope)
		if e.Type() == ERROR_OBJ {
			return e
		}

		if _, ok := e.(*Break); ok {
			break
		}
		if _, ok := e.(*Continue); ok {
			continue
		}
		if v, ok := e.(*ReturnValue); ok {
			return v
		}
	}

	if e == nil || e.Type() == BREAK_OBJ || e.Type() == CONTINUE_OBJ {
		return NIL
	}

	return e
}

//for item in array
//for item in string
//for item in tuple
//returns an Array-object or a Return-object
func evalForEachArrayExpression(fal *ast.ForEachArrayLoop, scope *Scope) Object { //fal:For Array Loop
	aValue := Eval(fal.Value, scope)
	if aValue.Type() == ERROR_OBJ {
		return &Array{Members: []Object{aValue}}
	}

	//first check if it's a Nil object
	if aValue.Type() == NIL_OBJ {
		return &Array{Members: []Object{}} //return empty array
	}

	iterObj, ok := aValue.(Iterable)
	if !ok {
		errObj := newError(fal.Pos().Sline(), ERR_NOTITERABLE)
		return &Array{Members: []Object{errObj}}
	}
	if !iterObj.iter() {
		errObj := newError(fal.Pos().Sline(), ERR_NOTITERABLE)
		return &Array{Members: []Object{errObj}}
	}

	var members []Object
	if aValue.Type() == STRING_OBJ {
		aStr, _ := aValue.(*String)
		runes := []rune(aStr.String)
		for _, rune := range runes {
			members = append(members, NewString(string(rune)))
		}
	} else if aValue.Type() == ARRAY_OBJ {
		arr, _ := aValue.(*Array)
		members = arr.Members
	} else if aValue.Type() == TUPLE_OBJ {
		tuple, _ := aValue.(*Tuple)
		members = tuple.Members
	}

	if len(members) == 0 {
		return &Array{Members: []Object{}} //return empty array
	}

	arr := &Array{}
	defer func() {
		scope.Del(fal.Var)
	}()
	for _, value := range members {
		scope.Set(fal.Var, value)

		result := Eval(fal.Block, scope)
		if result.Type() == ERROR_OBJ {
			arr.Members = append(arr.Members, result)
			return arr
		}

		if _, ok := result.(*Break); ok {
			break
		}
		if _, ok := result.(*Continue); ok {
			continue
		}
		if v, ok := result.(*ReturnValue); ok {
			return v
		} else {
			arr.Members = append(arr.Members, result)
		}
	}

	return arr
}

//for index, value in string
//for index, value in array
//for index, value in tuple
//returns an Array-object or a Return-object
func evalForEachArrayWithIndex(fml *ast.ForEachMapLoop, val Object, scope *Scope) Object {
	var members []Object
	if val.Type() == STRING_OBJ {
		aStr, _ := val.(*String)
		runes := []rune(aStr.String)
		for _, rune := range runes {
			members = append(members, NewString(string(rune)))
		}
	} else if val.Type() == ARRAY_OBJ {
		arr, _ := val.(*Array)
		members = arr.Members
	} else if val.Type() == TUPLE_OBJ {
		tuple, _ := val.(*Tuple)
		members = tuple.Members
	}

	if len(members) == 0 {
		return &Array{Members: []Object{}} //return empty error
	}

	arr := &Array{}
	defer func() {
		if fml.Key != "_" {
			scope.Del(fml.Key)
		}
		if fml.Value != "_" {
			scope.Del(fml.Value)
		}
	}()
	for idx, value := range members {
		if fml.Key != "_" {
			scope.Set(fml.Key, NewNumber(float64(idx)))
		}
		if fml.Value != "_" {
			scope.Set(fml.Value, value)
		}

		result := Eval(fml.Block, scope)
		if result.Type() == ERROR_OBJ {
			arr.Members = append(arr.Members, result)
			return arr
		}

		if _, ok := result.(*Break); ok {
			break
		}
		if _, ok := result.(*Continue); ok {
			continue
		}
		if v, ok := result.(*ReturnValue); ok {
			return v
		} else {
			arr.Members = append(arr.Members, result)
		}
	}

	return arr
}

//for k, v in X { block }
//returns an Array-object or a Return-object
func evalForEachMapExpression(fml *ast.ForEachMapLoop, scope *Scope) Object { //fml:For Map Loop
	aValue := Eval(fml.X, scope)
	if aValue.Type() == ERROR_OBJ {
		return &Array{Members: []Object{aValue}}
	}

	//first check if it's a Nil object
	if aValue.Type() == NIL_OBJ {
		//return an empty array object
		return &Array{Members: []Object{}}
	}

	iterObj, ok := aValue.(Iterable)
	if !ok {
		errObj := newError(fml.Pos().Sline(), ERR_NOTITERABLE)
		return &Array{Members: []Object{errObj}}
	}
	if !iterObj.iter() {
		errObj := newError(fml.Pos().Sline(), ERR_NOTITERABLE)
		return &Array{Members: []Object{errObj}}
	}

	//for index, value in arr
	//for index, value in string
	//for index, value in tuple
	if aValue.Type() == STRING_OBJ || aValue.Type() == ARRAY_OBJ || aValue.Type() == TUPLE_OBJ {
		return evalForEachArrayWithIndex(fml, aValue, scope)
	}

	hash, _ := aValue.(*Hash)
	if len(hash.Pairs) == 0 { //hash is empty
		return &Array{Members: []Object{}}
	}

	arr := &Array{}
	defer func() {
		if fml.Key != "_" {
			scope.Del(fml.Key)
		}
		if fml.Value != "_" {
			scope.Del(fml.Value)
		}
	}()

	for _, pair := range hash.Pairs {
		if fml.Key != "_" {
			scope.Set(fml.Key, pair.Key)
		}
		if fml.Value != "_" {
			scope.Set(fml.Value, pair.Value)
		}

		result := Eval(fml.Block, scope)
		if result.Type() == ERROR_OBJ {
			arr.Members = append(arr.Members, result)
			return arr
		}

		if _, ok := result.(*Break); ok {
			break
		}
		if _, ok := result.(*Continue); ok {
			continue
		}
		if v, ok := result.(*ReturnValue); ok {
			return v
		} else {
			arr.Members = append(arr.Members, result)
		}
	}

	return arr
}

//do { block }
// returns the last expression value or NIL
func evalDoLoopExpression(dl *ast.DoLoop, scope *Scope) Object {
	var e Object = NIL
	for {
		e = Eval(dl.Block, scope)
		if e.Type() == ERROR_OBJ {
			return e
		}

		if _, ok := e.(*Break); ok {
			break
		}
		if _, ok := e.(*Continue); ok {
			continue
		}
		if v, ok := e.(*ReturnValue); ok {
			return v
		}
	}

	if e == nil || e.Type() == BREAK_OBJ || e.Type() == CONTINUE_OBJ {
		return NIL
	}

	return e
}

//while condition { block }
// returns the last expression value or NIL
func evalWhileLoopExpression(wl *ast.WhileLoop, scope *Scope) Object {
	var result Object = NIL
	for {
		condition := Eval(wl.Condition, scope)
		if condition.Type() == ERROR_OBJ {
			return condition
		}

		if !IsTrue(condition) {
			return NIL
		}

		result = Eval(wl.Block, scope)
		if result.Type() == ERROR_OBJ {
			return result
		}

		if _, ok := result.(*Break); ok {
			break
		}
		if _, ok := result.(*Continue); ok {
			continue
		}
		if v, ok := result.(*ReturnValue); ok {
			return v
		}
	}

	if result == nil || result.Type() == BREAK_OBJ || result.Type() == CONTINUE_OBJ {
		return NIL
	}

	return result
}

func evalCallExpression(node *ast.CallExpression, scope *Scope) Object {
	args := evalExpressions(node.Arguments, scope)
	if len(args) == 1 && isError(args[0]) {
		return args[0]
	}

	function := Eval(node.Function, scope)
	if isError(function) {
		return function
	}

	return applyFunction(node.Pos().Sline(), scope, function, args)
}

func applyFunction(line string, scope *Scope, fn Object, args []Object) Object {
	switch fn := fn.(type) {
	case *Function:
		extendedScope := extendFunctionScope(fn, args)
		evaluated := Eval(fn.Literal.Body, extendedScope)
		return unwrapReturnValue(evaluated)
	case *Builtin:
		return fn.Fn(line, scope, args...)
	default:
		return newError(line, ERR_NOTFUNCTION, fn.Type())
	}
}

func extendFunctionScope(fn *Function, args []Object) *Scope {
	scope := NewScope(fn.Scope, nil)
	for paramIdx, param := range fn.Literal.Parameters {
		scope.Set(param.Value, args[paramIdx])
	}

	return scope
}

func unwrapReturnValue(obj Object) Object {
	if returnValue, ok := obj.(*ReturnValue); ok {
		// if function returns multiple-values
		// returns a tuple instead.
		if len(returnValue.Values) > 1 {
			return &Tuple{Members: returnValue.Values, IsMulti: true}
		}
		return returnValue.Value
	}

	return obj
}

func objectToNativeBoolean(o Object) bool {
	if r, ok := o.(*ReturnValue); ok {
		o = r.Value
	}
	switch obj := o.(type) {
	case *Boolean:
		return obj.Bool
	case *Nil:
		return false
	case *Number:
		if obj.Value == 0.0 {
			return false
		}
		return true
	case *String:
		return obj.String != ""
	case *Array:
		if len(obj.Members) == 0 {
			return false
		}
		return true
	case *Tuple:
		if len(obj.Members) == 0 {
			return false
		}
		return true
	case *Hash:
		if len(obj.Pairs) == 0 {
			return false
		}
		return true
	default:
		return true
	}
}

func nativeBoolToBooleanObject(input bool) *Boolean {
	if input {
		return TRUE
	}
	return FALSE
}

func IsTrue(obj Object) bool {
	switch obj {
	case TRUE:
		return true
	case FALSE:
		return false
	case NIL:
		return false
	default:
		switch obj.Type() {
		case NUMBER_OBJ:
			if obj.(*Number).Value == 0.0 {
				return false
			}
		case ARRAY_OBJ:
			if len(obj.(*Array).Members) == 0 {
				return false
			}
		case HASH_OBJ:
			if len(obj.(*Hash).Pairs) == 0 {
				return false
			}
		case TUPLE_OBJ:
			if len(obj.(*Tuple).Members) == 0 {
				return false
			}
		}
		return true
	}
}

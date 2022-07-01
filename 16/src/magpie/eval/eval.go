package eval

import (
	"magpie/ast"
	"math"
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
	case *ast.LetStatement:
		val := Eval(node.Value, scope)
		if isError(val) {
			return val
		}
		scope.Set(node.Name.Value, val)
	case *ast.ReturnStatement:
		if node.ReturnValue == nil {
			return &ReturnValue{Value: NIL}
		}

		val := Eval(node.ReturnValue, scope)
		return &ReturnValue{Value: val}
	case *ast.Identifier:
		return evalIdentifier(node, scope)
	case *ast.IfExpression:
		return evalIfExpression(node, scope)
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
	var result Object
	for _, statement := range block.Statements {
		result = Eval(statement, scope)
		if result != nil {
			rt := result.Type()
			if rt == RETURN_VALUE_OBJ || rt == ERROR_OBJ {
				return result
			}
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
	return &Function{Literal: fl, Scope: scope}
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
		return returnValue.Value
	}

	return obj
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
		}
		return true
	}
}

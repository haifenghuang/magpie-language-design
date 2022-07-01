package eval

import (
	"magpie/ast"
	"math"
)

func Eval(node ast.Node, scope *Scope) (val Object) {

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

func evalInfixExpression(node *ast.InfixExpression, left, right Object, scope *Scope) Object {
	switch {
	case left.Type() == NUMBER_OBJ && right.Type() == NUMBER_OBJ:
		return evalNumberInfixExpression(node, left, right, scope)
	case left.Type() == STRING_OBJ && right.Type() == STRING_OBJ:
		return evalStringInfixExpression(node, left, right, scope)
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
	default:
		return newError(node.Pos().Sline(), ERR_INFIXOP, left.Type(), node.Operator, right.Type())
	}
}

func evalIdentifier(node *ast.Identifier, scope *Scope) Object {
	val, ok := scope.Get(node.Value)
	if !ok {
		return newError(node.Pos().Sline(), ERR_UNKNOWNIDENT, node.Value)
	}
	return val
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
	function, ok := fn.(*Function)
	if !ok {
		return newError(function.Literal.Pos().Sline(), ERR_NOTFUNCTION, fn.Type())
	}

	extendedScope := extendFunctionScope(function, args)
	evaluated := Eval(function.Literal.Body, extendedScope)
	return unwrapReturnValue(evaluated)
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

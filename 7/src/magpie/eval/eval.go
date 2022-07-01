package eval

import (
	"magpie/ast"
	"math"
)

func Eval(node ast.Node, scope *Scope) (val Object) {

	switch node := node.(type) {
	case *ast.Program:
		return evalProgram(node, scope)
	case *ast.ExpressionStatement:
		return Eval(node.Expression, scope)
	case *ast.NumberLiteral:
		return evalNumber(node, scope)
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
	case *ast.Identifier:
		return evalIdentifier(node, scope)
	}

	return nil
}

func evalProgram(program *ast.Program, scope *Scope) (results Object) {
	for _, stmt := range program.Statements {
		results = Eval(stmt, scope)
		if errObj, ok := results.(*Error); ok {
			return errObj
		}
	}

	if results == nil {
		return NIL
	}
	return results
}

func evalNumber(n *ast.NumberLiteral, scope *Scope) Object {
	return NewNumber(n.Value)
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

func nativeBoolToBooleanObject(input bool) *Boolean {
	if input {
		return TRUE
	}
	return FALSE
}

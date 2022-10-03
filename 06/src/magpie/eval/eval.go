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
		return evalPrefixExpression(node, right, scope)
	case *ast.InfixExpression:
		left := Eval(node.Left, scope)

		right := Eval(node.Right, scope)
		return evalInfixExpression(node, left, right, scope)
	case *ast.BooleanLiteral:
		return nativeBoolToBooleanObject(node.Value)
	case *ast.NilLiteral:
		return NIL
	case *ast.LetStatement:
		val := Eval(node.Value, scope)
		scope.Set(node.Name.Value, val)
	case *ast.Identifier:
		return evalIdentifier(node, scope)
	}

	return nil
}

func evalProgram(program *ast.Program, scope *Scope) (results Object) {
	for _, stmt := range program.Statements {
		results = Eval(stmt, scope)
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
		return nil
	}
}

func evalPlusPrefixOperatorExpression(node *ast.PrefixExpression, right Object, scope *Scope) Object {
	if right.Type() != NUMBER_OBJ {
		return nil
	}
	return right
}

func evalMinusPrefixOperatorExpression(node *ast.PrefixExpression, right Object, scope *Scope) Object {
	if right.Type() != NUMBER_OBJ {
		return nil
	}
	value := right.(*Number).Value
	return NewNumber(-value)
}

func evalInfixExpression(node *ast.InfixExpression, left, right Object, scope *Scope) Object {
	switch {
	case left.Type() == NUMBER_OBJ && right.Type() == NUMBER_OBJ:
		return evalNumberInfixExpression(node, left, right, scope)
	default:
		return nil
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
		return &Number{Value: leftVal / rightVal}
	case "%":
		v := math.Mod(leftVal, rightVal)
		return &Number{Value: v}
	case "**":
		return &Number{Value: math.Pow(leftVal, rightVal)}
	default:
		return nil
	}
}

func evalIdentifier(node *ast.Identifier, scope *Scope) Object {
	val, _ := scope.Get(node.Value)
	return val
}

func nativeBoolToBooleanObject(input bool) *Boolean {
	if input {
		return TRUE
	}
	return FALSE
}

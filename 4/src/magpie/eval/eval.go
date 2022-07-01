package eval

import (
	"magpie/ast"
	"math"
)

func Eval(node ast.Node) (val Object) {

	switch node := node.(type) {
	case *ast.Program:
		return evalProgram(node)
	case *ast.ExpressionStatement:
		return Eval(node.Expression)
	case *ast.NumberLiteral:
		return evalNumber(node)
	case *ast.PrefixExpression:
		right := Eval(node.Right)
		return evalPrefixExpression(node, right)
	case *ast.InfixExpression:
		left := Eval(node.Left)

		right := Eval(node.Right)
		return evalInfixExpression(node, left, right)
	case *ast.BooleanLiteral:
		return nativeBoolToBooleanObject(node.Value)
	case *ast.NilLiteral:
		return NIL
	}

	return nil
}

func evalProgram(program *ast.Program) (results Object) {
	for _, stmt := range program.Statements {
		results = Eval(stmt)
	}

	if results == nil {
		return NIL
	}
	return results
}

func evalNumber(n *ast.NumberLiteral) Object {
	return NewNumber(n.Value)
}

func evalPrefixExpression(node *ast.PrefixExpression, right Object) Object {
	switch node.Operator {
	case "+":
		return evalPlusPrefixOperatorExpression(node, right)
	case "-":
		return evalMinusPrefixOperatorExpression(node, right)
	default:
		return nil
	}
}

func evalPlusPrefixOperatorExpression(node *ast.PrefixExpression, right Object) Object {
	if right.Type() != NUMBER_OBJ {
		return nil
	}
	return right
}

func evalMinusPrefixOperatorExpression(node *ast.PrefixExpression, right Object) Object {
	if right.Type() != NUMBER_OBJ {
		return nil
	}
	value := right.(*Number).Value
	return NewNumber(-value)
}

func evalInfixExpression(node *ast.InfixExpression, left, right Object) Object {
	switch {
	case left.Type() == NUMBER_OBJ && right.Type() == NUMBER_OBJ:
		return evalNumberInfixExpression(node, left, right)
	default:
		return nil
	}
}

func evalNumberInfixExpression(node *ast.InfixExpression, left, right Object) Object {
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

func nativeBoolToBooleanObject(input bool) *Boolean {
	if input {
		return TRUE
	}
	return FALSE
}

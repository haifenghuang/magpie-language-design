# 追加对`let`语句和标识符的解释（Evaluating）

上一篇讲了`Scope（作用域）`的概念及其实现，有了这个概念，我们就可以来实现之前没有实现的功能了。

由于我们之前对`let`语句和标识符（Identifier）的词法解析（Lexing）和语法解析（Parsing）都已经完成了，所以这一章我们主要的任务是追加对`let`语句和标识符的解释（Evaluating）。

现在来看看我们需要做哪些更改：

1. 有了`Scope（作用域）`的概念，解释器（Evaluator）的`Eval()`函数需要加入一个`Scope`的参数。同样，其它的解释（Evaluating）函数，也需要加入这个`Scope`参数。
2. 在`Eval()`函数的`switch`分支中增加对`let`语句和标识符`Identifier`的解析。
3. 因为`Eval`函数多了一个`Scope`参数，所以调用`Eval()`函数的地方也需要更改。



下面让我们来看看`Eval()`函数的更改吧：

```go
//eval.go
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
		val := Eval(node.Value, scope) //解释`let`语句的变量值
		scope.Set(node.Name.Value, val) //将解释后的变量值放入'Scope'中
	case *ast.Identifier:
		return evalIdentifier(node, scope)
	}

	return nil
}

//标识符的解释
func evalIdentifier(node *ast.Identifier, scope *Scope) Object {
	val, _ := scope.Get(node.Value) //从Scope中取出标识符中存储的值
	return val
}
```

我们给`Eval()`函数追加了`Scope`参数（第2行），其它的`evalXXX()`函数也追加了`Scope`参数。最后我们增加了对`let`语句和`标识符(Identifier)`的解释（22-26行）。

对`let`语句的解释（Evaluating）只是简单的将`LetStatement`语句中的`<expression>`取出来进行解释（Evaluating），将解释后生成的对象（Object）放入`Scope`中。这里再次贴出`LetStatement`这个抽象语法树的代码表示：

```go
//ast.go
//let <identifier> = <expression>
type LetStatement struct {
	Token token.Token
    Name  *Identifier // 变量名: <identifier>
    Value Expression  // 变量值: <expression>
}
```

对`标识符(Identifier)`的解释（Evaluating）也是比较简单的，从`Scope`中取出标识符中储存的值即可。

> 注：对于标识符（Identifier）的解释，我们并没有判断这个变量是否存在。如果不存在的话，则需要错误处理，这个留待之后的文章中进行完善。

下面是调用`Eval()`处的更改：

```go
//main.go
func TestEval() {
	tests := []struct {
		input    string
		expected string
	}{
		{"-1 - 2.333", "-3.333"},
		{"1 + 2", "3"},
		{"2 + (3 * 4) / ( 6 - 3 ) + 10", "16"},
		{"2 + 3 * 4 / 6 - 3  + 10", "11"},
		{"(5 + 2) * (4 - 2) + 6", "20"},
		{"5 + 2 * 4 - 2 + 6", "17"},
		{"5 + 2.1 * 4 - 2 + 6.2", "17.6"},
		{"2 + 2 ** 2 ** 3", "258"},
		{"10", "10"},
		{"nil", "nil"},
		{"true", "true"},
		{"false", "false"},
		{"let x = 2 + (3 * 4) / ( 6 - 3 ) + 10; x", "16"},
	}

	for idx, tt := range tests {
		l := lexer.NewLexer(tt.input)
		p := parser.NewParser(l)
		program := p.ParseProgram()

        //新创建一个scope对象，第一个参数`parentScope`为nil，
        //第二个参数为标准输出，这个之后再详细讲有什么用处.
		scope := eval.NewScope(nil, os.Stdout)
		evaluated := eval.Eval(program, scope)
		if evaluated != nil {
			if evaluated.Inspect() != tt.expected {
				fmt.Printf("Evaluator error(%d): expected %v, got %v\n", idx, 
							tt.expected, evaluated.Inspect())
			} else {
				fmt.Printf("%s = %s\n", tt.input, tt.expected)
			}
		}
	}
}

func main() {
	TestEval()
}
```

第19行我们增加了对`let`语句及其`标识符（Identifier）`的测试。

第29-30行给`Eval()`函数追加了`Scope`参数。

恭喜，测试通过！我们的自制语言慢慢的羽翼开始丰满了。多亏了第一节中，我们在实现简单的计算器时写的【复杂的】代码框架。



前面我在注释中提到了关于变量（Identifier）找不到时的错误处理的问题。下一节，我们就来探讨、实现这个问题。


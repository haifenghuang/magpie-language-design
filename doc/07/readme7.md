# 运行期错误处理（Run-Time error handling）

当我们的语言用户写代码的时候，难免会出现各种各样的问题，比如引用了没有定义的变量，除数为零等等。这就需要我们的自制脚本语言能够处理这种运行期（Run-Time）错误。

如何给我们的自制语言提供这种错误处理呢？请读者想一想，我们怎样用统一的方式来处理这种错误？这个错误处理系统中应该包含什么样的信息呢？

我还是别卖关子了。对于运行期错误，我们需要统一的方式来处理。而对于解释器，它最后处理的实际都是对象系统中的对象（Object），例如数字对象，布尔型对象等等。对于错误处理，同样的，我们需要一个错误对象。这个错误对象，很显然的需要包含错误的信息。有了这个说明，我们的错误处理对象就有了原型：

```go
//errors.go
type Error struct {
	Message string //错误信息
}
```

很简单是吧。没错，我们的错误对象就是这样一个简单的结构。

在继续解释前，先让我们回顾一下，我们在第一节【简单计算器】的实现中，讲了关于`对象（Object）`系统的内容。当脚本中出现一个数字【3】的时候，在解释（Evaluating）阶段，实际上这个数字【3】，是存放在我们的`数字对象（Number Object）`中的：

```go
//object.go
type Number struct {
	Value float64
}
```

读者应该还有印象。总的来说，【所有在脚本中出现的数字，都是存放在这个`数字对象（Number Object）`中的。】

请仔细理解一下上面这句话。

那么对于我们的运行期错误，是不是也可以得到类似的结论：

【所有在脚本中出现的错误，都是存放在一个`错误对象（Error Object）`中的】？

说到这个份上，我想很多细心的读者已经想到了，我们的`错误结构(Error)`也应该是一个对象，什么样的对象呢？当然是对象系统中的`Object`对象。也就是说，我们的`错误结构(Error)`需要实现`Object接口`中定义的方法，使其变成一个`错误对象(Error Object)`。

下面是`错误对象(Error Object)`的具体实现：

```go
//errors.go

//错误结构
type Error struct {
	Message string
}

//下面两个方法实现了'Object'接口中定义的方法
func (e *Error) Inspect() string  { return e.Message }
func (e *Error) Type() ObjectType { return ERROR_OBJ }
```

这就是我们的`错误对象（Error Object）`的全部实现，实在是再简单不过了！！！ 

上面的代码中，第10行的`ERROR_OBJ`这个常量我们还没有定义，我们需要在`object.go`中加入这个定义：

```go
//object.go

type ObjectType string
const (
	//...
	ERROR_OBJ   = "ERROR"   //加入错误对象的类型
)
```

到这里应该说`错误对象(Error Object)`的实现已经算是完成了。为了程序中处理方便，我们还为其提供了两个工具（utility）函数：

```go
//errors.go

//创建一个错误处理对象，用来报告运行期错误。
//'line'参数是错误的行号
func newError(line string, format string, args ...interface{}) Object {
	msg := "Runtime Error at " + strings.TrimLeft(line, " \t") + "\n\t" + 
			fmt.Sprintf(format, args...) + "\n"
	return &Error{Message: msg}
}

//判断一个给定的'obj'是否是一个错误对象
func isError(obj Object) bool {
	if obj != nil {
		return obj.Type() == ERROR_OBJ
	}
	return false
}
```

大功告成！现在我们来看看怎么处理下面的两个错误：

1. 无效的标识符（`Identifier`）
2. 除数为零

首先处理`无效的标识符`错误：

```go
//eval.go
//解释标识符
func evalIdentifier(node *ast.Identifier, scope *Scope) Object {
	val, ok := scope.Get(node.Value)
	if !ok {
		return newError(node.Pos().Sline(), ERR_UNKNOWNIDENT, node.Value)
	}
	return val
}
```

如果从`Scope`中取不到变量，则返回一个错误对象。

接着，我们处理`除数为零`错误：

```go
//eval.go

//解释数字中缀表达式： 1 + 2.5, 3 * 3.2, ...
func evalNumberInfixExpression(node *ast.InfixExpression, left, right Object, scope *Scope) Object {
	leftVal := left.(*Number).Value
	rightVal := right.(*Number).Value

	switch node.Operator {
	//...
	case "/":
		if rightVal == 0 {
			return newError(node.Pos().Sline(), ERR_DIVIDEBYZERO)
		}
		return &Number{Value: leftVal / rightVal}
	//...
	}
}
```

如果除数为零，则返回一个`错误对象（Error Object）`。

上面的错误处理用到了两个常量`ERR_UNKNOWNIDENT`和`ERR_DIVIDEBYZERO`。我们希望代码中所有的错误信息，都能够在一个统一的地方进行定义。这样易于管理，也易于更改。

```go
//errors.go
var (
	ERR_UNKNOWNIDENT = "unknown identifier: '%s' is not defined"
	ERR_DIVIDEBYZERO = "divide by zero"
)
```

另外，我们之前处理`前缀表达式`和`中缀表达式`的时候，对于不支持的操作符（目前为止，我们只支持`+`和`-`），都返回的是`nil`。有了这节讲的错误处理，我们就可以返回这个`错误对象（Error Object）`了。首先我们需要再额外定义两个错误常量：

```go
//errors.go
	ERR_PREFIXOP     = "unsupported operator for prefix expression:'%s' and type: %s"
	ERR_INFIXOP      = "unsupported operator for infix expression: %s '%s' %s"
```

接着，我们在解释`前缀表达式`和`中缀表达式`的代码中加入其相应的错误处理：

```go
//eval.go
func evalPrefixExpression(node *ast.PrefixExpression, right Object, scope *Scope) Object {
	switch node.Operator {
	//...
	default: //对于不支持的类型，返回错误对象
		return newError(node.Pos().Sline(), ERR_PREFIXOP, node.Operator, right.Type())
	}
}

func evalPlusPrefixOperatorExpression(node *ast.PrefixExpression, right Object, scope *Scope) Object {
	if right.Type() != NUMBER_OBJ { //如果对象不是数字对象，则返回错误对象
		return newError(node.Pos().Sline(), ERR_PREFIXOP, node.Operator, right.Type())
	}
	return right
}

func evalMinusPrefixOperatorExpression(node *ast.PrefixExpression, right Object, scope *Scope) Object {
	if right.Type() != NUMBER_OBJ { //如果对象不是数字对象，则返回错误对象
		return newError(node.Pos().Sline(), ERR_PREFIXOP, node.Operator, right.Type())
	}
	//...
}

func evalInfixExpression(node *ast.InfixExpression, left, right Object, scope *Scope) Object {
	switch {
	//...
	default: //对于不支持的中缀操作符，返回错误对象
		return newError(node.Pos().Sline(), ERR_INFIXOP, left.Type(), node.Operator, right.Type())
	}
}

func evalNumberInfixExpression(node *ast.InfixExpression, left, right Object, scope *Scope) Object {
	leftVal := left.(*Number).Value
	rightVal := right.(*Number).Value

	switch node.Operator {
	//...
	default:
		return newError(node.Pos().Sline(), ERR_INFIXOP, left.Type(), node.Operator, right.Type())
	}
}
```

第6、12、19、28、39行都从原来的返回`nil`变成了返回`错误对象(Error Object)`。

我们还有最后一件事要做。由于我们的解释器（Evaluator）解释抽象语法树的时候，实际上是递归解释的。因此当任何一个地方出错的时候，我们需要即时返回，以防止错误四处传播，从而对错误源头的查找变得困难。

具体来看一下代码是怎么处理的：

```go
//eval.go
func Eval(node ast.Node, scope *Scope) (val Object) {

	switch node := node.(type) {
	case *ast.Program:
		return evalProgram(node, scope)
	case *ast.PrefixExpression:
		right := Eval(node.Right, scope)
		if isError(right) { //如果是错误对象，则返回
			return right
		}
		return evalPrefixExpression(node, right, scope)
	case *ast.InfixExpression:
		left := Eval(node.Left, scope)
		if isError(left) { //如果是错误对象，则返回
			return left
		}
		right := Eval(node.Right, scope)
		if isError(right) { //如果是错误对象，则返回
			return right
		}
		return evalInfixExpression(node, left, right, scope)

	case *ast.LetStatement:
		val := Eval(node.Value, scope)
		if isError(val) { //如果是错误对象，则返回
			return val
		}
		scope.Set(node.Name.Value, val)

	}

	return nil
}

func evalProgram(program *ast.Program, scope *Scope) (results Object) {
	for _, stmt := range program.Statements {
		results = Eval(stmt, scope)
		if errObj, ok := results.(*Error); ok { //如果是错误对象，则返回
			return errObj
		}
	}

	if results == nil {
		return NIL
	}
	return results
}
```

我们在`Eval()`函数的分支`*ast.PrefixExpression(前缀表达式)`、`*ast.InfixExpression(中缀表达式)`、`*ast.LetStatement(let语句)`中都加入了相关的错误判断，如果`Eval()`的返回值是一个错误对象，那么就即时返回。同时我们在`evalProgram`函数中也加入了相关的判断（39-40行）。

现在来测试一下我们的解释器，看看效果：

```go
//main.go
func TestEval() {
	tests := []struct {
		input    string
		expected string
	}{
        //正确
		{"let x = 2 + (3 * 4) / ( 6 - 3 ) + 10; x", "16"},
        
         //由于y不存在，所以会报告运行期错误，这里的期待值是随意给的，不用在意这个
		{"let x = 2 + (3 * 4) / ( 6 - 3 ) + 10; y", "error"},

        //由于除数为0，所以会报告运行期错误，这里的期待值是随意给的，不用在意这个
		{"2 + (3 * 4) / 0", "error"},
        
	}

	for _, tt := range tests {
		l := lexer.NewLexer(tt.input)
		p := parser.NewParser(l)
		program := p.ParseProgram()

		scope := eval.NewScope(nil, os.Stdout)
		evaluated := eval.Eval(program, scope)
		if evaluated != nil {
			if evaluated.Inspect() != tt.expected {
				fmt.Printf("%s\n", evaluated.Inspect())
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

运行结果如下：

```
let x = 2 + (3 * 4) / ( 6 - 3 ) + 10; x = 16
Runtime Error at 1
        unknown identifier: 'y' is not defined

Runtime Error at 1
        divide by zero
```



下一节，我们会增加对`return`语句（statement）支持。






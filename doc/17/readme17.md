# `方法调用（Method Call）`支持

在这一节中，我们要加入一个令人兴奋地内容：`方法调用（Method Call）`。

我们看一下方法调用的形式：

```go
obj.var1 //变量获取
obj.var2 = 10 //变量赋值
obj.call(param1, param2, ...) //方法调用
```

> 本节中我们对第一、第二种形式不讨论，只讨论第三种方法调用的形式。后续章节会讲解。对于第一、第二种形式，只需要读者记住，所谓的方法调用，并不完全只是方法调用，还包括对象的变量获取和变量赋值。

从上面的例子中，可以看到，我们增加了一个新的`.`中缀操作符。这个中缀操作符的左边是一个表达式`obj`，右边是另一个`函数调用（Call）`表达式:

```go
<left-expression> . <call-expression>
```

通过上面的分析，来看一下我们需要对代码做的更改：

1. 在词元（Token）的源码`token.go`中加入新的词元（Token）类型
1. 在词法分析器（Lexer）的源码`lexer.go`中加入对新的词元的分析。
2. 在抽象语法树（AST）的源码`ast.go`中加入`方法调用（Method Call）`对应的抽象语法表示。
3. 在语法解析器（Parser）的源码`parser.go`中加入对`方法调用（Method Call）`的语法解析。
4. 在对象（Object）源码`object.go`中给`Object`接口加入一个新的方法。
5. 在解释器（Evaluator）的源码`eval.go`中加入对`方法调用（Method Call）`的解释。



## 词元（Token）更改

```go
//token.go
const (
    //...
	TOKEN_DOT       //.
)

//词元类型的字符串表示
func (tt TokenType) String() string {
	switch tt {
	//...
	case TOKEN_DOT:
		return "."
	//...
	}
}
```



## 词法分析器（Lexer）的更改

```go
func (l *Lexer) NextToken() token.Token {
	//...
	switch l.ch {
	//...
	case '.':
		tok = newToken(token.TOKEN_DOT, l.ch)
	//...
	}

	//...
}
```



## 抽象语法树（AST）的更改

方法调用`.`这个中缀操作符，它的左边是一个`表达式(expression)`，右边是另一个`函数调用（Call）`表达式（上面说过了，还可能是对象的变量赋值或者获取，从现在开始，我们只陈述为函数调用）:

```go
<left-expression> . <call-expression>
```

从上面的形式中，我们能够很容易的得出方法调用的抽象语法表示：

```go
//ast.go
//方法调用的抽象语法表示： obj.call(param1, param2, ...)
type MethodCallExpression struct {
	Token  token.Token
	Object Expression //左边的表达式（即'.'之前的'对象'）
	Call   Expression //右边的方法调用（即'.'之后的'方法调用'）
}

func (mc *MethodCallExpression) Pos() token.Position {
	return mc.Token.Pos
}

func (mc *MethodCallExpression) End() token.Position {
	return mc.Call.End()
}

func (mc *MethodCallExpression) expressionNode()      {}
func (mc *MethodCallExpression) TokenLiteral() string { return mc.Token.Literal }
func (mc *MethodCallExpression) String() string {
	var out bytes.Buffer
	out.WriteString(mc.Object.String())
	out.WriteString(".")
	out.WriteString(mc.Call.String())

	return out.String()
}
```



## 语法解析器（Parser）的更改

我们需要做三处更改：

1. 对新增加的词元类型（Token type）注册中缀表达式回调函数
2. 新增解析`方法调用`表达式的函数
2. 对中缀操作符`.`赋予优先级

```go
//parser.go
var precedences = map[token.TokenType]int{
	//...
	token.TOKEN_LPAREN:   CALL,
	token.TOKEN_DOT:      CALL,
}

func (p *Parser) registerAction() {
    //...
    p.registerInfix(token.TOKEN_DOT, p.parseMethodCallExpression)
	//...
}

//解析方法调用
func (p *Parser) parseMethodCallExpression(obj ast.Expression) ast.Expression {
	methodCall := &ast.MethodCallExpression{Token: p.curToken, Object: obj}
	p.nextToken()

	name := p.parseIdentifier()
    if !p.peekTokenIs(token.TOKEN_LPAREN) { //如果下一个词元不是'('， 例如obj.xxx = 10
		methodCall.Call = p.parseExpression(CALL)
    } else { //obj.call(param1, param2, ...)
		p.nextToken()
		methodCall.Call = p.parseCallExpression(name)
	}

	return methodCall
}
```

代码第5行，我们给`方法调用obj.xxx()`赋予了和`函数调用xxx()`相同的优先级。

第8行我们给词元类型`TOKEN_DOT`注册了中缀表达式回调函数`parseMethodCallExpression`。

第20行判断`.`后面的词元是不是`TOKEN_LPAREN（左括号）`，如果不是，则调用`parseExpression`方法（21行），这里需要注意的是我们传入的优先级是`CALL`，而不是`LOWEST`。原因请参加下面的例子：

```perl
logger.LDATE + 1
```

对于这样的代码，如果我们传入的优先级是`LOWEST`，则因为`+`的优先级比`LOWEST`高，所以表达式变成了下面这样：

```perl
 logger.(LDATE + 1)
```

这并不是我们希望的。虽然在语法解析阶段做了处理，但是对于这种形式，我们在`解释阶段（Evaluating Phase）`暂不做处理，后续章节会处理。

如果`.`后面的词元是`TOKEN_LPAREN（左括号）`， 则调用既有的`parseCallExpression()`函数实现解析。

## 对象（Object）的更改

因为我们增加了`方法调用`支持，所以用户可以书写如下的代码：

```go
obj.call(param1, param2, ...)
```

既然我们的对象（Object）系统中，所有的内置类型都是对象（比如数字对象，字符串对象，布尔型对象等等），那么我们也希望这些内置对象也能够有这种形式的调用，比如对于一个`数字对象(Number Object)`，我们可以给它提供一个`str`方法，将这个数字转换为字符串。看一个例子：

```javascript
a = 10
a.str() //将a转换成字符串"10"
10.str() //直接将10装换为字符串"10"
```

简单来说，就是我们希望对象系统中的所有内置对象（比如数字对象，字符串对象，布尔型对象等等）都支持这种`方法调用`。而所有的内置对象都必须实现`Object`接口，所以我们需要给`Object`接口新提供一个方法。那么这个新的方法需要哪些参数呢？

1. 行号（用来调试或者报告错误）
2. 方法调用时的作用域`Scope`
3. 方法名
4. 方法的参数

方法调用的返回值又是什么呢？很简单，和`Eval`函数的返回值一样，就是一个`Object`。比如上面的例子`a.str()`返回的是一个字符串对象。

> 记住：在解释阶段（Evaluating Phase），我们内部处理的都是`Object`，也就是说我们是和`Object`在打交道。

下面我们来看代码：

```go
//object.go
type Object interface {
	Type() ObjectType
	Inspect() string
	CallMethod(line string, scope *Scope, method string, args ...Object) Object
}
```

其中第5行是新增的方法。因为`Object`接口有了这个新增方法，所以我们之前的所有内置对象都必须进行修改，以实现这个方法。下面我们来列举其中的几个实现：

```go
//object.go
//数字对象
func (n *Number) CallMethod(line string, scope *Scope, method string, args ...Object) Object {
	switch method {
	case "str":
		return n.str(line, args...)
	}
	return newError(line, ERR_NOMETHOD, method, n.Type())
}

func (n *Number) str(line string, args ...Object) Object {
	argLen := len(args)
	if argLen != 0 {
		return newError(line, ERR_ARGUMENT, "0", argLen)
	}

	return NewString(fmt.Sprintf("%g", n.Value)) //返回数字对象的字符串表示，即返回一个新的字符串对象
}
```

我们给`数字对象(Number Object)`提供了一个`str`方法，用来将数字转换为字符串。

```go
//object.go
//字符串对象
func (s *String) CallMethod(line string, scope *Scope, method string, args ...Object) Object {
	switch method {
	case "lower":
		return s.lower(line, args...)
	case "upper":
		return s.upper(line, args...)
	}
	return newError(line, ERR_NOMETHOD, method, s.Type())
}

func (s *String) lower(line string, args ...Object) Object {
	if len(args) != 0 {
		return newError(line, ERR_ARGUMENT, "0", len(args))
	}
	if s.String == "" {
		return s
	}

	str := strings.ToLower(s.String)
	return NewString(str)
}

func (s *String) upper(line string, args ...Object) Object {
	if len(args) != 0 {
		return newError(line, ERR_ARGUMENT, "0", len(args))
	}
	if s.String == "" {
		return s
	}

	ret := strings.ToUpper(s.String)
	return NewString(ret)
}
```

我们给`字符串对象(String Object)`提供了两个方法`lower()`和`upper()`方法，用来将字符串转换为大写或者小写。

代码内容比较简单，无非就是实现下面的几个功能：

1. 判断参数个数
2. 判断参数类型
2. 调用`go`语言的相应方法
2. 返回期待的对象（Object）

对于一部分内置对象，我们可能不需要提供任何方法，那么怎么办呢？非常简单，我们只需返回一个`Error对象`。例如对于`Nil对象`和`返回值对象(Return Object)`：

```go
//object.go
//Nil对象
func (n *Nil) CallMethod(line string, scope *Scope, method string, args ...Object) Object {
	return newError(line, ERR_NOMETHOD, method, n.Type())
}

//返回值对象
func (rv *ReturnValue) CallMethod(line string, scope *Scope, method string, args ...Object) Object {
	return newError(line, ERR_NOMETHOD, method, rv.Type())
}
```

上面的代码中，我们用到了两个常量，它们定义在`errors.go`文件中：

```go
//errors.go
var (
	//...
	ERR_ARGUMENT     = "wrong number of arguments. expected=%d, got=%d"
	ERR_NOMETHOD     = "undefined method '%s' for object %s"
）
```



## 解释器（Evaluator）的更改

我们需要在解释器（Evaluator）的`Eval`函数的`switch`分支中加入对`方法调用`的处理：

```go
//eval.go

func Eval(node ast.Node) (val Object) {
	switch node := node.(type) {
	//...

	case *ast.MethodCallExpression: //方法调用
		return evalMethodCallExpression(node, scope)

	}
	//...
	return nil
}

//解释”方法调用"：
// 1. obj.xxx ：变量赋值或者变量获取 （对于这种形式，本节的代码暂不做处理）
// 2. obj.call(param1, param2, ...)
func evalMethodCallExpression(methCall *ast.MethodCallExpression, scope *Scope) Object {
	obj := Eval(methCall.Object, scope)
	if obj.Type() == ERROR_OBJ {
		return obj
	}

	if method, ok := methCall.Call.(*ast.CallExpression); ok { //如果是个函数调用
		args := evalExpressions(method.Arguments, scope) //处理函数的参数
		if len(args) == 1 && isError(args[0]) {
			return args[0]
		}

		//调用对象的"CallMethod"方法
		return obj.CallMethod(methCall.Call.Pos().Sline(), scope, method.Function.String(), args...)
	}

	return newError(methCall.Call.Pos().Sline(), ERR_NOMETHOD, call.String(), obj.Type())
}
```

这里主要讲一下`evalMethodCallExpression`方法，第19行，我们首先解释（Evaluating）方法调用`obj.call(param1, param2, ...)`中的`obj`。然后，我们判断方法调用`obj.call(param1, param2, ...)`中的`call`是否为一个函数（第24行），如果是的话，则解释函数参数，接着调用对象的`CallMethod`方法（第31行）。

> 注意：这里我们并没有对变量赋值或者变量获取的形式：`obj.xxx`进行判断，因为现在还没有讲到。后续文章会扩展这里的处理。



## 测试

```go
//main.go
func TestEval() {
	tests := []struct {
		input    string
		expected string
	}{
		{`let str = "Hello " + "World!"; str.upper()`, "HELLO WORLD!"},
		{"let i = 10.35; i.str()", "10.35"},
	}

	for _, tt := range tests {
		l := lexer.NewLexer(tt.input)
		p := parser.NewParser(l)
		program := p.ParseProgram()
		evaluated := eval.Eval(program)
		if evaluated != nil {
			if evaluated.Inspect() != tt.expected {
				fmt.Printf("%s", evaluated.Inspect())
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

一切似乎看起来都没啥问题，我们的测试也都正常通过了。这里需要注意的是第二个测试用例：

```javascript
let i = 10.35; i.str()
```

这个语句经过测试没有问题。但是如果我们将这个例子更改成如下的形式：

```javascript
10.35.str()
```

再次运行程序，这个就会导致程序异常。这是什么原因产生的呢？让我们再回忆一下词法分析器（Lexer）中读取数字的函数`readNumber`：

```go
//lexer.go
func (l *Lexer) readNumber() string {
	var ret []rune

	ch := l.ch
	ret = append(ret, ch)
	l.readNext()

	for isDigit(l.ch) || l.ch == '.' {
		ret = append(ret, l.ch)
		l.readNext()
	}

	return string(ret)
}
```

第9-12行的`for`循环判读当前字符为`数字`或者`.`就继续。那么上面的例子`10.35.str()`取出来的数字就变成了`10.35.`，最后多了一个点，问题就出在这里。我们只需要做很小的一点更改就可以解决这个问题：

```go
//lexer.go
func (l *Lexer) readNumber() string {
	var ret []rune

	ch := l.ch
	ret = append(ret, ch)
	l.readNext()

	for isDigit(l.ch) || l.ch == '.' {
		if l.ch == '.' {
            if !isDigit(l.peek()) { //例如： 10.35.str()
				return string(ret) //提前返回
			}
		} //end if
		ret = append(ret, l.ch)
		l.readNext()
	}

	return string(ret)
}
```

10-14行的`if`判断是新增的代码。如果遇到`.`，就继续判断下一个字符。如果下一个字符不是数字的话就返回，这样读出来的数字就是`10.35`了。

修正了这个问题后，编译并再次运行测试，正常通过！！！



下一节，我们讲一个轻松一点的话题。我们将加入对`后缀++`和`后缀--`的支持。

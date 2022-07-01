# 函数支持

这一节，我们将实现对函数的支持。先来看一下，我们将要支持的函数的形式：

```javascript
let funcName = fn (param1, param2) { block } //声明函数字面量
fncName(argument1, argument2)  //函数调用
```

从上面的`函数字面量(Function Literal)`的形式中，读者也许猜到了，我们需要加入一个新的`fn`关键字，还有一个函数参数之间的区分符`,`。

还是老一套，先来看看我们需要做什么样的更改：

1. 在词元（Token）源码`token.go`中加入两个新的词元（Token）类型
2. 在词法分析器（Lexer）源码`lexer.go`中加入对新的词元的识别
3. 在抽象语法树（AST）的源码`ast.go`中加入`函数字面量`及`函数调用(Function Calling)`对应的抽象语法表示。
4. 在语法解析器（Parser）的源码`parser.go`中加入对`函数字面量`和`函数调用`的语法解析。
4. 在对象（Object）系统中的源码`object.go`中加入一个新的`函数对象(Function Object)`。
5. 在解释器（Evaluator）的源码`eval.go`中加入对`函数字面量`的解释及`函数调用`的解释。

## 词元（Token）更改

### 第一处改动

```go
//token.go
const (
	//...
	TOKEN_COMMA     // ','
	TOKEN_FUNCTION  // fn
```



### 第二处改动

```go
//token.go
//词元类型的字符串表示
func (tt TokenType) String() string {
	switch tt {
	//...
	case TOKEN_COMMA:
		return ","
	case TOKEN_FUNCITON:
		return "FUNCTION"
	}
}
```



### 第三处更改

```go
//token.go
var keywords = map[string]TokenType{
	//...
	"return": TOKEN_RETURN,
	"fn":     TOKEN_FUNCTION,
}
```

第5行，我们在`关键字（keywords）`变量中加入了`fn`关键字。



## 词法分析器（Lexer）的更改

我们需要在词法分析器（Lexer）的`NextToken()`函数中加入对`逗号(,)`的识别：

```go
//lexer.go

//获取下一个词元（Token）
func (l *Lexer) NextToken() token.Token {
    //...
	switch l.ch {
	//...
	case ',':
		tok = newToken(token.TOKEN_COMMA, l.ch)
	//...
	}
}

```



## 抽象语法树（AST）的更改

### `函数字面量`的抽象语法表示

我们再来温习一下文章最开始的`函数字面量`的形式：

```javascript
let funcName = fn (param1, param2) { block } //声明函数字面量
```

从上面的`函数字面量(Function Literal)`的形式中，我们来看看，`函数字面量(Function Literal)`的抽象语法表示，需要什么样的信息：

1. 词元信息（Token）   （same old friends）
2. 形参列表（Parameters）
3. 函数体（Body）

下面是`函数字面量(Function Literal)`的抽象语法表示：

```go
//ast.go

//函数字面量的抽象语法表示： fn (param1, param2, ...) { block }
type FunctionLiteral struct {
	Token      token.Token     // 'fn' token
	Parameters []*Identifier   //参数数组（是一个标识符数组）
	Body       *BlockStatement //函数体（body）
}

func (fl *FunctionLiteral) Pos() token.Position {
	return fl.Token.Pos
}

func (fl *FunctionLiteral) End() token.Position {
	return fl.Body.End()
}

//函数字面量是一个表达式
func (fl *FunctionLiteral) expressionNode()      {}

func (fl *FunctionLiteral) TokenLiteral() string { return fl.Token.Literal }

//`函数字面量`的字符串表示
func (fl *FunctionLiteral) String() string {
	var out bytes.Buffer

	params := []string{}
	for _, p := range fl.Parameters {
		params = append(params, p.String())
	}

	out.WriteString(fl.TokenLiteral())
	out.WriteString("(")
	out.WriteString(strings.Join(params, ", "))
	out.WriteString(") {")
	out.WriteString(fl.Body.String())
	out.WriteString("}")

	return out.String()
}
```

上面的代码应该不难理解，这里就不过多做解释了。

下面再让我们来看看`函数调用`的抽象语法表示。

### `函数调用`的抽象语法表示

同样，我们也来温习一下`函数调用`的形式：

```go
funcName(argument1, argument2, ...)
```

读者从上面的`函数调用`形式中，也许已经猜到了一二：

1. 词元信息（Token）
2. 函数名（Function）
3. 实参列表（arguments）

让我们来看一下`函数调用`的抽象语法表示：

```go
//ast.go
type CallExpression struct {
	Token     token.Token  // The '(' token
	Function  *Identifier  // 函数名
	Arguments []Expression //参数列表
}
```

有什么问题吗？大部分的读者可能就好奇了？难道有错吗？没有错误，但是关于`函数名`这个信息，我需要多说一点。下面让我们来举第一个例子：

```javascript
let add = fn (x,y) { return x + y } //声明一个简单的函数字面量
add(2,3) //调用此函数，结果是5
```

这里`add`就是函数名，它是一个`标识符(Identifier)`，没有错误，对吧。

咱们再来看另外一个例子：

```javascript
let add_result = fn(x,y) { return x + y } (2,3) //add_result的结果是5
```

在这个例子中，`fn(x,y) { return x + y }`是一个字面量，我们直接对字面量进行了函数调用，传递的参数是`2`和`3`。这里的`函数名`是一个`函数字面量(Function Literal)`。看到这里细心的读者可能已经明白了。我们的函数名还可以是个`函数字面量`。

所以我们需要对函数调用的`Function`参数做一下更改，最终的`函数调用`的抽象语法表示如下：

```go
//ast.go
//函数调用的抽象语法表示：<expression>(<comma separated expressions>)
type CallExpression struct {
	Token     token.Token // The '(' token
	Function  Expression  // Identifier or FunctionLiteral（标识符或者一个函数字面量）
	Arguments []Expression //实参数组
}

//对于'函数调用',开始位置是`(`左边的函数名。也就是`(`的位置减去函数名的长度
func (ce *CallExpression) Pos() token.Position {
	length := utf8.RuneCountInString(ce.Function.String())
	return token.Position{Filename: ce.Token.Pos.Filename, Line: ce.Token.Pos.Line, 
                          Col: ce.Token.Pos.Col - length}
}

//终了位置(这里的终了位置实际上不是很准确，我们没有考虑右括号的位置)
func (ce *CallExpression) End() token.Position {
	aLen := len(ce.Arguments)
	if aLen > 0 {
		return ce.Arguments[aLen-1].End()
	}
	return ce.Function.End()
}

//'函数调用'是一个表达式
func (ce *CallExpression) expressionNode()      {}

func (ce *CallExpression) TokenLiteral() string { return ce.Token.Literal }
//'函数调用'的字符串表示
func (ce *CallExpression) String() string {
	var out bytes.Buffer

	args := []string{}
	for _, a := range ce.Arguments {
		args = append(args, a.String())
	}

	out.WriteString(ce.Function.String())
	out.WriteString("(")
	out.WriteString(strings.Join(args, ", "))
	out.WriteString(")")
	return out.String()
}
```



## 语法解析器（Parser）的更改

我们需要做三处更改：

1. 对新增加的`TOKEN_FUNCTION`词元类型注册前缀表达式回调函数（第5行）
1. 对`函数调用`增加中缀表达式回调函数（第10行）
2. 增加`函数字面量`和`函数调用`的语法解析

```go
//parser.go
func (p *Parser) registerAction() {
	p.prefixParseFns = make(map[token.TokenType]prefixParseFn)
    //...
    p.registerPrefix(token.TOKEN_FUNCTION, p.parseFunctionLiteral)
	//...

	p.infixParseFns = make(map[token.TokenType]infixParseFn)
	//...
    p.registerInfix(token.TOKEN_LPAREN, p.parseCallExpression)
}
```

第三点【增加`函数字面量`和`函数调用`的语法解析】这个之后再说明。现在主要是想说明一下第二点。

`函数调用`为啥是中缀表达式？让我们再来看一下函数调用的形式：

```javascript
<function-namn> ( <arguments> )
```

再来看一下中缀表达式的形式：

```
<left-expression> operator <right-expression>
```

对比一下，我们可以得出如下的结论：

```go
<function-name> ( <arguments>
```

这里，我们将`(`作为中缀表达式的操作符（operator），中缀表达式的左表达式就是`<function-name>`，右表达式是函数实参`<arguments>`。这样一对比，是不是有些清楚了。

下面让我们来看看`parseFunctionLiteral`和`parseCallExpression`的实现。

### 解析函数字面量

先来看一下`parseFunctionLiteral`的实现：

```go
//parser.go
//解析函数字面量:
//  fn (param1, param2) { block-statement }
func (p *Parser) parseFunctionLiteral() ast.Expression {
	lit := &ast.FunctionLiteral{Token: p.curToken} //构造一个新的`函数字面量`表达式结构
    if !p.expectPeek(token.TOKEN_LPAREN) { //判断下一个词元是不是左括号'('
		return nil
	}
	lit.Parameters = p.parseFunctionParameters() //解析函数形参
    if !p.expectPeek(token.TOKEN_LBRACE) { //判断下一个词元是不是左花括号'{'
		return nil
	}
    lit.Body = p.parseBlockStatement() //解析块语句(block-statement)
	return lit
}

//解析函数参数（形参），返回参数列表
func (p *Parser) parseFunctionParameters() []*ast.Identifier {
	identifiers := []*ast.Identifier{} //创建一个标识符数组
    if p.peekTokenIs(token.TOKEN_RPAREN) { //如果下个词元是右括弧')'，说明函数没有参数
		p.nextToken()
		return identifiers
	}
	p.nextToken()
	ident := &ast.Identifier{Token: p.curToken, Value: p.curToken.Literal}
	identifiers = append(identifiers, ident)
	for p.peekTokenIs(token.TOKEN_COMMA) { //如果下一个词元类型是','就继续处理
		p.nextToken() //越过当前词元
		p.nextToken() //越过TOKEN_COMMA词元
		ident := &ast.Identifier{Token: p.curToken, Value: p.curToken.Literal}
		identifiers = append(identifiers, ident)
	}
    if !p.expectPeek(token.TOKEN_RPAREN) { //上面的for循环处理完成后，期待的下一个字符必须是右括号')'
		return nil
	}
	return identifiers
}
```

代码中的注释写的比较详细，读者理解起来应该不难。

### 解析函数调用

一句代码顶的上100句话，让我们直接上代码：

```go
//parser.go
//解析函数调用
func (p *Parser) parseCallExpression(function ast.Expression) ast.Expression {
	exp := &ast.CallExpression{Token: p.curToken, Function: function} //构造一个新的`函数调用`表达式结构
	exp.Arguments = p.parseCallArguments() //解析函数参数（实参）
	return exp
}

//解析函数参数（实参），返回实参数组
func (p *Parser) parseCallArguments() []ast.Expression {
	args := []ast.Expression{}
	if p.peekTokenIs(token.TOKEN_RPAREN) {
		p.nextToken()
		return args
	}
	p.nextToken()
	args = append(args, p.parseExpression(LOWEST))
	for p.peekTokenIs(token.TOKEN_COMMA) {
		p.nextToken()
		p.nextToken()
		args = append(args, p.parseExpression(LOWEST))
	}
	if !p.expectPeek(token.TOKEN_RPAREN) {
		return nil
	}
	return args
}
```

上面这个`parseCallArguments`和之前读者看到的`parseFunctionParameters`是不是非常像？只不过一个是返回的`标识符(Identifier)`数组， 一个是返回的`表达式(Expression)`数组。 请读者自行理解一下。

还有最后一点需要说明，之前的分析，我们将`函数调用（Function-Calling）`作为中缀表达式来解析，具体来说就是将`(`作为中缀表达式来解析。既然是中缀表达式，就会有优先级的问题。而且作为编程人员应该也知道，函数调用的优先级应该是最高的优先级（至少到目前为止应该是最高的优先级）。因此，我们需要给`TOKEN_LPAREN，即左括号'('`提供一个较高的优先级。

```go
//parser.go
const (
	_ int = iota
	LOWEST

	SUM       //+, -
	PRODUCT  // *, /
	PREFIX   //-X, +X
    CALL     //add()
)

var precedences = map[token.TokenType]int{

	token.TOKEN_PLUS:     SUM,
	token.TOKEN_MINUS:    SUM,
	token.TOKEN_MULTIPLY: PRODUCT,
	token.TOKEN_DIVIDE:   PRODUCT,
	token.TOKEN_POWER:    PRODUCT,
	token.TOKEN_LPAREN:   CALL,  //左括号这个词元类型对应着最高的优先级（目前位置）
}
```

第9行，我们增加了一个新的`CALL`常量，并在`precedences`这个map中增加了一行（第19行）。

为啥这个`CALL`的优先级比`-X`这种前缀表达式还高呢，一个简单的例子就能够说明：

```go
-sum(2,3)
```

我们当然是希望将`sum(2,3)`这个函数的返回结果取得后，再对其结果取负数，这个应该不难理解。



## 对象系统的更改

我们需要往对象系统中增加一个`函数对象(Function Object)`。那么`函数对象`需要包含什么样的信息呢？让我们来分析一下。

`函数对象`中当然需要知道函数的参数等信息，所以为了简便，我们直接将`函数字面量(FunctionLiteral)`这个抽象表达式做为`函数对象`的一个字段，这样当`函数对象`希望取得函数参数的时候，就可以直接从这个`函数字面量FunctionLiteral`中去取。

学编程的大家应该都知道，每个函数有自己的`作用域（Scope）`，函数内部声明的变量在函数终了的时候会从栈中销毁。因此这个`函数对象(Function Object)`还需要一个`作用域（Scope）`字段，用来保存函数内部使用的变量。

下面就是`函数对象(Function Object)`的表示：

```go
//object.go


const (
	//...
	FUNCTION_OBJ     = "FUNCTION"
)

//函数对象
type Function struct {
	Literal *ast.FunctionLiteral //函数字面量结构
	Scope   *Scope //作用域
}

func (f *Function) Type() ObjectType { return FUNCTION_OBJ }
func (f *Function) Inspect() string {
	return f.Literal.String()
}
```

内容并不复杂。



## 解释器（Evaluator）的更改

我们需要在解释器（Evaluator）的`Eval`函数的`switch`分支中加入对`函数字面量`表达式和`函数调用`表达式的处理。

```go
//eval.go
func Eval(node ast.Node, scope *Scope) (val Object) {
	switch node := node.(type) {
	//...
	case *ast.FunctionLiteral:
		return evalFunctionLiteral(node, scope) //解释函数字面量表达式
	case *ast.CallExpression:
		return evalCallExpression(node, scope)
		args := evalExpressions(node.Arguments, scope) //解释参数
		if len(args) == 1 && isError(args[0]) { //如果有错，则提前返回
			return args[0]
		}

        function := Eval(node.Function, scope) //解释函数，即'xxx(param1, param2,...)'中的'xxx'
		if isError(function) {
			return function
		}

		return applyFunction(node.Pos().Sline(), scope, function, args)
	//...
	}

	return nil
}

```

第5行和第7行的分支，分别是对`函数字面量`表达式和`函数调用`表达式的处理。下面分别做详细说明。

### `函数字面量`的解释

我们先来看看`函数字面量`的解释（Evaluating）。代码比较简单，因为`函数字面量`仅仅是函数声明，所以我们只是简单的返回一个`函数对象（Function Object）`：

```go
//eval.go
func evalFunctionLiteral(fl *ast.FunctionLiteral, scope *Scope) Object {
	return &Function{Literal: fl, Scope: scope}
}
```

### `函数调用`的解释

我们来看一下`Eval()`函数中对`函数调用`分支的处理代码：

```go
//eval.go
func evalCallExpression(node *ast.CallExpression, scope *Scope) Object {
	args := evalExpressions(node.Arguments, scope) // 解释函数参数
	if len(args) == 1 && isError(args[0]) { //如果只有一个对象，且这个对象是个Error对象，则返回
		return args[0]
	}

	//解析函数名，之前说过这个函数名可能是一个标识符，也可能是一个函数字面量:
	// 1. add(2,3)                       -> 函数名为标识符
	// 2. fn(x,y) { return x + y }(2,3)  -> 函数字面量
	function := Eval(node.Function, scope) //这个解释后的对象就是前面介绍的'函数对象'
	if isError(function) { //如果出现错误，则返回这个错误对象
		return function
	}

	return applyFunction(node.Pos().Sline(), scope, function, args)
}
```

我们来看一下`evalExpressions`这个解释`函数参数`的函数。我怎么觉得这句话念起来这么别扭？

```go
//eval.go
//解释表达式数组(exps)，返回解释后的对象数组'[]Object'
func evalExpressions(exps []ast.Expression, scope *Scope) []Object {
	var result []Object
    //对数组中的每一个表达式分别进行解释
	for _, e := range exps {
		evaluated := Eval(e, scope)
		if isError(evaluated) { //如果出现错误
			return []Object{evaluated} //返回一个数组，这个数组中只包含一个错误对象
		}

		result = append(result, evaluated)
	}

	return result
}
```

这个函数使用`for`循环来逐个解释`exps`数组中的表达式。解释后的对象（Object）放入结果数组。

下面再来看一下`applyFunction`这个函数的实现：

```go
//eval.go
//参数：
//  line: 行号（报错用）
//  scope:作用域
//  fn: 函数对象
//  args: 函数的实参数组
func applyFunction(line string, scope *Scope, fn Object, args []Object) Object {
	function, ok := fn.(*Function) //如果不是函数对象，则报错
	if !ok {
		return newError(function.Literal.Pos().Sline(), ERR_NOTFUNCTION, fn.Type())
	}
   
    //构造一个新的函数作用域
	extendedScope := extendFunctionScope(function, args)
  
    //解释函数体(body)，注意这里传入的作用域就是上面的扩展了的新的作用域。
	evaluated := Eval(function.Literal.Body, extendedScope)
	return unwrapReturnValue(evaluated)
}

//参数:
// fn: 函数对象
// args: 函数参数数组
func extendFunctionScope(fn *Function, args []Object) *Scope {
	scope := NewScope(fn.Scope, nil) //创建一个新的函数作用域
	for paramIdx, param := range fn.Literal.Parameters { //遍历函数参数
		scope.Set(param.Value, args[paramIdx]) //将其函数参数对应的对象放入函数作用域中。
	}

    //返回新的作用域
	return scope
}

func unwrapReturnValue(obj Object) Object {
     //如果对象是`返回对象(return object)`的话，取出'返回对象'中存储的返回值
    if returnValue, ok := obj.(*ReturnValue); ok {
		return returnValue.Value
	}

	return obj
}
```

上面就是`解释器（Evaluator）`的全部代码，内容有点多，请读者好好理解上面的代码。



细心的读者可能会有疑问：如果函数形参和实参个数不一样怎么办？其实在`applyFunction`代码中我们完全可以加入这个判断，为了简便起见，就没有加入。如果读者确实希望有这个检查，那么在上面`applyFunction`函数的第12行位置就可以加入这个判断，具体如下：

``` go
func applyFunction(line string, scope *Scope, fn Object, args []Object) Object {
	function, ok := fn.(*Function) //如果不是函数对象，则报错
	if !ok {
		return newError(function.Literal.Pos().Sline(), ERR_NOTFUNCTION, fn.Type())
	}
   
    //len(args):实参个数
    //len(function.Literal.Parameters):形参个数
    if len(args) != len(function.Literal.Parameters) {
        return newError(line, ERR_ARGUMENT, len(fn.Literal.Parameters), len(args))
    }
    
    //构造一个新的函数作用域
	extendedScope := extendFunctionScope(function, args)
}
```

上面代码的9-11行就是判断形参和实参个数是否相等的代码。

代码中`ERR_ARGUMENT`和`ERR_NOTFUNCTION`这两个常量是新增的，需要在`errors.go`中定义一下：

```go
//errors.go
var (
	ERR_ARGUMENT    = "wrong number of arguments. expected=%d, got=%d"
	ERR_NOTFUNCTION = "expect a function, got %s"
)
```



## 测试

下面我们写一个简单的程序测试一下：
```go
//main.go
func TestEval() {
	tests := []struct {
		input    string
		expected string
	}{
		{`let add = fn(x,y) {return x+y}; add(1,2)`, "3"},
		{`let add = fn(x,y) {return x+y}; let sub = fn(x,y) {x-y}; add(sub(5,3), sub(4,2))`, "4"},
        {`let sum = fn(x,y) { return x+y}(2,3); sum`, "5"}, //函数字面量方式调用
	}

	for idx, tt := range tests {
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
let add = fn(x,y) {return x+y}; add(1,2) = 3
let add = fn(x,y) {return x+y}; let sub = fn(x,y) {x-y}; add(sub(5,3), sub(4,2)) = 4
let sum = fn(x,y) { return x+y}(2,3); sum = 5
```

恭喜一下自己吧！！！:smile:



这一章的内容比较重要，请读者多读几遍，多体会一下，多写一些测试代码。就是要大家勤上手、多练习。



通过本节的学习，小喜鹊现在已经基本具备了飞行的能力。下一节，我们会增加对`内置函数`的支持。

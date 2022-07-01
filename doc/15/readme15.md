# 数组支持

在这一篇文章中，我们要加入对于`数组`的支持。先来看一下使用数组的例子：

```go
arr = [1, 10.5, "Hello", true];
println(arr[2])
```

从上面的例子中，我们可以得出如下信息：

数组是以`[]`括起来的部分，因此，我们词法解析器（Lexer）必须能够识别`[`和`]`。

下面看一下我们需要做的更改：

1. 在词元（Token）源码`token.go`中加入新的词元（Token）类型
2. 在词法分析器（Lexer）源码`lexer.go`中加入对`[`和`]`的识别
3. 在抽象语法树（AST）的源码`ast.go`中加入`数组`对应的抽象语法表示。
4. 在语法解析器（Parser）的源码`parser.go`中加入对`数组`的语法解析。
4. 在对象（Object）系统中的源码`object.go`中加入新的`数组对象(Array Object)`。
5. 在解释器（Evaluator）的源码`eval.go`中加入对`数组`的解释。



## 词元（Token）更改

无需太多解释，还是直接上代码：

```go
//token.go
const (
	//...
	TOKEN_LBRACKET  // [
	TOKEN_RBRACKET  // ]
)

//词元类型的字符串表示
func (tt TokenType) String() string {
	switch tt {
	//...
	case TOKEN_LBRACKET:
		return "["
	case TOKEN_RBRACKET:
		return "]"
	}
}
```



## 词法分析器（Lexer）的更改

我们需要在词法分析器（Lexer）的`NextToken()`函数中加入对`[`和`]`的识别：

```go
//lexer.go

//获取下一个词元（Token）
func (l *Lexer) NextToken() token.Token {
	//...

	switch l.ch {
	//...
	case '[':
		tok = newToken(token.TOKEN_LBRACKET, l.ch)
	case ']':
		tok = newToken(token.TOKEN_RBRACKET, l.ch)
	}
	//...
}

```



## 抽象语法树（AST）的更改

`数组`是由一系列的表达式组成的，类似下面的形式：

```go
arr = [<expression1>, <expression2>, ...]
```

从上面的形式我们可以得出`数组`的抽象语法表示：

```go
//ast.go
type ArrayLiteral struct { //数组字面量
	Token   token.Token
	Members []Expression //数组的元素，由一系列的表达式组成
}

func (a *ArrayLiteral) Pos() token.Position {
	return a.Token.Pos
}

func (a *ArrayLiteral) End() token.Position {
	aLen := len(a.Members)
	if aLen > 0 {
		return a.Members[aLen-1].End()
	}
	return a.Token.Pos
}

func (a *ArrayLiteral) expressionNode()      {}
func (a *ArrayLiteral) TokenLiteral() string { return a.Token.Literal }
func (a *ArrayLiteral) String() string {
	var out bytes.Buffer

	members := []string{}
	for _, m := range a.Members {
		members = append(members, m.String())
	}

	out.WriteString("[")
	out.WriteString(strings.Join(members, ", "))
	out.WriteString("]")
	return out.String()
}
```

这里的代码也是大家非常熟悉的。

上面看到的是数组的声明语法，再来看看我们`取数组元素`的形式：

```go
arr[<expression>]
```

很简单，从这个形式中，我们可以得出`取数组元素`的表达式（也称之为`索引表达式`）：

```go
//ast.go
//<Left-Expression>[<Index-Expression>]
type IndexExpression struct { // 索引表达式
	Token token.Token //'['对应的词元
	Left  Expression //左表达式
	Index Expression //索引表达式
}

func (ie *IndexExpression) Pos() token.Position {
	return ie.Token.Pos
}

func (ie *IndexExpression) End() token.Position {
	return ie.Index.End()
}

func (ie *IndexExpression) expressionNode()      {}
func (ie *IndexExpression) TokenLiteral() string { return ie.Token.Literal }
func (ie *IndexExpression) String() string {
	var out bytes.Buffer
	out.WriteString("(")
	out.WriteString(ie.Left.String())
	out.WriteString("[")
	out.WriteString(ie.Index.String())
	out.WriteString("]")
	out.WriteString(")")
	return out.String()
}
```

上面就是`数组字面量表达式(ArrayLiteral)`和`索引表达式(IndexExpression)`的抽象语法表示。



## 语法解析器（Parser）的更改

我们需要做两处更改：

1. 给新增的`TOKEN_LBRACKET`词元类型注册前缀表达式回调函数

   > 主要是处理`数组定义部分`： [1, 2, 3]

2. 给新增的`TOKEN_LBRACKET`词元类型注册中缀表达式回调函数

   > 主要是处理`取数组元素`: array1[1]

3. 既然`TOKEN_LBRACKET`是一个中缀操作符，因此我们还需要给其赋优先级

来看一下代码：

```go
//parser.go
func (p *Parser) registerAction() {
	p.prefixParseFns = make(map[token.TokenType]prefixParseFn)
	//...
	p.registerPrefix(token.TOKEN_LBRACKET, p.parseArrayLiteral) //注册前缀表达式


	p.infixParseFns = make(map[token.TokenType]infixParseFn)
	//...
	p.registerInfix(token.TOKEN_LBRACKET, p.parseIndexExpression) //注册中缀表达式
}

```

第5行和第10行我们分别给`TOKEN_LBRACKET`词元类型注册了`前缀表达式`和`中缀表达式`回调函数。读者可以回想一下第一篇文章【简单计算器】中关于`Pratt解析器`的说明：

> 对于每一个词元（Token）类型，我们可以有两个函数去处理它：`infix（中缀）`或者`prefix（前缀）`。选择哪个函数取决于Token在哪个位置。

例如：对于`[1,2,3]`这种，它就会使用前缀表达式来处理。而对于`arr[1]`这种，它就会使用中缀表达式来处理。对于为什么`[`可以作为中缀表达式来处理，我们在分析函数调用`xxx(a,b)`的时候其实已经有过类似的说明：

```javascript
//对于add(1)的函数调用，我们可以将`(`作为中缀操作符，而`add`标识符作为左表达式，`1`作为右表达式。形式如下：
"<expression> ( <expression>"
```

同样我们也可以将`[`作为中缀操作符：

```javascript
//对于`arr[1]` 的索引元素获取，我们可以将`[`作为中缀操作符，`arr`作为左表达式，`1`作为右表达式。形式如下：
"<expression> [ <expression>"
```



下面来看一下`parseArrayLiteral()`函数和`parseIndexExpression()`函数的实现。

先来看一下`parseArrayLiteral()`函数的实现：

```go
//parser.go
func (p *Parser) parseArrayLiteral() ast.Expression {
	array := &ast.ArrayLiteral{Token: p.curToken}
	array.Members = p.parseExpressionList(token.TOKEN_RBRACKET)
	return array
}

//获取表达式列表的共通函数:  xxx, xxx, xxx, ...
func (p *Parser) parseExpressionList(end token.TokenType) []ast.Expression {
	list := []ast.Expression{}
	if p.peekTokenIs(end) { //如果列表为空,比如：arr=[]
		p.nextToken()
		return list //返回空数组
	}

	p.nextToken()
	list = append(list, p.parseExpression(LOWEST))
	for p.peekTokenIs(token.TOKEN_COMMA) { //如果遇到逗号就继续
		p.nextToken()
		p.nextToken()
		list = append(list, p.parseExpression(LOWEST))
	}

	if !p.expectPeek(end) {
		return nil
	}

	return list
}
```

还记得`parseCallExpression()`函数的代码吗？我们来温习一下：

```go
//parser.go
func (p *Parser) parseCallExpression(function ast.Expression) ast.Expression {
	exp := &ast.CallExpression{Token: p.curToken, Function: function}
	exp.Arguments = p.parseCallArguments()
	return exp
}

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

当时我们处理函数调用的列表参数`param1, param2, ...`时候调用的是`parseCallArguments`函数。

而我们处理数组元素列表`element1, element2, ...`的逻辑和这个`parseCallArguments`内部处理是多么的相似（形式都是`xxx,xxx,xxx,...`）。只不过函数调用的终止符是`)`，数组定义的终止符是`]`。所以我们将其抽象了出来，写了一个共通的函数`parseExpressionList()`，并将`parseCallArguments()`删除掉。

因此我们的`parseCallExpression()`变成了下面这样：

```go
//parser.go
func (p *Parser) parseCallExpression(function ast.Expression) ast.Expression {
	exp := &ast.CallExpression{Token: p.curToken, Function: function}
    //调用'parseExpressionList'函数来处理函数参数，终止符是')'
	exp.Arguments = p.parseExpressionList(token.TOKEN_RPAREN)
	return exp
}
```

第4行，我们将`parseCallArguments()`函数更改成了`parseExpressionList()`函数。

接着，我们来看一下`parseIndexExpression()`函数的实现：

```go
//parser.go
func (p *Parser) parseIndexExpression(left ast.Expression) ast.Expression {
	exp := &ast.IndexExpression{Token: p.curToken, Left: left}
	p.nextToken()
	exp.Index = p.parseExpression(LOWEST) //获取索引表达式
	if !p.expectPeek(token.TOKEN_RBRACKET) { //如果下一个词元类型不是']'
		return nil
	}

	return exp
}
```

这个代码比较简单，应该无需太多说明。

最后，我们要追加的是`TOKEN_LBRACKET`（即`[`）的优先级：

```go
//parser.go
const (
	_ int = iota
	LOWEST

	EQUALS      //==, !=
	LESSGREATER //<, <=, >, >=
	SUM         //+, -
	PRODUCT     //*, /, **
	PREFIX      //!true, -10
	CALL        //add(1,2), array[index]
)

var precedences = map[token.TokenType]int{
	//...
	token.TOKEN_LPAREN:   CALL,
	token.TOKEN_LBRACKET: CALL,
}
```

从第17行我们可以知道，`索引访问`和`函数调用`的优先级是一样的。

> 这个是参照了python语言的优先级。



## 对象（Object）系统的更改

我们需要往对象系统中增加一个`数组对象(Array Object)`。数组里面存放的是什么呢？我想很多读者已经猜到了，就是任意多个对象（Object）：

```go
//object.go

const (
	//...
	ARRAY_OBJ       = "ARRAY"
)

//数组对象
type Array struct {
	Members []Object //数组的元素
}

func (ao *Array) Type() ObjectType { return ARRAY_OBJ }
func (ao *Array) Inspect() string {
	var out bytes.Buffer
	members := []string{}
	for _, e := range ao.Members {
		if e.Type() == STRING_OBJ {
			members = append(members, "\""+e.Inspect()+"\"")
		} else {
			members = append(members, e.Inspect())
		}
	}

	out.WriteString("[")
	out.WriteString(strings.Join(members, ", "))
	out.WriteString("]")
	return out.String()
}
```



## 解释器（Evaluator）的更改

我们需要在解释器（Evaluator）的`Eval`函数的`switch`分支中加入对`数组字面量表达式`和`索引表达式`的处理：

```go
//eval.go

func Eval(node ast.Node, scope *Scope) (val Object) {
	switch node := node.(type) {
	//...

	case *ast.ArrayLiteral: //处理数组字面量
		members := evalExpressions(node.Members, scope) //解释所有元素
		if len(members) == 1 && isError(members[0]) { //如果是有错误，则返回
			return members[0]
		}
		return &Array{Members: members} //返回数组对象


	case *ast.IndexExpression: //索引表达式：<left-expression> [ <index-expression>
		left := Eval(node.Left, scope) //处理左表达式<left-expression>
		if isError(left) {
			return left
		}

		index := Eval(node.Index, scope) //处理索引表达式<index-expression>
		if isError(index) {
			return index
		}

		return evalIndexExpression(node, left, index)
	}

	return nil
}

//处理索引表达式
func evalIndexExpression(node *ast.IndexExpression, left, index Object) Object {
	switch {
	case left.Type() == STRING_OBJ:
		return evalStringIndex(node.Pos().Sline(), left, index)
	case left.Type() == ARRAY_OBJ:
		return evalArrayIndexExpression(node.Pos().Sline(), left, index)
	default:
		return newError(node.Pos().Sline(), ERR_NOINDEXABLE, left.Type())
	}
}

//处理字符串的索引表达式：str[idx]
func evalStringIndex(lien string, left, index Object) Object {
	str := left.(*String)

	idx := int64(index.(*Number).Value)
	max := int64(utf8.RuneCountInString(str.String)) - 1
	if idx < 0 || idx > max {
		return newError(line, ERR_INDEX, idx)
	}

	//这里我们把字符串转换成`rune`数组，然后对其进行索引，之后再转换成字符串（不是很高效）
	return NewString(string([]rune(str.String)[idx]))
}

//处理数组的索引表达式：arr[idx]
func evalArrayIndexExpression(line string, array, index Object) Object {
	arrayObject := array.(*Array)

	idx := int64(index.(*Number).Value) //得到索引值
	max := int64(len(arrayObject.Members) - 1) //取得数组的元素数量
	if idx < 0 || idx > max { //范围检查
		return newError(line, ERR_INDEX, idx)
	}

	return arrayObject.Members[idx] //返回索引处的对象
}
```

代码的7-12行处理`数组字面量表达式`。15-26行处理`索引表达式`。58行`evalArrayIndexExpression`是实际处理数组索引的函数，它首先判断索引是否越界，越界的话就报错，否则就返回指定索引处的数组元素。

因为我们之前介绍过字符串，所以这里介绍索引表达式的时候也一并考虑了字符串（45-55行）。

还有一个地方需要说明，我们的语言允许用户书写如下语句：

```javascript
arr = [1, "hello"]
if arr { //判断数组长度
    println("array length is larger than zero")
}
```

因此，我们需要在`IsTrue`函数中加入相关的判断：

```go
//eval.go
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
		}
		return true
	}
}
```

代码的16-19行是新增的判断。

最后一个，差点忘记了。还记得前几节介绍的`len内置函数`吧，有了数组，我们希望`len内置函数`也能够接受数组参数，并返回数组的元素个数：

```go
//builtin.go

func lenBuiltin() *Builtin {
	return &Builtin{
		Fn: func(line string, scope *Scope, args ...Object) Object {
			if len(args) != 1 {
				return newError(line, "wrong number of arguments. got %d, want=1", len(args))
			}

			switch arg := args[0].(type) {
			case *String:
				n := utf8.RuneCountInString(arg.String)
				return NewNumber(float64(n))
			case *Array:
				return NewNumber(float64(len(arg.Members)))
			default:
				return newError(line, "argument to `len` not supported, got %s", args[0].Type())
			}
		},
	}
}
```

14-15行是新增加的判断。

## 测试

下面我们写一个简单的程序测试一下：
```go
//main.go
func TestEval() {
	tests := []struct {
		input    string
		expected string
	}{
		{`let arr = [1, 10.5, "Hello", true]; arr[0]`, "1"},
		{`let arr = [1, 10.5, "Hello", true]; arr[1]`, "10.5"},
		{`let arr = [1, 10.5, "Hello", true]; arr[2]`, "Hello"},
		{`let arr = [1, 10.5, "Hello", true]; arr[3]`, "true"},
		{`let arr = [1, 10.5, "Hello", true]; len(arr)`, "4"},
		{`let arr = []; len(arr)`, "0"},
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

通过上面的操作，我们完成了对数组的支持。小喜鹊现在能够组队（排列成数组）玩飞行躲猫猫的游戏了。:smile:



下一节，我们会提供对`哈希(Hash)`的支持。

> 我们这里所说的`哈希（Hash）`，有的语言中叫`字典(Dictionary)`，有的语言中叫`map`。

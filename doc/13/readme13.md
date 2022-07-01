# if-else判断支持

我们的小喜鹊（magpie）现在能够正常的飞行了，但是还无法辨别方向：

```go
if 方向 == "东" {
    flyHome()
} else if 方向 == "南" {
    goDrink()
} else {
    goSleep()
}
```

学习完这一节，我们的小喜鹊就可以具备辨别东西南北的能力了。

在这一节中，我们要加入对于`if-else`的支持。在正式实现`if-else`前，我们需要实现比较操作符（==, >=, <, <=）的词法分析(Lexer），语法解析（Parser）及解释（Evaluator）。



## 比较操作符

和之前的文章一样，我们来看看对于`比较操作符`，我们需要做的更改：

1. 在词元（Token）源码`token.go`中加入新的词元（Token）类型
2. 在词法分析器（Lexer）源码`lexer.go`中加入对`比较操作符`的识别
3. 在语法解析器（Parser）的源码`parser.go`中加入对`比较操作符`的语法解析。
5. 在解释器（Evaluator）的源码`eval.go`中加入对`比较操作符`的解释。

### 词元（Token）的更改

#### 第一处改动

```go
//token.go
const (
	//...
	TOKEN_LT  // <
	TOKEN_LE  // <=
	TOKEN_GT  // >
	TOKEN_GE  // >=
	TOKEN_EQ  // ==
	TOKEN_NEQ // !=

}
```



#### 第二处改动

```go
//token.go
//词元类型的字符串表示
func (tt TokenType) String() string {
	switch tt {
	//...
	case TOKEN_LT:
		return "<"
	case TOKEN_LE:
		return "<="
	case TOKEN_GT:
		return ">"
	case TOKEN_GE:
		return ">="
	case TOKEN_EQ:
		return "=="
	case TOKEN_NEQ:
		return "!="
	}
}
```



### 词法分析器（Lexer）的更改

我们需要在词法分析器（Lexer）的`NextToken()`函数中加入对`比较操作符`的识别：

```go
//lexer.go

//获取下一个词元（Token）
func (l *Lexer) NextToken() token.Token {
	var tok token.Token
	l.skipWhitespace()

	pos := l.getPos()

	switch l.ch {
	//...
	case '=':
		if l.peek() == '=' {
			tok = token.Token{Type: token.TOKEN_EQ, 
                              Literal: string(l.ch) + string(l.peek())}
			l.readNext()
		} else {
			tok = newToken(token.TOKEN_ASSIGN, l.ch)
		}
	case '>':
		if l.peek() == '=' {
			tok = token.Token{Type: token.TOKEN_GE, 
                              Literal: string(l.ch) + string(l.peek())}
			l.readNext()
		} else {
			tok = newToken(token.TOKEN_GT, l.ch)
		}
	case '<':
		if l.peek() == '=' {
			tok = token.Token{Type: token.TOKEN_LE, 
                              Literal: string(l.ch) + string(l.peek())}
			l.readNext()
		} else {
			tok = newToken(token.TOKEN_LT, l.ch)
		}
	case '!':
		if l.peek() == '=' {
			tok = token.Token{Type: token.TOKEN_NEQ, 
                              Literal: string(l.ch) + string(l.peek())}
			l.readNext()
		} else {
			//tok = newToken(token.TOKEN_BANG, l.ch)
		}
	}

}

```

代码虽然比较多，但都是之前学过的内容，而且理解起来应该不算困难。



### 语法解析器（Parser）的更改

我们需要做两处更改：

1. 对新增加的几个`比较操作符`词元类型注册中缀表达式回调函数
2. 既然`比较操作符`是中缀操作符，因此我们需要给其赋予优先级

我们先来看第一处的修改：

```go
//parser.go
func (p *Parser) registerAction() {
	//...

	p.infixParseFns = make(map[token.TokenType]infixParseFn)
	//...
    p.registerInfix(token.TOKEN_LT, p.parseInfixExpression)  // <
	p.registerInfix(token.TOKEN_LE, p.parseInfixExpression)  // <=
	p.registerInfix(token.TOKEN_GT, p.parseInfixExpression)  // >
	p.registerInfix(token.TOKEN_GE, p.parseInfixExpression)  // >=
	p.registerInfix(token.TOKEN_EQ, p.parseInfixExpression)  // ==
	p.registerInfix(token.TOKEN_NEQ, p.parseInfixExpression) // !=
}
```

第7-12行，我们给新增的几个比较操作符注册了中缀表达式回调函数。我们再来看第二处的修改（赋优先级）：

```go
//parser.go

const (
	_ int = iota
	LOWEST

	EQUALS       // ==, !=
	LESSGREATER  // >, >=, <, <=
	SUM
	//...
)

var precedences = map[token.TokenType]int{
	token.TOKEN_EQ:  EQUALS,
	token.TOKEN_NEQ: EQUALS,
	token.TOKEN_LT:  LESSGREATER,
	token.TOKEN_LE:  LESSGREATER,
	token.TOKEN_GT:  LESSGREATER,
	token.TOKEN_GE:  LESSGREATER,

	token.TOKEN_PLUS:     SUM,
	token.TOKEN_MINUS:    SUM,
	//...
}
```

我们给`==`和`!=`两个操作符赋予了最低的优先权（`EQUALS=2`），给`>`、`>=`、`<`、`<=`是个操作符赋予了稍微高一点的优先权（`LESSGREATER=3`）。这个和`go`语言及`c`语言的优先级是一样的。



### 解释器（Evaluator）的更改

既然我们增加的是中缀操作符，所以更改的地方就主要是`evalInfixExpression()`这个函数了：

```go
//eval.go
//解释中缀表达式
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
```

我们给`evalInfixExpression()`函数增加了两个`case`分支（第10-13行），用来解释`==`和`!=`操作符。

对于`>`, `>=`，`<`，`<=`这几个操作符来说，数字和字符串都适用。就是说数字之间可以比较，字符串之间也可以比较。下面分别是`字符串`的解释函数和`数字`的解释函数。

```go
//eval.go
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
		return newError(node.Pos().Sline(), ERR_INFIXOP, left.Type(), 
                        node.Operator, right.Type())
	}
}
```



```go
//eval.go
func evalNumberInfixExpression(node *ast.InfixExpression, left, right Object, scope *Scope) Object {
	leftVal := left.(*Number).Value
	rightVal := right.(*Number).Value

	switch node.Operator {
	//...
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
		return newError(node.Pos().Sline(), ERR_INFIXOP, left.Type(), 
                        node.Operator, right.Type())
	}
}
```

都是增加了几个简单的比较分支。



### 测试

对于比较操作符的所有更改都完成了，下面我们写一个简单的程序测试一下：
```go
//main.go
func TestEval() {
	tests := []struct {
		input    string
		expected string
	}{
		{`let x = 5 > 3; x`, "true"},
		{`let x = 3 >= 5; x`, "false"},
		{`let x = 5 < 3; x`, "false"},
		{`let x = 5 <= 3; x`, "false"}

        {`let x = 3 == 5; x`, "false"},
		{`let x = 3 != 5; x`, "true"},
        
		{`let x = "a" > "b"; x`, "false"},
		{`let x = "a" >= "b"; x`, "false"},
		{`let x = "Hello" >= "Hell"; x`, "true"},
        {`let x = "Hello" == "hello"; x`, "false"},
        {`let x = "Hello" == "Hello"; x`, "true"},
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

上面实现了对于`比较操作符`的解释（Evaluating），下面开始我们真正的主角登场了。



## `if-else`表达式

细心的读者可能会有疑问，`if-else`不是个语句（Statement）吗？为什么你这里说的是表达式（Expression），难道`if-else`还能返回值？没错，我们这里要实现的`if-else`确实是个表达式，能够返回值。给大家看个例子：

```javascript
let x = if 10 > 5 { 10 } else { 5 } //x的结果为10
```



还是老一套，为了能够实现`if-else`，让我们来看看我们需要做哪些更改：

1. 在词元（Token）源码`token.go`中加入两个新的词元（Token）类型（`TOKEN_IF`和`TOKEN_ELSE`）
2. 在抽象语法树（AST）的源码`ast.go`中加入`if-else`表达式对应的抽象语法表示。
3. 在语法解析器（Parser）的源码`parser.go`中加入对`if-else`表达式的语法解析。
4. 在解释器（Evaluator）的源码`eval.go`中加入对`if-esle`的解释。

### 词元的更改

因为改动比较简单，让我们直接来看下代码：

```go
//token.go
const (
	//...

    //reserved keywords
	//...
    TOKEN_IF       //if
	TOKEN_ELSE     //else
)

//词元类型的字符串表示
func (tt TokenType) String() string {
	switch tt {
	//...
	case TOKEN_IF:
		return "IF"
	case TOKEN_ELSE:
		return "ELSE"
	}
}

//关键字
var keywords = map[string]TokenType{
	//...
	"if":     TOKEN_IF,
	"else":   TOKEN_ELSE,
}
```



### 抽象语法树（AST）的更改

在开始分析之前，让我们看一下要实现的`if-else`表达式的形式：

```go
if <condition1> {
    <block-statement>
} else if <condition2> {
    <block-statement>
} else if <condition3> {
    <block-statement>
} else {
    <block-statement>
}
```

从上面的形式中，我们大致可以得到如下的信息：

1. 多个判断条件（`<condition>`）
2. 判断条件中的块代码（`<block-statement>`）
3. `else`（无判断条件）中的块代码（（`<block-statement>`））

仔细分析，可以看到`if`和`else if`分支实际上是一样的：

* 都有一个条件判断
* 都有对应的块语句

因此我们可以把`if`和`else-if`这两个抽象出一个结构：

```go
//ast.go

//if及其else-if表达式
type IfConditionExpr struct {
	Token token.Token
	Cond  Expression      //条件
    Body  *BlockStatement //块语句
}

func (ic *IfConditionExpr) Pos() token.Position {
	return ic.Token.Pos
}

func (ic *IfConditionExpr) End() token.Position {
	return ic.Body.End()
}

func (ic *IfConditionExpr) expressionNode()      {}
func (ic *IfConditionExpr) TokenLiteral() string { return ic.Token.Literal }

func (ic *IfConditionExpr) String() string {
	var out bytes.Buffer

	out.WriteString(ic.Cond.String())
	out.WriteString(" { ")
	out.WriteString(ic.Body.String())
	out.WriteString(" }")

	return out.String()
}
```

`IfConditionExpr`作为`if`和`else-if`的抽象语法表示就是这么简单。有了这个分析以后我们的`if-else`表达式的抽象语法表示就很明了了：

```go
//ast.go
type IfExpression struct {
	Token       token.Token //'if'词元
	Conditions  []*IfConditionExpr //'if'或者'else-if'部分，因为会有多个，所以这里是个数组
	Alternative *BlockStatement    //'else'部分的块语句
}

func (ifex *IfExpression) Pos() token.Position {
	return ifex.Token.Pos
}

func (ifex *IfExpression) End() token.Position {
	if ifex.Alternative != nil { //如果有else部分
		return ifex.Alternative.End()
	}

	aLen := len(ifex.Conditions)
	return ifex.Conditions[aLen-1].End()
}

func (ifex *IfExpression) expressionNode()      {}
func (ifex *IfExpression) TokenLiteral() string { return ifex.Token.Literal }

func (ifex *IfExpression) String() string {
	var out bytes.Buffer

	for i, c := range ifex.Conditions {
		if i == 0 {
			out.WriteString("if ")
		} else {
			out.WriteString("else if ")
		}
		out.WriteString(c.String())
	}

	if ifex.Alternative != nil {
		out.WriteString(" else ")
		out.WriteString(" { ")
		out.WriteString(ifex.Alternative.String())
		out.WriteString(" }")
	}

	return out.String()
}
```



### 语法解析器（Parser）的更改

对于语法解析器的更改，主要有两大部分：

1. 对`if`词元注册前缀表达式回调函数
2. 解析`if`表达式

先来看一下第一点的更改：

```go
//parser.go
func (p *Parser) registerAction() {
	p.prefixParseFns = make(map[token.TokenType]prefixParseFn)
	//...
	p.registerPrefix(token.TOKEN_IF, p.parseIfExpression) //对'if'注册前缀表达式回调函数
}
```



下面是对`if-else`表达式的解析代码：

```go
//parser.go
func (p *Parser) parseIfExpression() ast.Expression {
	ie := &ast.IfExpression{Token: p.curToken}
    //解析if/else-if/else的表达式
	ie.Conditions = p.parseConditionalExpressions(ie)
	return ie
}

func (p *Parser) parseConditionalExpressions(ie *ast.IfExpression) []*ast.IfConditionExpr {
	// if部分
	ic := []*ast.IfConditionExpr{p.parseConditionalExpression()}

	//else-if及else部分
	for p.peekTokenIs(token.TOKEN_ELSE) { //下一个词元类型是'else’就继续
		p.nextToken()

		if !p.peekTokenIs(token.TOKEN_IF) { //下一个词元类型是'if'吗
			if p.peekTokenIs(token.TOKEN_LBRACE) { //例如：'else {'
				p.nextToken()
				ie.Alternative = p.parseBlockStatement()
			} else {
				msg := fmt.Sprintf("Syntax Error:%v- 'else' part must be followed by a '{'.", 
                                   p.curToken.Pos)
				p.errors = append(p.errors, msg)
				p.errorLines = append(p.errorLines, p.curToken.Pos.Sline())
				return nil
			}
			break
		} else { //例如：'else if'
			p.nextToken()
			ic = append(ic, p.parseConditionalExpression())
		}
	}

	return ic
}

//解析if <conditon> { block }及else if <condition> { block }
func (p *Parser) parseConditionalExpression() *ast.IfConditionExpr {
	ic := &ast.IfConditionExpr{Token: p.curToken}
	p.nextToken()

    //解析<condition>
	ic.Cond = p.parseExpressionStatement().Expression

    if !p.peekTokenIs(token.TOKEN_LBRACE) { //如果下一个词元类型不是'{',则报错
		msg := fmt.Sprintf("Syntax Error:%v- 'if' expression must be followed by a '{'.", 
                           p.curToken.Pos)
		p.errors = append(p.errors, msg)
		p.errorLines = append(p.errorLines, p.curToken.Pos.Sline())
		return nil
	} else {
		p.nextToken()
		ic.Body = p.parseBlockStatement()
	}

	return ic
}
```

代码中的注释应该写的比较详细。如果读者觉得有些复杂，那么可以通过打印的方式来调试这段代码。我在`parser.go`文件中，写了一个简单的调试词元的函数`debugToken`函数，你可以很方便的调用：

```go
//parser.go
//DEBUG ONLY
func (p *Parser) debugToken(message string) {
	fmt.Printf("%s, curToken = %s, curToken.Pos = %d, peekToken = %s, peekToken.Pos=%d\n", 
               message, p.curToken.Literal, p.curToken.Pos.Line, p.peekToken.Literal,
               p.peekToken.Pos.Line)
}
```

调试的时候，可以像下面这样调用`debugToken`函数：

```go
p.debugToken("1111111111-1")
//some code
p.debugToken("1111111111-2")
```



### 解释器（Evaluator）的更改

我们需要在`Eval()`函数的`switch`语句的`case`分支中加入对`if-else`表达式的解释（5-6行）：

```go
//eval.go
func Eval(node ast.Node, scope *Scope) (val Object) {
	switch node := node.(type) {
	//...
	case *ast.IfExpression:
		return evalIfExpression(node, scope)
	}

	return nil
}
```



下面是实际的`if-else`表达式的解释代码：

```go
//eval.go
func evalIfExpression(ie *ast.IfExpression, scope *Scope) Object {
	//解释"if/else-if"部分
	for _, c := range ie.Conditions {
		condition := Eval(c.Cond, scope)
		if condition.Type() == ERROR_OBJ { //如果是错误对象，则提前返回
			return condition
		}

		if IsTrue(condition) { //如果条件满足
			return evalBlockStatement(c.Body, scope) //执行块语句
		}
	}

	//解释"else"部分
	if ie.Alternative != nil { //如果有`else`部分
		return evalBlockStatement(ie.Alternative, scope)
	}

	return NIL
}
```

这个解释代码比想象中的简单。先计算条件判断，如果判断满足，就执行相应的块语句（3-12行）。如果有`else`语句，则执行`else`中的块语句。

读者可能已经注意到了，对于条件是否为`真`的判断，我们使用了一个函数`IsTrue`函数。什么是`真`，什么是`假`，对于语言的开发者来说，需要做出选择。举个例子 ：

```javascript
let x = "xxx"
if x {
    println("x is not empty")
} else {
    println("x is empty")
}
```

对于上面一个简单的例子，有的语言会打印`x is not empty`。但是有的语言可能就会直接报错。对于第一种类型的语言，它会判断字符串的长度，如果字符串的长度大于0，则认为是`真`。而对于第二种类型的语言，则认为不能将字符串作为条件来判断。这就是语言开发者需要做出的选择。

下面来看一下我们的`IsTrue`实现：

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
		}
		return true
	}
}
```

下面是简单的描述：

* TRUE 返回`真`
* FALSE返回`假`
* NIL返回`假`
* 数字类型如果值为零，返回`假`
* 上述以外的情形，返回`真`

对于上面的简单描述中的【上述以外的情形，返回`真`】这种情况，为什么不是返回`假`，而是返回了`真`？

还是以上面的例子来说明情况：

```javascript
let x = "xxx"
if x {
    println("x is not empty")
} else {
    println("x is empty")
}
```

如果用户写了上面的代码，它的本意，可能是希望程序走那个`if`条件中的语句吧。

关于这个，争论的意义也不大。到底如何处理程序中所谓的`真`，所谓的`假`，是由语言作者来选择的。

正所谓`真亦假时假亦真,假亦真时真亦假`。:smile:



至此，本节对`if-else`表达式的支持就全部完成了。下面让我们来写一个测试程序。



### 测试

```go
//main.go
func TestEval() {
	tests := []struct {
		input    string
		expected string
	}{
		{"let x = 12; let result = if x > 10 {2} else if x > 5 {3} else {4}; result", "2"},
		{"let x = 10; let result = if x > 10 {2} else if x > 5 {3} else {4}; result", "3"},
		{"let x = 3; let result = if x > 10 {2} else if x > 5 {3} else {4}; result", "4"},
		{"let x = 8; let result = if x >= 8 {2} else if x > 5 {3} else {4}; result", "2"},
		{"let x = 8; let result = if x <= 8 {2} else if x > 5 {3} else {4}; result", "2"},
		{"let x = 8; let result = if x == 8 {2} else if x > 5 {3} else {4}; result", "2"},
		{"let x = 8; let result = if x != 8 {2} else if x > 5 {3} else {4}; result", "3"},
        {`let x = "hello"; let result = if len(x) == 5 { x }; result`, "hello"},
	}

	for _, tt := range tests {
		l := lexer.NewLexer(tt.input)
		p := parser.NewParser(l)
		program := p.ParseProgram()

		scope := eval.NewScope(nil, os.Stdout)
		evaluated := eval.Eval(program, scope)
		if evaluated != nil {
			if evaluated.Inspect() != tt.expected {
				fmt.Printf(%s\n", evaluated.Inspect())
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
let x = 12; let result = if x > 10 {2} else if x > 5 {3} else {4}; result = 2
let x = 10; let result = if x > 10 {2} else if x > 5 {3} else {4}; result = 3
let x = 3; let result = if x > 10 {2} else if x > 5 {3} else {4}; result = 4
let x = 8; let result = if x >= 8 {2} else if x > 5 {3} else {4}; result = 2
let x = 8; let result = if x <= 8 {2} else if x > 5 {3} else {4}; result = 2
let x = 8; let result = if x == 8 {2} else if x > 5 {3} else {4}; result = 2
let x = 8; let result = if x != 8 {2} else if x > 5 {3} else {4}; result = 3
let x = "hello"; let result = if len(x) == 5 { x }; result = hello
```

恭喜恭喜！！我们的小喜鹊现在能够完全辨别方向了。



这一节的内容比较长。下一节我们轻松点，仅仅增加`!（取反）`的操作。


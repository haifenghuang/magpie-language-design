# Let语句支持

在这一节中，我们要加入对`let`语句的支持。在我们的自制语言中，可以使用`let`语句声明变量。其形式如下：

```javascript
let <identifier> = <expression>;
let <identifier>; //值可以为空
```

从上面的`let`语句中，读者可能已经猜到了，我们需要增加三个词元（Token）类型：

```
TOKEN_LET        ----> let
TOKEN_ASSIGN     ----> =
TOKEN_SEMICOLON  ----> ;
```

> 实际上，分号（；）是可选的。

现在来看一下，我们需要做哪些更改：

1. 在词元（Token）源码`token.go`中加入新的词元（Token）类型
1. 在词法解析器（Lexer）源码中，增加对新的词元（Token）类型的解析
2. 在抽象语法树（AST）的源码`ast.go`中加入`let`语句对应的抽象语法表示
3. 在语法解析器（Parser）的源码`parser.go`中加入对`let`语句的语法解析。
4. 在解释器（Evaluator）的源码`eval.go`中加入对语句（statement）的解释。

## 词元（Token）更改

### 第一处改动

```go
//token.go
const (
    //...
    
	TOKEN_ASSIGN    // =
	TOKEN_SEMICOLON //;

	//reserved keywords
	//...
	TOKEN_LET    //let
)

```

我们加入了三个新的词元（Token）类型。

### 第二处改动

```go
//token.go
//词元类型的字符串表示
func (tt TokenType) String() string {
	switch tt {
	//...
	case TOKEN_ASSIGN:
		return "="
	case TOKEN_SEMICOLON:
		return ";"
	case TOKEN_LET:
		return "let"
	//...
	}
}
```

在词元类型（Token Type）的字符串表示中，加入了三个`case`分支。

### 第三处改动

```go
//token.go
//关键字map
var keywords = map[string]TokenType{
    //...
    "let":    TOKEN_LET,
}
```

第5行， 我们给`keywords`变量增加了`let`关键字。



## 词法解析器（Lexer）的更改

我们只需要在`NextToken()`函数中加入对新的词元类型（`TOKEN_ASSIGN`和`TOKEN_SEMICOLON`）的解析即可：

```go
//lexer.go
func (l *Lexer) NextToken() token.Token {
	var tok token.Token
	l.skipWhitespace()

	pos := l.getPos()

	switch l.ch {
	//...
	case '=':
		tok = newToken(token.TOKEN_ASSIGN, l.ch)
	case ';':
		tok = newToken(token.TOKEN_SEMICOLON, l.ch)
	//...
```

第10-13行，我们在`NextToken()`函数的`switch`分支中加入了对`=`和`;`的判断。

> 有的读者会有疑问，为啥没有对`let`关键字的解析呢？因为`let`关键字的解析已经包含在了`readIdentifier`的那个分支里面了。所以，以后如果仅仅只是加入对关键字（keyword）的支持，词法解析器（Lexer）不用更改。



## 抽象语法树（AST）的更改

由于我们这次增加了`let`语句（statement），所以现在我们的程序就不仅仅只包含前几节介绍的表达式（expression）了，这次还多了语句（statement，这里是let语句）。因此，我们的抽象语法树，需要加入对`语句（statement）`的支持。先来看一下代码中如何表示`语句（statement）`。

```go
//ast.go
type Expression interface { //表达式
	Node
	expressionNode()
}

type Statement interface { //语句
	Node
	statementNode()
}
```

7-10行我们加入了一个`Statement`的接口，和`Expression`这个接口一样，这个`Statement`接口也是`节点（Node）`。

我们程序中的所有语句（statement），都必须实现这个`Statement`接口。`let`语句当然也不例外。

请读者思考一下，我们的`let`语句需要什么信息呢？现在我把`let`语句的形式再给读者看一下：

```javascript
let <identifier> = <expression>;
let <identifier>; //值可以为空
```

1. 词元（Token）的信息，这是所有的节点（Node）都必须包含的信息（用来调试、报错等）
2. 变量名（左边的<identifier>）
3. 变量值（右边的<expression>）

```go
//ast.go
//let语句： 
//    let <identifier> = <expression>;
//    let <identifier>;
type LetStatement struct {
    Token token.Token   //词元(Token)信息
	Name  *Identifier  //变量名
	Value Expression   //变量值
}

//开始位置
func (ls *LetStatement) Pos() token.Position {
	return ls.Token.Pos
}

//结束位置
func (ls *LetStatement) End() token.Position {
    return ls.Value.End() //值（Value）的结束位置
}

//表明`let`是个语句（statement）
func (ls *LetStatement) statementNode()       {}

func (ls *LetStatement) TokenLiteral() string { return ls.Token.Literal }

//`let`语句的字符串表示(主要调试用)
func (ls *LetStatement) String() string {
	var out bytes.Buffer

	out.WriteString(ls.TokenLiteral() + " ")
	out.WriteString(ls.Name.String())
	out.WriteString(" = ")

	if ls.Value != nil {
		out.WriteString(ls.Value.String())
	}

	out.WriteString(";")

	return out.String()
}
```

请读者再想一想，我们还有什么遗漏的地方吗？ 前面我说过，我们的程序现在不仅支持表达式（Expression），现在还支持语句（Statement）了。所以，我们的程序（Program）节点，就必须反映这种变动。

先看一下我们的程序（Program）节点：

```go
//ast.go
type Program struct {
	Expression Expression
}
```

可以看到，之前的程序（Program）节点中，只有一个表达式（expression）节点。现在有了`let`语句（statement），我们的程序（Program）节点，就可以包含多个节点（Node）了。因此，更改后的程序（Program）节点，变成了下面这样：

```go
//ast.go
type Program struct {
	Statements []Statement
}
```

细心的读者可能会问了：不对啊？这样的话，程序（Program）节点就变成了只支持`语句(Statement)`了，而不支持表达式（Expression）了，是吗？

这个问题问的好，值得详细的说明一下。为了统一，更为了方便代码的编写，我们加入了`表达式语句（Expression-Statement）`。what？ `表达式语句（Expression-Statement）`？ 是不是更晕了？ 相信我，你没有看错，就是`表达式语句（Expression-Statement）`。 还是来看一下`表达式语句（Expression-Statement）`的代码吧：

```go
//ast.go
//表达式语句
type ExpressionStatement struct {
	Token      token.Token
    Expression Expression //表达式语句中只包含表达式(expression)节点
}
```

让我们来分析一下，如果没有这个`表达式语句（Expression-Statement）`，如何表示我们的程序（Program）节点？像下面这样表示吗：

```go
//ast.go
type Program struct {
	Statements []Statement
	Expressions []Expression
}
```

这样表示的问题是显而易见的。我们的程序节点必须区分处理表达式（Expression）和语句（Statement）。但是有了这个`表达式语句（Expression-Statement）`，我们就可以用统一的形式来解析程序（Program）节点了。有了这个说明，理解起来是不是更容易一些了？

我们来看一下`表达式语句（Expression-Statement）`的完整代码：

```go
//ast.go
//表达式语句
type ExpressionStatement struct {
	Token      token.Token
	Expression Expression //表达式
}

func (es *ExpressionStatement) Pos() token.Position {
	return es.Token.Pos
}

func (es *ExpressionStatement) End() token.Position {
	return es.Expression.End()
}

//`表达式语句`是一个语句(只包含表达式的语句）
func (es *ExpressionStatement) statementNode()       {}

func (es *ExpressionStatement) TokenLiteral() string { return es.Token.Literal }

func (es *ExpressionStatement) String() string {
	if es.Expression != nil {
		return es.Expression.String()
	}
	return ""
}
```

接下来看一下变动后的`程序（Program）节点`的完整代码：

```go
//ast.go
//程序节点
type Program struct {
	Statements []Statement //程序中包含多个语句（statement）
}

//开始位置
func (p *Program) Pos() token.Position {
	if len(p.Statements) > 0 {
		return p.Statements[0].Pos() //返回第一条语句的开始位置
	}
	return token.Position{} //程序中没有语句的时候（比如程序为空），返回一个空的位置。
}

//结束位置
func (p *Program) End() token.Position {
	aLen := len(p.Statements)
	if aLen > 0 {
		return p.Statements[aLen-1].End() //返回最后一条语句的结束位置
	}
	return token.Position{} //程序中没有语句的时候（比如程序为空），返回一个空的位置。
}

func (p *Program) TokenLiteral() string {
	if len(p.Statements) > 0 {
		return p.Statements[0].TokenLiteral() //返回第一个语句的TokenLiteral
	}
	return ""
}
//程序（Program）节点的字符串表示
func (p *Program) String() string {
	var out bytes.Buffer

    //循环输出语句（statement）的字符串表示
	for _, s := range p.Statements {
		out.WriteString(s.String())
	}

	return out.String()
}
```

请读者仔细理解我上面说的关于`表达式语句（Expression-Statement）`的含义。



## 语法解析器（Parser）的更改

我们需要做下面几处更改：

1. 对变更后的程序（Program）节点的重新解析。
2. 对新增的`let`语句（`LetStatement`）的解析。
2. 对新增的`ExpressionStatement`语句的解析。

### 程序（Program）节点的解析

```go
//parser.go

//程序节点的解析
func (p *Parser) ParseProgram() *ast.Program {
	program := &ast.Program{} //生成程序节点
   
     //循环解析语句
	for p.curToken.Type != token.TOKEN_EOF { //如果没有遇到结束词元类型，就继续处理
		stmt := p.parseStatement() //解析语句
		if stmt != nil {
			program.Statements = append(program.Statements, stmt)
		}
		p.nextToken()
	}

	return program
}

//解析语句
//将来我们增加对其它语句（比如：函数语句，return语句）的支持的时候，会扩展这个'switch'分支
func (p *Parser) parseStatement() ast.Statement {
	switch p.curToken.Type {
	case token.TOKEN_LET:
		return p.parseLetStatement() //解析'let'语句
	default:
		return p.parseExpressionStatement() //解析'表达式语句'
	}
}
```

### `let`语句的解析

```go
//parser.go
// let语句：
//    let <identifier> = expression;
//    let <identifier>;
func (p *Parser) parseLetStatement() *ast.LetStatement {
	stmt := &ast.LetStatement{Token: p.curToken} //生成'let'节点

	if p.expectPeek(token.TOKEN_IDENTIFIER) { //期望下一个词元类型为标识符<identifier>类型
		stmt.Name = &ast.Identifier{Token: p.curToken, Value: p.curToken.Literal}
	}

    if p.expectPeek(token.TOKEN_ASSIGN) { //期待下一个词元类型为'TOKEN_ASSIGN(=)'
		p.nextToken()
		stmt.Value = p.parseExpressionStatement().Expression //调用表达式语句获取`let`语句右侧表达式的值。
	}

	return stmt
}
```

### 表达式语句（Expression-Statement）的解析

```go
//parser.go
//表达式语句
func (p *Parser) parseExpressionStatement() *ast.ExpressionStatement {
	stmt := &ast.ExpressionStatement{Token: p.curToken} //生成'表达式语句'节点

	stmt.Expression = p.parseExpression(LOWEST) //解析表达式

    //如果下一个词元类型为TOKEN_SEMICOLON(;)，则忽略这个';'号。
    //就是说我们的语句可以有';'号，也可以没有';'号
    if p.peekTokenIs(token.TOKEN_SEMICOLON) {
		p.nextToken()
	}
	return stmt
}
```

从`parseExpressionStatement`函数中我们可以知道，我们的语句末尾可以有分号，也可以没有分号：

```javascript
let x = 10
let x = 10;
```

上面的两个语句是等价的。



## 解释器（Evaluator）的更改

由于我们变更了程序（Program）节点的表示，同时新增了表达式语句`Expression-Statement`。我们需要在解释器（Evaluator）的`Eval`函数的`switch`分支中加入相应的处理：

```go
//eval.go
func Eval(node ast.Node) (val Object) {

	switch node := node.(type) {
	case *ast.Program:
		return evalProgram(node)
	case *ast.ExpressionStatement: //'表达式语句'节点
		return Eval(node.Expression) //解释'表达式语句'中包含的表达式
	//...
	}

	return nil
}

//解释程序（Program）节点
func evalProgram(program *ast.Program) (results Object) {
    //循环解释'程序（Program）节点'中的语句
	for _, stmt := range program.Statements {
		results = Eval(stmt)
	}

	if results == nil { //如果结果为nil，比如程序中没有任何语句，则默认返回`NIL`对象
		return NIL 
	}
	return results
}
```

这样我们就完成了对解释器的更改。慢着，这里并没有增加对`let`语句节点的解释啊？不好意思，解释`let`语句节点，需要一些额外的知识。等相关的知识具备后，会补充这个实现。



下一节，我们将介绍`Scope（作用域）`的相关知识，也是非常重要的知识。为后续语言的扩展打好基础。

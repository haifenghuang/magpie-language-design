# true, false和nil支持

在这一节中，我们要加入对于`true`，`false`和`nil`的支持。

我们需要做如下的更改：

1. 在词元（Token）源码`token.go`中加入新的词元（Token）类型
2. 在抽象语法树（AST）的源码`ast.go`中加入`true`， `false`和`nil`对应的抽象语法表示。
3. 在语法解析器（Parser）的源码`parser.go`中加入对`true`， `false`和`nil`语法解析。
4. 在对象（Object）源码`object.go`中加入新的对象类型（布尔对象和Nil对象）
5. 在解释器（Evaluator）的源码`eval.go`中加入对`true`， `false`和`nil`的解释。

## 词元（Token）更改

### 第一处改动

```go
//token.go
const (
    //...

	TOKEN_NUMBER     //10 or 10.1
    TOKEN_IDENTIFIER //identifier: a, b, var1, ...
    
    //reserved keywords
	TOKEN_TRUE  //true
	TOKEN_FALSE //false
	TOKEN_NIL   // nil
)
```

第9-11行，我们加入了三个新的词元（Token）类型。

### 第二处改动

```go
//token.go
//词元类型的字符串表示
func (tt TokenType) String() string {
	switch tt {
    //...

	case TOKEN_IDENTIFIER:
		return "IDENTIFIER"
	case TOKEN_TRUE:
		return "TRUE"
	case TOKEN_FALSE:
		return "FALSE"
	case TOKEN_NIL:
		return "NIL"
	default:
		return "UNKNOWN"
	}
}
```

代码9-14行是新增的代码。

### 第三处改动

```go
//token.go

//关键字map
var keywords = map[string]TokenType{
    "true":  TOKEN_TRUE,
	"false": TOKEN_FALSE,
	"nil":   TOKEN_NIL,
}
```

给`keywords`变量增加了三个关键字。

## 抽象语法树（AST）的更改

对于脚本中出现的`true`，`false`，`nil` 我们也需要为其增加节点（Node）表示。

```go
//ast.go
//nil表达式： nil
type NilLiteral struct {
	Token token.Token
}

func (n *NilLiteral) Pos() token.Position {
	return n.Token.Pos
}

func (n *NilLiteral) End() token.Position {
	length := len(n.Token.Literal)
	pos := n.Token.Pos
	return token.Position{Filename: pos.Filename, Line: pos.Line, Col: pos.Col + length}
}

//nil是个表达式
func (n *NilLiteral) expressionNode()      {}
func (n *NilLiteral) TokenLiteral() string { return n.Token.Literal }
func (n *NilLiteral) String() string       { return n.Token.Literal }

//布尔表达式: true, false
type BooleanLiteral struct {
	Token token.Token
	Value bool //存放布尔值
}

func (b *BooleanLiteral) Pos() token.Position {
	return b.Token.Pos
}

//结束位置 = 【开始位置】+【"true/false"的长度】
func (b *BooleanLiteral) End() token.Position {
	length := utf8.RuneCountInString(b.Token.Literal)
	pos := b.Token.Pos
	return token.Position{Filename: pos.Filename, Line: pos.Line, Col: pos.Col + length}
}
//布尔类型也是一个表达式
func (b *BooleanLiteral) expressionNode()      {}
func (b *BooleanLiteral) TokenLiteral() string { return b.Token.Literal }
func (b *BooleanLiteral) String() string       { return b.Token.Literal }
```

这都是读者比较熟悉的代码了，所以也没有太多需要解释的。



## 语法解析器（Parser）的更改

我们需要做两处更改：

1. 对新增加的三个词元类型（Token type）注册前缀表达式回调函数（代码中的6-8行）
2. 解析两个表达式节点的函数（代码中的14-21行）

```go
//parser.go
func (p *Parser) registerAction() {
	//...

    //给三个新的词元类型注册前缀表达式回调函数
    p.registerPrefix(token.TOKEN_TRUE, p.parseBooleanLiteral)
	p.registerPrefix(token.TOKEN_FALSE, p.parseBooleanLiteral)
	p.registerPrefix(token.TOKEN_NIL, p.parseNilExpression)

	//...
}

//解析布尔表达式
func (p *Parser) parseBooleanLiteral() ast.Expression {
	return &ast.BooleanLiteral{Token: p.curToken, Value: p.curTokenIs(token.TOKEN_TRUE)}
}

//解析nil表达式
func (p *Parser) parseNilExpression() ast.Expression {
	return &ast.NilLiteral{Token: p.curToken}
}
```



## 对象（Object）的更改

由于我们新增了两个类型（布尔类型和Nil类型），我们就需要在对象（Object）系统中加入这两个新的类型的支持。

1. 新增两个对象类型（Object Type）
2. 实现两个新增对象（即实现`object`接口的方法）

### 新增两个对象类型（Object Type）

```go
//object.go
const (
	//...
	NIL_OBJ     = "NIL_OBJ"
	BOOLEAN_OBJ = "BOOLEAN"
)
```

第4-5行，我们新增加了两个对象类型（Object Type）。



### 实现两个新增对象（即实现`object`接口的方法）

```go
//object.go

//Nil对象（Nil ojbect）
type Nil struct {
}

func (n *Nil) Inspect() string {
	return "nil"
}
func (n *Nil) Type() ObjectType { return NIL_OBJ }

//布尔对象（Boolean object）
type Boolean struct {
	Bool bool //存放布尔值
}

func (b *Boolean) Inspect() string {
	return fmt.Sprintf("%v", b.Bool)
}

func (b *Boolean) Type() ObjectType { return BOOLEAN_OBJ }
```

内容也比较简单，就不多解释了。

现在让我们想一个问题，对于`Nil对象`和`布尔对象`，我们是否需要多个这种类型的对象？比如下面的脚本代码：

```go
a = nil
b = nil
if a == nil {}
if b == nil {}

c = true
d = true
if c == true { }
if d == true { }

e = false
f = false 
if e == false { }
if f == false { }
```

仔细分析上面的代码，看出什么问题吗？

第3行和第4行：我们判断变量`a`和变量`b`是否为nil值。我们是把变量`a`和一个`Nil对象`比较，而变量`b`和另外一个全新的`Nil对象`比较吗？

第8行和第8行：我们判断变量`c`和变量`d`是否为true值。我们是把变量`c`和一个`布尔对象`比较，而变量`d`和另外一个全新的`布尔对象`比较吗？

第13行和第14行：我们判断变量`e`和变量`f`是否为false值。我们是把变量`e`和一个`布尔对象`比较，而变量`f`和另外一个全新的`布尔对象`比较吗？

相信通过上面的分析，细心的读者在脑子里已经有了答案。就是我们的对象（object）系统中，实际上只需要有一个布尔真、一个布尔假、一个Nil对象就足够了。长话短说就是：

1. true和true没有区别，它们是相同的
2. false和false没有区别，它们是相同的
3. nil和nil没有区别，它们是相同的

为了实现上面的逻辑，我们在`object.go`中新增几行代码：

```go
//object.go
var (
	TRUE  = &Boolean{Bool: true}
	FALSE = &Boolean{Bool: false}
	NIL   = &Nil{}
)
```

就是说，在我们的对象（Object）系统中，只有一个`TRUE`布尔对象用来表示真，只有一个`FALSE`布尔对象用来表示假，只有一个`NIL`对象用来表示nil值。

反过来说，就是所有的【真值】都用`TRUE`这个布尔对象表示，所有的【假值】都用`FALSE`这个布尔对象表示，所有的【nil值】都用`NIL`这个对象来表示。



有了上面的布尔对象和Nil对象，我们就可以修改解释器（Evaluator）的代码了。



## 解释器（Evaluator）的更改

我们需要在解释器（Evaluator）的`Eval`函数的`switch`分支中加入对布尔类型和Nil类型的处理：

```go
//eval.go

func Eval(node ast.Node) (val Object) {
	switch node := node.(type) {
	//...

	case *ast.BooleanLiteral:
		return nativeBoolToBooleanObject(node.Value)
	case *ast.NilLiteral:
		return NIL //直接返回唯一的NIL对象
	}

	return nil
}

//将go的bool类型转换为布尔对象返回
func nativeBoolToBooleanObject(input bool) *Boolean {
	if input {
		return TRUE //返回唯一的布尔真值'TRUE'
	}
	return FALSE //返回唯一的布尔假值'FALSE'
}
```

我们已经完成了对解释器的更改，就是这么简单。哦耶！！！



### 测试解释器

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
		{"2 + 2 ** 2 ** 3", "258"},
		{"10", "10"},
		{"nil", "nil"},
		{"true", "true"},
		{"false", "false"},
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



是不是觉得渐入佳境了? 也许吧。。。但愿吧。。。



下一节，我们将加入对`let`语句（statement）的支持。

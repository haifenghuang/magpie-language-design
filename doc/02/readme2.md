# 识别标识符（identifier）

上一节，我们实现了一个简单的【四则运算】计算器。这一节开始，我们将要在上一节的基础上，加入识别标识符（identifier）的逻辑。

那么我们需要做哪些更改呢？

1. 在词元（Token）源码`token.go`中加入新的词元（Token）类型。
2. 在词法分析器（Lexer）源码`lexer.go`中加入对标识符的识别。
3. 在抽象语法树（AST）的源码`ast.go`中加入标识符的抽象语法表示。
4. 在语法解析器（Parser）的源码`parser.go`中加入对`标识符`的语法解释。

实现起来的话，可能比你想象的简单。

## 词元（Token）更改

### 第一处改动

```go
//token.go
const (
	TOKEN_ILLEGAL TokenType = (iota - 1) // Illegal token
	TOKEN_EOF                            //End Of File

    //...

	TOKEN_NUMBER     //10 or 10.1
    TOKEN_IDENTIFIER //identifier: a, b, var1, ...
)
```

第9行，我们加入了一个新的词元（Token）类型`TOKEN_IDENTIFIER`。

### 第二处改动

```go
//token.go
//词元类型的字符串表示
func (tt TokenType) String() string {
	switch tt {
	case TOKEN_EOF:
		return "EOF"

    //...

	case TOKEN_NUMBER:
		return "NUMBER"
	case TOKEN_IDENTIFIER:
		return "IDENTIFIER"
	default:
		return "UNKNOWN"
	}
}
```

12-13行，我们加入了一个`case TOKEN_IDENTIFIER`的分支。

### 第三处改动

```go
//token.go

//关键字map
var keywords = map[string]TokenType{}

//判断传入的变量'ident'是关键字，还是一个标识符
func LookupIdent(ident string) TokenType {
	if tok, ok := keywords[ident]; ok {
		return tok
	}
	return TOKEN_IDENTIFIER
}
```

这里我们增加了一个存放关键字的map（当前我们还没有实现任何的关键字），和判断变量是否为关键字还是普通的标识符的函数。

将来我们会往这个map中加入`magpie`语言的关键字（keyword），类似如下：

```go
//token.go
var keywords = map[string]TokenType{
	"true":  TOKEN_TRUE,
	"false": TOKEN_FALSE,
	"nil":   TOKEN_NIL,
}
```

> 关键字也是标识符，只不过它们在语言中有特殊的含义。通常你不能随意在语言中使用，比如声明一个变量的名字是关键字：
>
> ```go
> let if = 10
> ```
>
> 由于`if`是个关键字，所以不能够将其作为变量名。



## 词法解析器（Lexer）的更改

### 第一处更改

```go
//lexer.go
func (l *Lexer) NextToken() token.Token {
	var tok token.Token
	l.skipWhitespace()

	pos := l.getPos()

	switch l.ch {
	case '+':
		tok = newToken(token.TOKEN_PLUS, l.ch)

    //。。。

	default:
		if isDigit(l.ch) {
			tok.Literal = l.readNumber()
			tok.Type = token.TOKEN_NUMBER
			tok.Pos = pos
			return tok
		} else if isLetter(l.ch) { //如果是字母
			tok.Literal = l.readIdentifier()
			tok.Pos = pos
			tok.Type = token.LookupIdent(tok.Literal)//调用LookupIdent函数判断标识符是否为普通标识符还是关键字
			return tok
		} else {
			tok = newToken(token.TOKEN_ILLEGAL, l.ch)
		}
	}

	tok.Pos = pos
	l.readNext()
	return tok
}
```

第20-24行，我们在`NextToken()`函数中增加了一个判断标识符（Identifier）的分支。



### 第二处更改

```go
//lexer.go

//标识符：ident_1, 三, var1, ...
func (l *Lexer) readIdentifier() string {
	position := l.position
	for isLetter(l.ch) || isDigit(l.ch) {//如果是字符或者数字就继续
		l.readNext()
	}
	return string(l.input[position:l.position])
}

//判断传入的'ch'字符是否为字母
func isLetter(ch rune) bool {
	return unicode.IsLetter(ch) || ch == '_'
}
```

这里增加了一个读取标识符的函数`readIdentifier()`。还增加了一个判断字符是否为字母的函数`isLetter()`。这里唯一需要注意的是`isLetter()`函数中的`unicode.IsLetter`这个判断。有了这个判断，标识符中就可以包含中文：

```go
姓名 = "黄海峰"
```

> 这里说一个题外话，很多语言都要求，标识符不能以数字开头，但是之后可以是数字。这是为什么呢？其实很简单，假设我声明一个变量如下：
>
> ```javascript
> let 333 = 10
> ```
>
> 变量名为`333`，它是以数字`3`开头的。这里我们将数字`10`赋值给`333`。是不是会让人刚到很困惑？



### 测试词法解析器

```go
//main.go
func TestLexer() {
	input := " 2 + (3 * 4) / ( 5 - 3 ) + 10 -  年龄 * 2 + a ** 3"
	fmt.Printf("Input = %s\n", input)

	l := lexer.NewLexer(input)
	for {
		tok := l.NextToken()
		fmt.Printf("%s\n", tok)
		if tok.Type == token.TOKEN_EOF {
			break
		}
	}
}

func main() {
	TestLexer()
}
```

输出结果（为了编译查看，做了相应的格式化）：

```
Input =  2 + (3 * 4) / ( 5 - 3 ) + 10 -  年龄 * 2 + abc ** 3
Position:  <1:3> ,      Type: NUMBER,       Literal: 2
Position:  <1:5> ,      Type: +,            Literal: +
Position:  <1:7> ,      Type: (,            Literal: (
Position:  <1:8> ,      Type: NUMBER,       Literal: 3
Position:  <1:10> ,     Type: *,            Literal: *
Position:  <1:12> ,     Type: NUMBER,       Literal: 4
Position:  <1:13> ,     Type: ),            Literal: )
Position:  <1:15> ,     Type: /,            Literal: /
Position:  <1:17> ,     Type: (,            Literal: (
Position:  <1:19> ,     Type: NUMBER,       Literal: 5
Position:  <1:21> ,     Type: -,            Literal: -
Position:  <1:23> ,     Type: NUMBER,       Literal: 3
Position:  <1:25> ,     Type: ),            Literal: )
Position:  <1:27> ,     Type: +,            Literal: +
Position:  <1:29> ,     Type: NUMBER,       Literal: 10
Position:  <1:32> ,     Type: -,            Literal: -
Position:  <1:35> ,     Type: IDENTIFIER,   Literal: 年龄
Position:  <1:38> ,     Type: *,            Literal: *
Position:  <1:40> ,     Type: NUMBER,       Literal: 2
Position:  <1:42> ,     Type: +,            Literal: +
Position:  <1:44> ,     Type: IDENTIFIER,   Literal: abc
Position:  <1:48> ,     Type: **,           Literal: **
Position:  <1:51> ,     Type: NUMBER,       Literal: 3
Position:  <1:51> ,     Type: EOF,          Literal: <EOF>
```

可以看到第18行和第22行都正确的识别出了标识符。

> 注意：这里最后一个词元是EOF，它的Position信息（即位置信息）没有任何意义。



## 抽象语法树（AST）的更改

对于程序中的标识符（Identifier），我们也需要为其增加`标识符节点（Identifier Node）`表示。

```go
//ast.go

// 标识符节点
// var1, 姓名, ...
type Identifier struct {
	Token token.Token
	Value string //标识符的字面量表示
}

//开始位置
func (i *Identifier) Pos() token.Position { return i.Token.Pos }
//终了位置 = 开始位置 + 标识符的长度
func (i *Identifier) End() token.Position {
	length := utf8.RuneCountInString(i.Value)
	return token.Position{Filename: i.Token.Pos.Filename, Line: i.Token.Pos.Line, 
						  Col: i.Token.Pos.Col + length}
}
//标识符是一个表达式节点
func (i *Identifier) expressionNode()      {}
func (i *Identifier) TokenLiteral() string { return i.Token.Literal }
func (i *Identifier) String() string       { return i.Value }
```

和我们在前一节中讲的`NumberLiteral`节点的内容很类似，对吧？



## 语法解析器（Parser）的更改

我们需要做两处更改：

1. 对新增加的`TOKEN_IDENTIFIER`词元类型注册前缀表达式回调函数（代码第7行）
2. 解析标识符表达式的函数`parseIdentifier()`。（代码中的13-16行）

```go
//parser.go
func (p *Parser) registerAction() {
	p.prefixParseFns = make(map[token.TokenType]prefixParseFn)
	p.registerPrefix(token.TOKEN_NUMBER, p.parseNumber)
    
    //给`TOKEN_IDENTIFIER`词元类型注册前缀表达式回调函数
	p.registerPrefix(token.TOKEN_IDENTIFIER, p.parseIdentifier)

	//...
}

//解析标识符表达式
func (p *Parser) parseIdentifier() ast.Expression {
    //返回标识符表达式节点
	return &ast.Identifier{Token: p.curToken, Value: p.curToken.Literal}
}
```



### 测试语法解析器

```go
//main.go
func TestParser() {
	input := " 2 + (3 * 4) / ( 5 - 3 ) + 10 + 2 ** 2 ** 3 + xyz"
	expected := "((((2 + ((3 * 4) / (5 - 3))) + 10) + (2 ** (2 ** 3))) + xyz)"
	l := lexer.NewLexer(input)
	p := parser.NewParser(l)
	program := p.ParseProgram()
	if len(p.Errors()) != 0 {
		for _, err := range p.Errors() {
			fmt.Println(err)
		}
		os.Exit(1)
	}
	if program.String() != expected {
		fmt.Printf("Syntax error: expected %s, got %s\n", expected, program.String())
		os.Exit(1)
	}

	fmt.Printf("input  = %s\n", input)
	fmt.Printf("output = %s\n", program.String())
}

func main() {
	TestParser()
}
```

运行结果如下：

```
input  =  2 + (3 * 4) / ( 5 - 3 ) + 10 + 2 ** 2 ** 3 + xyz
output = ((((2 + ((3 * 4) / (5 - 3))) + 10) + (2 ** (2 ** 3))) + xyz)
```

我们可以看到`xyz`标识符被正确的解析了。

上面就是所要修改的全部内容，是不是比想象中的要简单？



> 注：上面的更改并没有包含对解释器（Evaluator）的更改。因为对于解释器的更改涉及到其它的一些我们还未学到的内容，这些内容将会在后续的章节中详细说明。



下一节，我们将加入`true`, `false`和`nil`的支持。

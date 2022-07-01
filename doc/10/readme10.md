# 字符串表达式支持

在这一节中，我们要加入对于`字符串表达式`的支持。

我们需要做如下的更改：

1. 在词元（Token）源码`token.go`中加入一个新的词元（Token）类型
2. 在词法分析器（Lexer）源码`lexer.go`中加入对`字符串`的识别
3. 在抽象语法树（AST）的源码`ast.go`中加入`字符串`对应的抽象语法表示。
4. 在语法解析器（Parser）的源码`parser.go`中加入对`字符串`的语法解析。
4. 在对象（Object）系统中的源码`object.go`中加入一个新的`字符串对象(String Object)`。
5. 在解释器（Evaluator）的源码`eval.go`中加入对`字符串`的解释。

## 词元（Token）更改

### 第一处改动

```go
//token.go
const (
	//...
	TOKEN_STRING     //""
```



### 第二处改动

```go
//token.go
//词元类型的字符串表示
func (tt TokenType) String() string {
	switch tt {
	//...
	case TOKEN_STRING:
		return "STRING"
	}
}
```



## 词法分析器（Lexer）的更改

我们需要在词法分析器（Lexer）的`NextToken()`函数中加入对`字符串`的识别：

```go
//lexer.go

//获取下一个词元（Token）
func (l *Lexer) NextToken() token.Token {
	//...
	switch l.ch {
	//...
	default:
		if isDigit(l.ch) {
			tok.Literal = l.readNumber()
			tok.Type = token.TOKEN_NUMBER
			tok.Pos = pos
			return tok
		}
		//...
		} else if l.ch == 34 { //34的ASCII码对应的是双引号
			if s, err := l.readString(l.ch); err == nil {
				tok.Type = token.TOKEN_STRING
				tok.Pos = pos
				tok.Literal = s
				return tok
			} else {
				tok.Type = token.TOKEN_ILLEGAL
				tok.Pos = pos
				tok.Literal = err.Error()
				return tok
			}
		} else {
			tok = newToken(token.TOKEN_ILLEGAL, l.ch)
		}
	}
}

```

第16行的`else if`判断，我们加入了对`字符串`的识别。下面是实际识别`字符串`的代码：

```go
//lexer.go
func (l *Lexer) readString(r rune) (string, error) {
	var ret []rune
eos:
	for {
		l.readNext()
		switch l.ch {
		case '\n':
			return "", errors.New("unexpected EOL")
		case 0:
			return "", errors.New("unexpected EOF")
		case r: //遇到了字符串结束符
			l.readNext()
			break eos //eos:end of string
		case '\\': //如果有转义字符
			l.readNext()
			switch l.ch {
			case 'b':
				ret = append(ret, '\b')
				continue
			case 'f':
				ret = append(ret, '\f')
				continue
			case 'r':
				ret = append(ret, '\r')
				continue
			case 'n':
				ret = append(ret, '\n')
				continue
			case 't':
				ret = append(ret, '\t')
				continue
			}
			ret = append(ret, l.ch)
			continue
		default:
			ret = append(ret, l.ch)
		}
	}

	return string(ret), nil
}
```

上面的代码看起来有点多。处理逻辑大致如下：

循环读取下一个字符，如果遇到字符为`双引号`（即字符串结束符）的话就退出循环。如果遇到了文件结束标志（EOF），那么就表明处理到文件结束也没有找到字符串的结束标记，就报错。其它的处理主要是处理字符串中的转义字符。



## 抽象语法树（AST）的更改

`字符串`的抽象语法表示与我们第一节介绍的数字字面量（`NumberLiteral`）的抽象语法表示几乎类似，下面直接上代码：

```go
//ast.go
//'字符串'的抽象语法表示
type StringLiteral struct {
	Token token.Token
	Value string //字符串的值
}

func (s *StringLiteral) Pos() token.Position {
	return s.Token.Pos
}

//结束位置 = 开始位置 + 字符串的长度
func (s *StringLiteral) End() token.Position {
	length := utf8.RuneCountInString(s.Value)
	return token.Position{Filename: s.Token.Pos.Filename, Line: s.Token.Pos.Line, 
                          Col: s.Token.Pos.Col + length}
}

func (s *StringLiteral) expressionNode()      {}
func (s *StringLiteral) TokenLiteral() string { return s.Token.Literal }
func (s *StringLiteral) String() string       { return s.Token.Literal }
```

这里的代码也比较简单，也无需做太多解释。



## 语法解析器（Parser）的更改

我们需要做两处更改：

1. 对新增加的`TOKEN_STRING`词元类型注册前缀表达式回调函数（代码中的5行）
2. 新增解析字符串表达式的函数`parseStringLiteral()`（代码中的9-11行）

```go
//parser.go
func (p *Parser) registerAction() {
	p.prefixParseFns = make(map[token.TokenType]prefixParseFn)
	//...
	p.registerPrefix(token.TOKEN_STRING, p.parseStringLiteral) //对字符串注册前缀表达式回调函数
}

//解析字符串
func (p *Parser) parseStringLiteral() ast.Expression {
	return &ast.StringLiteral{Token: p.curToken, Value: p.curToken.Literal}
}

```



## 对象系统的更改

我们需要往对象系统中增加一个`字符串对象(String Object)`。因为也是我们熟悉的代码且比较简单，所以也是直接上代码：

```go
//object.go

const (
	//...
	STRING_OBJ       = "STRING"
)

//字符串对象
type String struct {
	String string //存储字符串的值
}

func (s *String) Inspect() string {
	return s.String
}

func (s *String) Type() ObjectType { return STRING_OBJ }

//一个工具(utility)函数，用来生成新的字符串对象
func NewString(s string) *String {
	return &String{String: s}
}
```



## 解释器（Evaluator）的更改

我们需要在解释器（Evaluator）的`Eval`函数的`switch`分支中加入对`字符串表达式`的处理：

```go
//eval.go

func Eval(node ast.Node, scope *Scope) (val Object) {
	switch node := node.(type) {
	//...
	case *ast.StringLiteral: //字符串表达式
		return evalStringLiteral(node, scope)
	}

	return nil
}

//解释'字符串表达式'
func evalStringLiteral(s *ast.StringLiteral, scope *Scope) Object {
	return NewString(s.Value) //生成一个新的字符串对象返回
}
```



另外，我们希望给字符串加入连接操作符，即允许两个字符串相加，生成一个新的字符串，类似下面这样：

```go
let x = "Hello " + "World!" // x = "Hello World!"
```

因此我们需要修改`evalInfixExpression()`函数：

```go
//eval.go
func evalInfixExpression(node *ast.InfixExpression, left, right Object, scope *Scope) Object {
	switch {
	case left.Type() == NUMBER_OBJ && right.Type() == NUMBER_OBJ:
		return evalNumberInfixExpression(node, left, right, scope)
	case left.Type() == STRING_OBJ && right.Type() == STRING_OBJ:
		return evalStringInfixExpression(node, left, right, scope)
}

func evalStringInfixExpression(node *ast.InfixExpression, left, right Object, scope *Scope) Object {
	leftVal := left.(*String).String //取出字符串对象中保存的字符串
	rightVal := right.(*String).String

	switch node.Operator {
	case "+":
		return NewString(leftVal + rightVal)
	default:
		return newError(node.Pos().Sline(), ERR_INFIXOP, 
                        left.Type(), node.Operator, right.Type())
	}
}
```

我们给`evalInfixExpression()`函数的`switch`语句增加了一个`case`分支。第10行的`evalStringInfixExpression`函数是实现字符串拼接的函数。



是不是有一种【水到渠成】的感觉？别漂，我们的喜鹊（magpie）的翅膀还没有涨硬呢。



## 测试

下面我们写一个简单的程序测试一下：
```go
//main.go
func TestEval() {
	tests := []struct {
		input    string
		expected string
	}{
		{`let x = "hello world"; x`, "hello world"},
        {`let x = "hello " + "world"; x`, "hello world"}
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

通过前十节的学习，我们的小喜鹊（magpie）的五脏六腑已经长全，也有了强有力的腿部，翅膀也涨慢慢变得更硬了。

第二部分，小喜鹊就将进入一个新的领域：飞行。

# `块语句`支持

在这一篇文章中，我们要加入对于`块语句`的支持。所谓的`块语句`其实就是`{}`包起来的语句块。

我们需要做如下的更改：

1. 在词元（Token）源码`token.go`中加入两个新的词元（Token）类型（`{`和`}`）
2. 在词法分析器（Lexer）源码`lexer.go`中加入对`{`和`}`的识别
3. 在抽象语法树（AST）的源码`ast.go`中加入`块语句`对应的抽象语法表示。
4. 在语法解析器（Parser）的源码`parser.go`中加入对`块语句`的语法解析。
5. 在解释器（Evaluator）的源码`eval.go`中加入对`块语句`的解释。

## 词元（Token）更改

### 第一处改动

```go
//token.go
const (
	//...
	TOKEN_LBRACE    // {
	TOKEN_RBRACE    // }
)
```



### 第二处改动

```go
//token.go
//词元类型的字符串表示
func (tt TokenType) String() string {
	switch tt {
	//...
	case TOKEN_LBRACE:
		return "{"
	case TOKEN_RBRACE:
		return "}"
	}
}
```



## 词法分析器（Lexer）的更改

我们需要在词法分析器（Lexer）的`NextToken()`函数中加入对`{`和`}`的识别：

```go
//lexer.go

//获取下一个词元（Token）
func (l *Lexer) NextToken() token.Token {
	//...
    switch l.ch {
	//...
	case '{':
		tok = newToken(token.TOKEN_LBRACE, l.ch)
	case '}':
		tok = newToken(token.TOKEN_RBRACE, l.ch)
	//...
    }

	//...
}
```



## 抽象语法树（AST）的更改

我们的`块语句`，需要什么信息呢？

1. 词元信息
2. 语句数组（`块语句`中可以包含多条语句）
2. 右花括弧的位置（这个主要是为了计算`块语句`的结束位置用）

```go
//ast.go

//块语句:
// {
//    stmt1
//    stmt2
//    ...
//  }
// 
type BlockStatement struct {
	Token       token.Token
	Statements  []Statement
	RBraceToken token.Token //'End()'方法会使用
}

//开始位置
func (bs *BlockStatement) Pos() token.Position {
	return bs.Token.Pos
}

//结束位置
func (bs *BlockStatement) End() token.Position {
	return token.Position{Filename: bs.Token.Pos.Filename, Line: 
                          bs.RBraceToken.Pos.Line, Col: bs.RBraceToken.Pos.Col + 1}
}

func (bs *BlockStatement) statementNode() {}

func (bs *BlockStatement) TokenLiteral() string { return bs.Token.Literal }

//块语句的字符串表示
func (bs *BlockStatement) String() string {
	var out bytes.Buffer

    //循环遍历块语句中的每个statement
	for _, s := range bs.Statements {
		str := s.String()

		out.WriteString(str)
		if str[len(str)-1:] != ";" {//如果statement的字符串表示的最后一个字符不是';'号的话
			out.WriteString(";")
		}
	}

	return out.String()
}
```

这种代码我们已经见过太多了，就不再解释了。



## 语法解析器（Parser）的更改

我们需要做两处更改：

1. 在`parseStatement()`函数的`switch`分支中加入对词元类型为`TOKEN_LBRACE`的判断
2. 增加解析`块语句`的函数

```go
//parser.go
func (p *Parser) parseStatement() ast.Statement {
	switch p.curToken.Type {
	//...
    case token.TOKEN_LBRACE: // '{'
		return p.parseBlockStatement()
    default:
		return p.parseExpressionStatement()
	}
}

//解析'块语句'
func (p *Parser) parseBlockStatement() *ast.BlockStatement {
	blockStmt := &ast.BlockStatement{Token: p.curToken} //生成块语句节点

	p.nextToken()
    for !p.curTokenIs(token.TOKEN_RBRACE) {//如果没有遇到右'}'，就继续解析
		stmt := p.parseStatement() //解析块语句中的每个语句
		if stmt != nil {
			blockStmt.Statements = append(blockStmt.Statements, stmt)
		}
		if p.peekTokenIs(token.TOKEN_EOF) {
			break
		}
		p.nextToken()
	}

    blockStmt.RBraceToken = p.curToken //记录右'}'的位置
	return blockStmt
}
```

我们在`ParseStatement()`函数的`switch`分支中增加了对词元类型为`TOKEN_LBRACE)`即`{`的判断（第5-6行）。

同时增加了对`块语句`的解析（代码13-30行）。



## 解释器（Evaluator）的更改

我们需要在解释器（Evaluator）的`Eval`函数的`switch`分支中加入对`块语句`的处理：

```go
//eval.go

func Eval(node ast.Node, scope *Scope) (val Object) {
	switch node := node.(type) {
	//...
	case *ast.BlockStatement:
		return evalBlockStatement(node, scope)
	//...
	}

	return nil
}

//解释'块语句'
func evalBlockStatement(block *ast.BlockStatement, scope *Scope) Object {
	var result Object

	//循环解释块语句中的每一个语句
	for _, statement := range block.Statements {
		result = Eval(statement, scope)
		if result != nil {
			rt := result.Type()
		 	//如果处理结果为`return类型`或者'错误类型'就提前结束循环，无需处理后续的语句
			if rt == RETURN_VALUE_OBJ || rt == ERROR_OBJ {
				return result
			}
		}
	}
	return result
}
```

一切都是熟悉的味道！！！一切都是熟悉的老干妈!:smile:



## 测试

下面我们写一个简单的程序测试一下`块语句`：
```go
//main.go
func TestEval() {
	tests := []struct {
		input    string
		expected string
	}{
		{`{ let x = 10 { x } }`, "10"},
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

第7行看起来比较费劲，我们把它格式化一下：

```{}
{
   let x = 10
   {
       x
   }
}
```



下一节，我们将加入对`字符串表达式`的支持。

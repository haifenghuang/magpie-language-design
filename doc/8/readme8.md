# `return`语句支持

在这一篇文章中，我们要加入对于`return`语句的处理，为将来要说的函数提供返回值支持。

我们需要做如下的更改：

1. 在词元（Token）源码`token.go`中加入新的词元（Token）类型
2. 在抽象语法树（AST）的源码`ast.go`中加入`return`语句对应的抽象语法表示。
3. 在语法解析器（Parser）的源码`parser.go`中加入对`return`语句的语法解析。
4. 在对象（Object）源码`object.go`中加入新的对象类型（返回值对象）
5. 在解释器（Evaluator）的源码`eval.go`中加入对`return`语句的解释。

## 词元（Token）更改

### 第一处改动

```go
//token.go
const (
    //...
    
    //reserved keywords
	TOKEN_RETURN //return
)
```

第6行，我们加入了一个新的词元（Token）类型。

### 第二处改动

```go
//token.go
//词元类型的字符串表示
func (tt TokenType) String() string {
	switch tt {
    //...

	case TOKEN_RETURN:
		return "RETURN"
	}
}
```



### 第三处改动

```go
//token.go

//关键字map
var keywords = map[string]TokenType{
    //...
    "return": TOKEN_RETURN,
}
```



## 抽象语法树（AST）的更改

我们的`return`语句，需要什么信息呢？

1. 词元信息（我们的老朋友了，调试、输出或者报错用）
2. 返回值表达式

```go
//ast.go

//return语句
type ReturnStatement struct {
	Token       token.Token // the 'return' token
	ReturnValue Expression //返回的表达式。如果为nil，表明不返回任何值
}

//开始位置
func (rs *ReturnStatement) Pos() token.Position {
	return rs.Token.Pos
}

//结束位置
func (rs *ReturnStatement) End() token.Position {
	if rs.ReturnValue == nil { //如果没有返回值
		length := utf8.RuneCountInString(rs.Token.Literal)
		pos := rs.Token.Pos
		return token.Position{Filename: pos.Filename, Line: pos.Line, Col: pos.Col + length}
	}
	return rs.ReturnValue.End()
}

//表面'return'是一个语句
func (rs *ReturnStatement) statementNode()       {}

func (rs *ReturnStatement) TokenLiteral() string { return rs.Token.Literal }

//'return'语句的字符串表示
func (rs *ReturnStatement) String() string {
	var out bytes.Buffer

	out.WriteString(rs.TokenLiteral() + " ")

	if rs.ReturnValue != nil {
		out.WriteString(rs.ReturnValue.String())
	}

	out.WriteString("; ")

	return out.String()
}
```

这都是读者比较熟悉的代码了，所以也没有太多需要解释的。



## 语法解析器（Parser）的更改

我们需要做两处更改：

1. 在`parseStatement()`函数的`switch`分支中加入对词元类型为`TOKEN_RETURN`的判断
2. 增加解析`return`语句的函数

```go
//parser.go
func (p *Parser) parseStatement() ast.Statement {
	switch p.curToken.Type {
	//...
	case token.TOKEN_RETURN:
		return p.parseReturnStatement()

	}
}

//解析'return'语句:
// 1. return <expression>;
// 2. return;
func (p *Parser) parseReturnStatement() *ast.ReturnStatement {
	stmt := &ast.ReturnStatement{Token: p.curToken}
	if p.peekTokenIs(token.TOKEN_SEMICOLON) { //e.g.{ return; }
		p.nextToken()
		return stmt
	}

	p.nextToken()
	stmt.ReturnValue = p.parseExpressionStatement().Expression

	return stmt
}
```

我们在`ParseStatement()`函数的`switch`分支中增加了对词元类型为`TOKEN_RETURN`的判断（第5-6行）。

同时增加了对`return`语句的解析（代码14-25行）。



## 对象（Object）的更改

由于我们新增了一个`return`类型，我们就需要在对象（Object）系统中加入对这个新的类型的支持。

1. 新增一个对象类型（Object Type）
2. 实现这个新增对象类型（即实现`object`接口的方法）

### 新增一个对象类型（Object Type）

```go
//object.go
const (
	//...
	RETURN_VALUE_OBJ = "RETURN_VALUE"
)
```



### 实现`return`对象类型（即实现`object`接口的方法）

```go
//object.go

//'return'对象
type ReturnValue struct {
	Value Object
}

func (rv *ReturnValue) Type() ObjectType { return RETURN_VALUE_OBJ }
func (rv *ReturnValue) Inspect() string  { return rv.Value.Inspect() }
```

内容也比较简单，就不多解释了。



## 解释器（Evaluator）的更改

我们需要在解释器（Evaluator）的`Eval`函数的`switch`分支中加入对`return`类型的处理：

```go
//eval.go

func Eval(node ast.Node, scope *Scope) (val Object) {

	switch node := node.(type) {
	//...
	case *ast.ReturnStatement:
        if node.ReturnValue == nil {
			return &ReturnValue{Value: NIL} //如果没有返回值，我们默认返回NIL对象
		}

        val := Eval(node.ReturnValue, scope)
		return &ReturnValue{Value: val} //返回'return'对象
	}

	return nil
}
```

我们还剩下一个工作，在解释程序（Program）节点的函数中（`parseProgram`），我们也需要对返回值进行处理：

```go
//eval.go

//解释程序节点
func evalProgram(program *ast.Program, scope *Scope) (results Object) {
	for _, stmt := range program.Statements {
		results = Eval(stmt, scope)
		if returnValue, ok := results.(*ReturnValue); ok {
			return returnValue.Value //返回'return'对象中保存的返回值
		}
	}

	if results == nil {
		return NIL
	}
	return results
}
```

上面代码的第7-9行为新增代码，如果处理的语句的返回值为`return对象`的时候，就返回`return对象`中存储的值。

至此，对`return语句`的支持就完成了。后续的章节我们会扩展这个`return语句`，使其支持多个返回值。



下一节，我们将加入对`块语句（block statement）`的支持。

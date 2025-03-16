# `try-catch-finally`支持

这一节中，我们将给`magpie语言`提供对异常处理的支持。先来看一下使用例，让读者对这一节要实现的功能有个大概的了解：

```javascript
try {
    let th = 1 + 2
    println(th)
    #throw "Hello"
    throw 10
    println("After throw")
}
catch e {
    if type(e) == "string" {
        printf("Catched, e=%s\n", e)
    } else if type(e) == "number" {
        printf("Catched, e=%d\n", e)
    }
}
finally {
    println("Finally running")
}

println("After try\n\n")
```

这里有几点需要注意：

1. `catch e`中的`e`可以省略。
2. `catch`和`finally`语句都可以省略。
3. 如果有`finally`语句，那么无论`try`语句中是否抛出异常，它都会运行。
4. 如果`try`语句中抛出了异常，而没有`catch`语句来捕获异常，程序会报告异常没有捕获。



现在让我们看一下需要做哪些更改：

1. 在词元（Token）源码`token.go`中加入新的词元（Token）类型
2. 在抽象语法树（AST）的源码`ast.go`中加入`try语句`和`throw语句`对应的抽象语法表示
3. 在语法解析器（Parser）的源码`parser.go`中加入对`try语句`和`throw语句`的语法解析
4. 在对象（Object）系统源码`object.go`中，新增一个`Throw对象(Throw Object)`
5. 在解释器（Evaluator）的源码`eval.go`，加入对`throw语句`和`try语句`的解释

## 词元（Token）的更改

因为都是读者再熟悉不过的内容，不做解释，直接看代码：

```go
//token.go
const (
	//...
	TOKEN_TRY      //try
	TOKEN_CATCH    //catch
	TOKEN_FINALLY  //finally
	TOKEN_THROW    //throw
)

//词元的字符串表示
func (tt TokenType) String() string {
	switch tt {
	//...
	case TOKEN_TRY:
		return "TRY"
	case TOKEN_CATCH:
		return "CATCH"
	case TOKEN_FINALLY:
		return "FINALLY"
	case TOKEN_THROW:
		return "THROW"
	}
}

var keywords = map[string]TokenType{
	//...
	"try":      TOKEN_TRY,
	"catch":    TOKEN_CATCH,
	"finally":  TOKEN_FINALLY,
	"throw":    TOKEN_THROW,
}
```



## 抽象语法树（AST）的更改

从上面的例子中我们可以得出`throw语句`的一般表示形式如下：

```javascript
throw <expression>
```

`try语句`的一般表示形式如下：

```javascript
try { block }
catch e { block } //'e'也可以省略
finally { block }
```

有了上面的说明，我们可以很容易得到它们的抽象语法表示：

```go
//ast.go
/*
   try {block }
   catch e { block }
   finally {block }
*/

type TryStmt struct {
	Token   token.Token
	Try     *BlockStatement
	Var     string //对应catch中的'e', 可以为空
	Catch   *BlockStatement
	Finally *BlockStatement
}

func (t *TryStmt) Pos() token.Position {
	return t.Token.Pos
}

func (t *TryStmt) End() token.Position {
	if t.Finally != nil {
		return t.Finally.End()
	}

	if t.Catch != nil {
		return t.Catch.End()
	}

	return t.Try.End()
}

func (t *TryStmt) statementNode()       {}
func (t *TryStmt) TokenLiteral() string { return t.Token.Literal }

func (t *TryStmt) String() string {
	var out bytes.Buffer

	out.WriteString("try { ")
	out.WriteString(t.Try.String())
	out.WriteString(" }")

	if t.Catch != nil {
		if t.Var != "" {
			out.WriteString(" catch " + t.Var + " { ")
		} else {
			out.WriteString(" catch { ")
		}
		out.WriteString(t.Catch.String())
		out.WriteString(" }")
	}

	if t.Finally != nil {
		out.WriteString(" finally { ")
		out.WriteString(t.Finally.String())
		out.WriteString(" }")
	}

	return out.String()
}

//throw <expression>
type ThrowStmt struct {
	Token token.Token
	Expr  Expression
}

func (ts *ThrowStmt) Pos() token.Position {
	return ts.Token.Pos
}

func (ts *ThrowStmt) End() token.Position {
	return ts.Expr.End()
}

func (ts *ThrowStmt) statementNode()       {}
func (ts *ThrowStmt) TokenLiteral() string { return ts.Token.Literal }

func (ts *ThrowStmt) String() string {
	var out bytes.Buffer

	out.WriteString("throw ")
	out.WriteString(ts.Expr.String())
	out.WriteString(";")

	return out.String()
}
```

内容都是读者见过很多次的，也不做解释了。

## 语法解析器（Parser）的更改

由于加入了`try语句`和`throw语句`的支持，我们需要在`parseStatement`函数中加入两个`case`分支：

```go
//parser.go

func (p *Parser) parseStatement() ast.Statement {
	switch p.curToken.Type {
	//...
	case token.TOKEN_TRY:
		return p.parseTryStatement()
	case token.TOKEN_THROW:
		return p.parseThrowStatement()
	}
}

//解析try语句
func (p *Parser) parseTryStatement() ast.Statement {
	tryStmt := &ast.TryStmt{Token: p.curToken}

	p.nextToken()
	tryStmt.Try = p.parseBlockStatement() //解析'try'块

	if p.peekTokenIs(token.TOKEN_CATCH) {
		p.nextToken() //skip '}'

        if p.peekTokenIs(token.TOKEN_IDENTIFIER) { //判断是否有标识符:'catch e'中的'e'
			p.nextToken()
			tryStmt.Var = p.curToken.Literal
		}

		if !p.expectPeek(token.TOKEN_LBRACE) {
			return nil
		}

		tryStmt.Catch = p.parseBlockStatement() //解析'catch'块
	}

	if p.peekTokenIs(token.TOKEN_FINALLY) { //解析'finally'块
		p.nextToken() //skip '}'
		if !p.expectPeek(token.TOKEN_LBRACE) {
			return nil
		}

		tryStmt.Finally = p.parseBlockStatement()
	}

	return tryStmt
}

//解析throw语句: throw <expression>
func (p *Parser) parseThrowStatement() *ast.ThrowStmt {
	stmt := &ast.ThrowStmt{Token: p.curToken}
	if p.peekTokenIs(token.TOKEN_SEMICOLON) {
		p.nextToken()
		return stmt
	}
	p.nextToken()
	stmt.Expr = p.parseExpressionStatement().Expression

	return stmt

}
```

代码也是比较简单的。

## 对象（Object）系统的更改

我们需要创建一个`Throw对象`。这个`Throw对象`中应该包含什么样的信息呢？

1. throw的值（Value）
2. `throw语句(throw-statement)`（用来报告错误用）

来看一下代码：

```go
//object.go
const (
	//...
	THROW_OBJ       = "THROW"
)

//throw对象
type Throw struct {
	stmt  *ast.ThrowStmt //报错用
	value Object
}

func (t *Throw) Inspect() string  { return t.value.Inspect() }
func (t *Throw) Type() ObjectType { return THROW_OBJ }
func (t *Throw) CallMethod(line string, scope *Scope, method string, args ...Object) Object {
	return newError(line, ERR_NOMETHOD, method, t.Type())
}
```

## 解释器（Evaluator）的更改

首先，我们需要在`Eval()`函数的`switch`语句中增加两个`case`分支：

```go
//eval.go

func Eval(node ast.Node, scope *Scope) (val Object) {
    //...

	switch node := node.(type) {
	//...
	case *ast.TryStmt:
		return evalTryStatement(node, scope)
	case *ast.ThrowStmt:
		return evalThrowStatement(node, scope)

	}

	return nil
}

//解释throw
func evalThrowStatement(t *ast.ThrowStmt, scope *Scope) Object {
	throwObj := Eval(t.Expr, scope)
	if throwObj.Type() == ERROR_OBJ {
		return throwObj
	}

	return &Throw{stmt: t, value: throwObj} //返回一个throw对象
}

//解释try-catch-finally语句
func evalTryStatement(tryStmt *ast.TryStmt, scope *Scope) Object {
	rv := Eval(tryStmt.Try, scope) //解释try块
	if rv.Type() == ERROR_OBJ {
		return rv
	}

	//如果有'catch e'这样的语句出现，我们会将其加入scope中，以便在catch块中使用这个变量。
	//而这个'e'只在这个catch块中有效，所以结束后，我们需要将其从scope中删除
	defer func() {
		if tryStmt.Catch != nil {
			if tryStmt.Var != "" {
				scope.Del(tryStmt.Var)
			}
		}
	}()

	throwNotHandled := false
	var throwObj Object = NIL
	if rv.Type() == THROW_OBJ { //如果try块返回的是一个throw对象或者是Error对象
		if tryStmt.Catch != nil { //如果有'catch'语句
			if rv.Type() == THROW_OBJ {
				throwObj = rv.(*Throw)
				if tryStmt.Var != "" { //'cathe e'
				 	//将throw对象放入这个key为'e'的scope中
					scope.Set(tryStmt.Var, rv.(*Throw).value)
				}
			} else { //Error对象
				throwObj = rv
				if tryStmt.Var != "" {
					scope.Set(tryStmt.Var, rv)
				}
			}
			rv = evalBlockStatement(tryStmt.Catch, catchScope) //解释catch块
			if rv.Type() == ERROR_OBJ {
				return rv
			}
		} else {
			throwNotHandled = true //没有'catch'块，则说明没有处理throw
		}
	}

	if tryStmt.Finally != nil { //解释'finally'块（如果有）
		rv = evalBlockStatement(tryStmt.Finally, scope)
		if (rv.Type() == ERROR_OBJ) {
			return rv
		}
	}

	if throwNotHandled { //如果没有处理throw，则继续抛出throw对象
		return throwObj
	}

	return rv
}
```

`evalThrowStatement`函数仅仅生成一个`Throw对象`返回。`evalTryStatement`函数稍微复杂一点，它先解释`try`的语句块（block），如果语句块中返回的是一个`throw对象`，说明有抛出的异常。然后判断是否有`catch`语句，如果有的话，就处理`catch块`。如果没有的话，说明异常没有被处理，我们会在66行的判断语句中继续抛出这个异常，最后是处理`finally块`。

上面说过，解释`try语句块(Block)`的时候，它可能会返回一个`throw`对象。但是我们还没有处理这种情况，因此需要修改`evalBlockStatement`函数：

```go
//eval.go
func evalBlockStatement(block *ast.BlockStatement, scope *Scope) Object {
	var result Object
	for _, statement := range block.Statements {
		result = Eval(statement, scope)
		if result != nil {
			rt := result.Type()
			if rt == RETURN_VALUE_OBJ || rt == ERROR_OBJ || rt == THROW_OBJ {
				return result
			}
		}
		//...
	}
	return result
}
```

我们仅仅修改了代码的第8行的判断，加入了`rt == THROW_OBJ`的条件。就是说如果我们在解释块语句的时候，遇到了`throw对象`的话，就不再继续处理，直接返回这个`throw对象`。

还有一点，如果`try语句`中抛出了异常，而却没有`catch`语句，那么我们的解释器 必须捕获这种情况。因此我们需要更改`evalProgram`函数，来捕获这种没有处理的异常：

```c
//eval.go
func evalProgram(program *ast.Program, scope *Scope) (results Object) {
	//...
	for _, stmt := range program.Statements {
		results = Eval(stmt, scope)
		//...
		if throwObj, ok := results.(*Throw); ok { //如果是trhow对象
			return newError(throwObj.stmt.Pos().Sline(), ERR_THROWNOTHANDLED, 
                            throwObj.value.Inspect())
		}
	}
	//...
}
```

第7-10行是新增的判断，如果`evalProgram`函数捕获到了`throw对象`，则说明这个`throw对象`没有被处理，我们就返回运行期错误。`ERR_THROWNOTHANDLED`变量是新增的，它定义在`errors.go`文件中：

```go
//errors.go
var (
	//...
	ERR_THROWNOTHANDLED = "throw object '%s' not handled"
)
```



最后我们再看一下最开始的例子：

```javascript
try {
    let th = 1 + 2
    println(th)
    throw 10
    println("After throw")
}
catch e {
    if type(e) == "string" {
        printf("Catched, e=%s\n", e)
    } else if type(e) == "number" {
        printf("Catched, e=%d\n", e)
    }
}
```

8-12行的代码，我们使用了`type`这个内置函数用来判断对象的类型，它返回一个字符串对象。这个`type`内置函数我们还没有实现。来看一下实现代码：

```go
//builtin.go
func init() {
	builtins = map[string]*Builtin{
		//...
		"type":    typeBuiltin(),
	}
}

func typeBuiltin() *Builtin {
	return &Builtin{
		Fn: func(line string, scope *Scope, args ...Object) Object {
			if len(args) != 1 {
				return newError(line, ERR_ARGUMENT, 1, len(args))
			}

			switch args[0].(type) {
			case *Number:
				return NewString("number")
			case *Nil:
				return NewString("nil")
			case *Boolean:
				return NewString("bool")
			case *Error:
				return NewString("error")
			case *Break:
				return NewString("break")
			case *Continue:
				return NewString("continue")
			case *ReturnValue:
				return NewString("return")
			case *Function:
				return NewString("function")
			case *Builtin:
				return NewString("builtin")
			case *RegEx:
				return NewString("regex")
			case *GoObject:
				return NewString("go")
			case *GoFuncObject:
				return NewString("gofunction")
			case *FileObject:
				return NewString("file")
			case *Os:
				return NewString("os")
			case *Struct:
				return NewString("struct")
			case *Throw:
				return NewString("throw")
			case *String:
				return NewString("string")
			case *Array:
				return NewString("array")
			case *Tuple:
				return NewString("tuple")
			case *Hash:
				return NewString("hash")
			default:
				return newError(line, "argument to `type` not supported, got=%s",
                                args[0].Type())
			}
		},
	}
}
```

最后，让我们来测试一下。

## 测试

```javascript
# try.mp
try { 
    try {
		println("inner try")
        throw "inner catch error"
		println("after try")
    } finally {
        println("finally")
    } 
} catch ex { 
    println(ex)
    try {
		println("try in catch")
        throw [1,2,3]
    } catch ex {
        printf("ex=%s\n", ex)
    }
}
```

运行结果：

```
inner try
finally
inner catch error
try in catch
ex=[1, 2, 3]
```



再来看一个没有`catch`的例子：

```javascript
# try2.mp
try {
    let th = 1 + 2
    println(th)
    throw "Hello"
    println("After throw")
}
#catch e {
#    printf("Catched, e=%s\n", e)
#}
finally {
    println("Finally running")
}

println("After try")
```

运行结果如下：

```
3
Finally running
Runtime Error at <examples/try.mp:5>
        throw object 'Hello' not handled
```



下一节，我们将介绍匿名函数（有时也称为Lambda函数）。




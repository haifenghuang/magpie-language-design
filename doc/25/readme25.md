# 多重赋值 & 多重返回值

我们之前介绍的`赋值表达式`仅支持单个变量赋值，同样的`return语句`也仅支持一个返回值，这一节我们将改变这一限制。我们将加入多重赋值和多重返回值的支持。看一下例子：

```javascript
a, b, c = 10, "hello", [12,3,"world"]
a, _, c = 10, "hello", [12,3,"world"] //_:表示废弃(或者忽略)这个值

fn calc(x,y) {
	return x+y, x-y    
}
add_result, sub_result = calc(10,2)
```

来看看我们需要做哪些更改呢？

3. 在抽象语法树（AST）的源码`ast.go`中修改`let`语句和`return`语句的抽象语法表示。
4. 在语法解析器（Parser）的源码`parser.go`中加入修改`let`语句和`return`语句的语法解释。
5. 在对象（Object）系统的源码`object.go`中修改`Return对象`的内容。
6. 在解释器（Evaluator）的源码`eval.go`中修改`let`语句和`return`语句的解释。



## 抽象语法树（AST）的更改

对于多重赋值，类似下面这样：

```javascript
a, b, c = 10, "hello", [12,3,"world"]
```

我们需要修改`let`语句的抽象语法表示。细心的读者可能要问了，为啥不是修改`赋值表达式`的抽象语法表示呢？

因为我们在代码中使用了个小技巧，使多重赋值变成了一个`let`语句。先来看一下更改后的`let`语句的抽象语法表示：

```go
//ast.go
//let <identifier1>,<identifier2>,... = <expression1>,<expression2>,...
type LetStatement struct {
	Token  token.Token
	Names  []*Identifier //变量现在变成了数组
	Values []Expression  //值也变成了一个数组
}

func (ls *LetStatement) Pos() token.Position {
	return ls.Token.Pos
}

func (ls *LetStatement) End() token.Position {
	aLen := len(ls.Values)
	if aLen > 0 {
		return ls.Values[aLen-1].End()
	}

	return ls.Names[0].End()
}

func (ls *LetStatement) statementNode()       {}
func (ls *LetStatement) TokenLiteral() string { return ls.Token.Literal }
func (ls *LetStatement) String() string {
	var out bytes.Buffer

	out.WriteString(ls.TokenLiteral() + " ")

    //名字部分
	names := []string{}
	for _, name := range ls.Names {
		names = append(names, name.String())
	}
	out.WriteString(strings.Join(names, ", "))

	if len(ls.Values) == 0 { //e.g. 'let x'
		out.WriteString(";")
		return out.String()
	}

	out.WriteString(" = ")

    //值部分
	values := []string{}
	for _, value := range ls.Values {
		values = append(values, value.String())
	}
	out.WriteString(strings.Join(values, ", "))

	return out.String()
}

```

第5行和第6行，我们分别将名字和值更改成了数组。其它部分的代码也做了相应的更改。

接下来，让我们看看对于`return`语句的更改：

```go
//ast.go
type ReturnStatement struct {
	Token        token.Token // 'return'词元
	ReturnValue  Expression  //为了向后兼容
	ReturnValues []Expression //新加入的字段（存储多个返回值）
}

func (rs *ReturnStatement) Pos() token.Position {
	return rs.Token.Pos
}

func (rs *ReturnStatement) End() token.Position {
	aLen := len(rs.ReturnValues)
	if aLen > 0 {
		return rs.ReturnValues[aLen-1].End()
	}

	return token.Position{Filename: rs.Token.Pos.Filename, Line: rs.Token.Pos.Line, 
                          Col: rs.Token.Pos.Col + len(rs.Token.Literal)}

}

func (rs *ReturnStatement) statementNode()       {}
func (rs *ReturnStatement) TokenLiteral() string { return rs.Token.Literal }
func (rs *ReturnStatement) String() string {
	var out bytes.Buffer

	out.WriteString(rs.TokenLiteral() + " ")

	//	if rs.ReturnValue != nil {
	//		out.WriteString(rs.ReturnValue.String())
	//	}

	values := []string{}
	for _, value := range rs.ReturnValues {
		values = append(values, value.String())
	}
	out.WriteString(strings.Join(values, ", "))

	out.WriteString(";")

	return out.String()
}
```

第5行，我们给`return语句`的结构增加了`ReturnValues`数组，但是第4行，还是保留了这个`ReturnValue`字段，主要是为了向下兼容，不影响之前的代码。



## 语法解析器（Parser）的更改

先来看一下对于`return`语句的更改，这个比较简单一些，所以我放在前面来讲：

```go
//parser.go
func (p *Parser) parseReturnStatement() *ast.ReturnStatement {
	//创建一个`ReturnStatement`结构
	stmt := &ast.ReturnStatement{Token: p.curToken, ReturnValues: []ast.Expression{}}

	//如果后面是跟着分号的话，就说明没有返回值
	if p.peekTokenIs(token.TOKEN_SEMICOLON) { //e.g.{ return; }
		p.nextToken()
		return stmt
	}

	//如果后面跟着右花括弧，同样说明没有返回值
	if p.peekTokenIs(token.TOKEN_RBRACE) { //e.g. { return }
		return stmt
	}

    //循环处理多个返回值
	p.nextToken()
	for {
		v := p.parseExpressionStatement().Expression
		stmt.ReturnValues = append(stmt.ReturnValues, v)

		if !p.peekTokenIs(token.TOKEN_COMMA) {
			break
		}
		p.nextToken()
		p.nextToken()
	}

    //如果返回值个数'>0'的话，为了向下兼容，我们需要给'ReturnValue'赋值
	if len(stmt.ReturnValues) > 0 {
		stmt.ReturnValue = stmt.ReturnValues[0]
	}
	return stmt
}
```

代码看起来也比较简单。

下面让我们看一下多重赋值的情况，多重赋值可以有两种形式：

```javascript
let a, b, c = 1, 2, 3
a, b, c = 1, 2, 3
```

第一种是有`let`关键字的，第二种是没有关键字的。对于第一种情况，我们只需要更改`parseLetStatement`方法，对于第二种，我们就需要更改`parseStatement`方法：

```go
//parser.go
func (p *Parser) parseStatement() ast.Statement {
	switch p.curToken.Type {
	case token.TOKEN_LET:
		return p.parseLetStatement(true)
	//...
	case token.TOKEN_IDENTIFIER:
		if p.peekTokenIs(token.TOKEN_COMMA) { //如果标识符后跟着','的话，则认为是多重赋值
			return p.parseLetStatement(false)
		}
		fallthrough
	default:
		return p.parseExpressionStatement()
	}
}
```

第7行代码判断我们遇到的是一个标识符（Identifier），我们还需要再判断下一个词元类型（Token Type）是不是`TOKEN_COMMA(逗号)`，如果是的话，我们就认为是一个多重赋值，否则的话，我们就`fallthrough`，就是和原来一样，让程序继续走第13行的`parseExpressionStatement`方法。

现在来看一下`parseLetStatement`方法的实现：

```go
//parser.go
// let a, b, c = 1, 2, 3   ---> nextFlag = true
// a, b, c = 1, 2, 3       ---> nextFlag = false
func (p *Parser) parseLetStatement(nextFlag bool) *ast.LetStatement {
	var stmt *ast.LetStatement
  
    var tok token.Token
    //如果是'a,b,c=1,2,3'这种情况，因为没有let关键字，我们手动生成一个`let`词元（Token）
	if !nextFlag {
		tok = token.Token{Pos: p.curToken.Pos, Type: token.TOKEN_LET, Literal: "let"}
	} else {
		tok = p.curToken
	}
	stmt = &ast.LetStatement{Token: tok}

	//names部分
	for {
		if nextFlag {
			p.nextToken()
		}
		nextFlag = true

        //当前词元必须是标识符或者`_`占位符，不是的话报错
		if !p.curTokenIs(token.TOKEN_IDENTIFIER) && p.curToken.Literal != "_" {
            //报错，为简便起见内容略
		}
		name := &ast.Identifier{Token: p.curToken, Value: p.curToken.Literal}
		stmt.Names = append(stmt.Names, name)

		p.nextToken()
        //遇到'='号或者';'则退出循环
		if p.curTokenIs(token.TOKEN_ASSIGN) || p.curTokenIs(token.TOKEN_SEMICOLON) {
			break
		}

        //如果标识符或者占位符后不是','号则报错
		if !p.curTokenIs(token.TOKEN_COMMA) {
			//报错，为简便起见内容略
		}
	}

	if p.curTokenIs(token.TOKEN_SEMICOLON) { //let x;
		return stmt
	}

	//'x, y, z = ;'的情况，则报错
	if p.curTokenIs(token.TOKEN_ASSIGN) && p.peekTokenIs(token.TOKEN_SEMICOLON) {
		//报错，为简便起见内容略
	}

	//values部分
	p.nextToken()
	for {
		v := p.parseExpressionStatement().Expression
		stmt.Values = append(stmt.Values, v)

		if !p.peekTokenIs(token.TOKEN_COMMA) {
			break
		}
		p.nextToken()
		p.nextToken()
	}

	return stmt
}
```

我在注释中将实现的一些细节写的比较详细，读者应该能够看懂。需要说明的一点就是18-21行：

```go
if nextFlag {
	p.nextToken()
}
nextFlag = true
```

有的读者可能不太理解，上面的`if`语句如果成立了，下面再将这个`nextFlag`设置为`true`是不是多余的？

当然不是。假设我们给`parseLetStatement`传入的`nextFlag`为`false`的话，第一个`if`判断就不会走，然后我们将`nextFlag`设置为`true`。因为这段语句是在`for`循环中，所以以后每次都会走这个`if`分支，即走这个`p.nextToken`语句。哦，你会说太复杂了。好吧，简单点说。对于有`let`的情况：

```javascript
let a, b, c = 1,2,3
```

每次都会走`p.nextToken()`语句。对于没有`let`的情况：

```javascript
a,b, c = 1,2,3
```

只有第一次进入`for`循环不会走`p.nextToken()`，以后每次都会走。

> 注意：细心的读者可能已经发现，这里的多重赋值是不支持下面几种形式的：
>
> ```go
> arr=[1,2,3]
> arr[0], arr[1] = arr[1], arr[0]
> 
> x， arr[1] = 10, 20
> ```
>
> 从解析代码中，就可以看出来，我们只支持左边的变量全部是标识符(Identifier)类型的场合。这个其实也很好理解，因为`let`语句是变量声明（声明的同时还可以赋初值）。



## 对象（Object）系统的更改

由于我们对`return`语句增加了返回多个值的功能，所以之前定义的`return对象`就需要加入这个变更：

```go
//object.go
type ReturnValue struct {
	Value  Object   // 为了向下兼容，所以保留
	Values []Object // 存储返回的多个值
}

func (rv *ReturnValue) Type() ObjectType { return RETURN_VALUE_OBJ }
func (rv *ReturnValue) Inspect() string {
	//return rv.Value.Inspect()

	var out bytes.Buffer
	values := []string{}
	for _, v := range rv.Values {
		values = append(values, v.Inspect())
	}

	out.WriteString(strings.Join(values, ", "))

	return out.String()
}
```

第4行是新增的代码。同时我们也更改了`Inspect`方法。



## 解释器（Evaluator）的更改

由于`LetStatement`和`ReturnStatement`都进行了更改，所以我们需要更改其相应的解释代码：

```go
//eval.go
func Eval(node ast.Node, scope *Scope) (val Object) {
	switch node := node.(type) {
	//...
	case *ast.LetStatement:
		return evalLetStatement(node, scope)
	case *ast.ReturnStatement:
		return evalReturnStatement(node, scope)
	//...
	}

	return nil
}
```

5-8行是新修改的代码。我们新建了`evalLetStatement`和`evalReturnStatement`两个函数。

先来看比较简单的`evalReturnStatement`函数：

```go
//eval.go
func evalReturnStatement(r *ast.ReturnStatement, scope *Scope) Object {
	if r.ReturnValue == nil { //如果没有返回值，则默认返回一个`NIL`对象
		return &ReturnValue{Value: NIL, Values: []Object{NIL}}
	}

	ret := &ReturnValue{}
	for _, value := range r.ReturnValues {
		ret.Values = append(ret.Values, Eval(value, scope))
	}

    // 为了向下兼容(for old campatibility)
	ret.Value = ret.Values[0]

	return ret
}
```

还有一个`unwrapReturnValue`函数需要修改，更改前的代码如下：

```go
//eval.go
func unwrapReturnValue(obj Object) Object {
	if returnValue, ok := obj.(*ReturnValue); ok {
		return returnValue.Value
	}

	return obj
}
```

对于多返回值，我们这里将会把多个返回值组装成一个`元组(Tuple)对象`然后返回：

```go
//eval.go
func unwrapReturnValue(obj Object) Object {
	if returnValue, ok := obj.(*ReturnValue); ok {
		if len(returnValue.Values) > 1 { //如果返回多个值
			//注意：这里的'IsMulti'为true，表明是函数的返回值，且有多个
			return &Tuple{Members: returnValue.Values, IsMulti: true}
		}
		return returnValue.Value
	}

	return obj
}
```



接下来看一下`evalLetStatement`函数的实现代码：

```go
//eval.go
func evalLetStatement(l *ast.LetStatement, scope *Scope) (val Object) {
	values := []Object{}
	valuesLen := 0
	for _, value := range l.Values { //循环遍历'Values数组'
		val := Eval(value, scope)
		if val.Type() == TUPLE_OBJ {
			tupleObj := val.(*Tuple)
			//如果值是一个元组且`IsMulti`为true的话，说明是函数的多个返回值
			if tupleObj.IsMulti {
				valuesLen += len(tupleObj.Members)
				values = append(values, tupleObj.Members...)

			} else { //否则就是一个元组
				valuesLen += 1
				values = append(values, tupleObj)
			}

		} else { //非元组的情况
			valuesLen += 1
			values = append(values, val)
		}
	}

    //循环遍历'Names'数组
	for idx, item := range l.Names {
        if idx >= valuesLen { //如果Names比Values多， 例如: 'let a, b, c = 1, 2'， 则我们会给'c'赋为NIL
			//如果不是`_`占位符，那么我们固定返回NIL值。
			if item.Token.Literal != "_" {
				val = NIL
				scope.Set(item.String(), val)
			}
		} else {
			//如果是`_`占位符，就继续
			if item.Token.Literal == "_" {
				continue
			}
			//取出相应的值，放入Scope
			val = values[idx]
			if val.Type() != ERROR_OBJ {
				scope.Set(item.String(), val)
			} else {
				return
			}
		}
	}

	return
}
```

3-23行处理`Let语句`的`Values`。这里需要说明的是如果Values中有元组的话，那么我们需要判断是否是函数的多个返回值。举个例子：

```javascript
let x, y = (12,34), "hello"

fn math(x,y) {
    return x+y, x-y
}
let x, y = math(3,2)
```

对于第1行的`let语句`，我们希望`(12,34)`这个元组作为一个整体，因此`x和y`结果如下：

```javascript
x = (12,34)
y = "hello"
```

对于第6行的`let语句`，因为是`math函数`的多个返回值，因此我们希望将它的返回值`（元组）`分开处理。我们的`x和y`结果如下：

```javascript
x = 5
y = 1
```

如果不这么处理的话，那么结果就是：

```javascript
x = (5, 1)
y = nil
```

这可能并不是我们期望的。



## 测试

```go
//main.go

func TestEval() {
	tests := []struct {
		input    string
		expected string
	}{
		{`a,b,c=1,true,"hello"; println(a) println(b) println(c)`, "nil"},
		{`a,b,c=2,false,["x","y","z"]; println(a) println(b) println(c[1])`, "nil"},
        {`a,b,c=2,(3,4); println(a) println(b) println(c)`, "nil"},
		{`fn math(x,y){return x+y,x-y} a,s=math(5,3) println(a) println(s)`, "nil"},
		{`fn xx(x, y){ return x+y, x-y, x*y} a,_,c=xx(5,3) println(a) println(c)`, "nil"},
	}

	for _, tt := range tests {
		l := lexer.NewLexer(tt.input)
		p := parser.NewParser(l)
		program := p.ParseProgram()
		if len(p.Errors()) != 0 {
			for _, err := range p.Errors() {
				fmt.Println(err)
			}
			break
		}

		scope := eval.NewScope(nil, os.Stdout)
		evaluated := eval.Eval(program, scope)
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



下一节，我们将加入对内置函数`printf`的支持。


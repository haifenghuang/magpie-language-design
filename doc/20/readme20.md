# 元祖（Tuple）支持

在之前的文章中，我们提供了对`数组`的支持。这一节，我们将提供对`元祖(tuple)`的支持。元祖和数组非常相像，区别是对于元祖，你无法给索引位置的元素赋值，也没法增加、删除元祖的元素。看一下使用元祖的例子：

```go
tup = () //空元祖
tup = (1,) //有一个元素的元祖。注意这里不能写成(1)，这是数字1，只是外面加了个括号。
tup = ("hello", true, 12.3) //有三个元素的元祖
println(tup[1]) //取元祖的第二个元素的值
```

从上面可以看出，由于我们没有引入任何新的操作符或者关键字，所以就没有新的词元类型，词法解析器（Lexer）也不用更改。

下面看一下我们需要做的更改：

1. 在抽象语法树（AST）的源码`ast.go`中加入`元祖`对应的抽象语法表示。
4. 在语法解析器（Parser）的源码`parser.go`中加入对`元祖`的语法解析。
4. 在对象（Object）系统的源码`object.go`中加入新的`元祖对象(Tuple Object)`。
5. 在解释器（Evaluator）的源码`eval.go`中加入对`元祖`的解释。



## 抽象语法树（AST）的更改

`元祖`是由一系列的表达式组成的，类似下面的形式：

```go
tup = (<expression1>,<expression2>, ...)
```

我们再来看一下`数组`，它也是由一系列的表达式组成的：

```go
arr = [<expression1>,<expression2>, ...]
```

除了使用的界定符不同外，其它都是一样的。

因此`元祖`的抽象语法树表示和`数组`的抽象语法树表示几乎是一样的：

```go
//ast.go
type TupleLiteral struct { //元祖字面量
	Token       token.Token
	Members     []Expression //成员数组
}

func (t *TupleLiteral) Pos() token.Position {
	return t.Token.Pos
}

func (t *TupleLiteral) End() token.Position {
	tlen := len(t.Members)
	if tLen > 0 {
		return t.Members[tLen-1].End()
	}
	return t.Token.Pos
}

func (t *TupleLiteral) expressionNode()      {}
func (t *TupleLiteral) TokenLiteral() string { return t.Token.Literal }
func (t *TupleLiteral) String() string {
	var out bytes.Buffer

	out.WriteString("(")

	members := []string{}
	for _, m := range t.Members {
		members = append(members, m.String())
	}

	out.WriteString(strings.Join(members, ", "))
	out.WriteString(")")

	return out.String()
}
```

这个`TupleLiteral`几乎就是从`ArrayLiteral`拷贝过来，然后改了一下名称而已。



## 语法解析器（Parser）的更改

由于`元祖`的界定符是`()`，所以我们要在原来解析括号的地方，来更改我们的代码逻辑，原来代码解析括号的函数是`parseGroupedExpression`，因此我们需要修改这个函数的逻辑，来看一下代码：

```go
//parser.go
func (p *Parser) parseGroupedExpression() ast.Expression {
    savedToken := p.curToken //保存'('处的词元（Token）
	p.nextToken() //取下一个词元

    //如果前一个词元类型为TOKEN_LPAREN, 当前词元类型为TOKEN_RPAREN, 即`()`，说明这是一个空的元祖
    //我们需要提前返回。
	if savedToken.Type == token.TOKEN_LPAREN && p.curTokenIs(token.TOKEN_RPAREN) {
		//空元祖'()'
		return &ast.TupleLiteral{Token: savedToken, Members: []ast.Expression{}}
	}

	exp := p.parseExpression(LOWEST) //解析第一个表达式

	if p.peekTokenIs(token.TOKEN_COMMA) {//如果下一个词元是类型是TOKEN_COMMA（即逗号）
		p.nextToken()
		ret := p.parseTupleExpression(savedToken, exp)
		return ret
	}

	if !p.expectPeek(token.TOKEN_RPAREN) {
		return nil
	}

	return exp
}

//解析元祖表达式
func (p *Parser) parseTupleExpression(tok token.Token, expr ast.Expression) ast.Expression {
	members := []ast.Expression{expr}

	oldToken := tok
	for {
		switch p.curToken.Type {
		case token.TOKEN_RPAREN: //右括号，说明遇到了元祖结束界定符
			ret := &ast.TupleLiteral{Token: tok, Members: members}
			return ret
		case token.TOKEN_COMMA: //逗号
			p.nextToken()
			//对于仅有一个元素的元祖，比如(1,)后面的逗号是必须的，以此来区分用括号括起来的'(1)'这种表达式。即
			//（1,)是个元祖
			//(1)是数字1，只不过用括号括起来了
			if p.curTokenIs(token.TOKEN_RPAREN) { //逗号后面跟了一个右括号
				ret := &ast.TupleLiteral{Token: tok, Members: members}
				return ret
			}
			members = append(members, p.parseExpression(LOWEST))
			oldToken = p.curToken
			p.nextToken()
		default:
			oldToken.Pos.Col = oldToken.Pos.Col + len(oldToken.Literal)
			msg := fmt.Sprintf("Syntax Error:%v- expected token to be ',' or ')', got %s instead",
								oldToken.Pos, p.curToken.Type)
			p.errors = append(p.errors, msg)
			p.errorLines = append(p.errorLines, oldToken.Pos.Sline())
			return nil
		}
	}
}
```

代码的注释中，我写了详细的说明，以便于读者理解。



## 对象系统的更改

我们需要往对象系统中增加一个`元祖对象(Tuple Object)`。代码几乎和`数组对象(Array Object)`差不多：

```go
//object.go

const (
	//...
	TUPLE_OBJ       = "TUPLE"
)

//元祖对象
type Tuple struct {
	// 这个`IsMulti`主要使用在将来我们要扩展的函数多个返回值上。
	// 如果一个函数返回多个值，那么我们会将多个返回值放入tuple，并将这个
	// 字段设置为true
	IsMulti bool
	Members []Object
}

func (t *Tuple) iter() bool { return true } //表示元祖是可以遍历的

func (t *Tuple) Inspect() string {
	var out bytes.Buffer
	members := []string{}
	for _, m := range t.Members {
		if m.Type() == STRING_OBJ {
			members = append(members, "\""+m.Inspect()+"\"")
		} else {
			members = append(members, m.Inspect())
		}
	}
	out.WriteString("(")
	out.WriteString(strings.Join(members, ", "))
    if (len(t.Members)) == 1 { //如果元祖中只有一个元素，我们需要在后面加一个','， 例如(1,)
		out.WriteString(",")
	}
	out.WriteString(")")

	return out.String()
}

func (t *Tuple) Type() ObjectType { return TUPLE_OBJ }

func (t *Tuple) CallMethod(line string, scope *Scope, method string, args ...Object) Object {
	switch method {
	case "get": //获取指定indx处的元祖值
		return t.get(line, args...)
	case "empty": //判断元祖是否为空
		return t.empty(line, args...)
	case "len": //计算元祖长度
		return t.len(line, args...)

	}
	return newError(line, ERR_NOMETHOD, method, t.Type())
}

func (t *Tuple) len(line string, args ...Object) Object {
	if len(args) != 0 {
		return newError(line, ERR_ARGUMENT, "0", len(args))
	}
	return NewNumber(float64(len(t.Members)))
}

func (t *Tuple) get(line string, args ...Object) Object {
	if len(args) != 1 {
		return newError(line, ERR_ARGUMENT, "1", len(args))
	}

	idxObj, ok := args[0].(*Number)
	if !ok {
		return newError(line, ERR_PARAMTYPE, "first", "get", "*Number", args[0].Type())
	}

	val := int64(idxObj.Value)
	if val < 0 || val >= int64(len(t.Members)) {
		return newError(line, ERR_INDEX, val)
	}
	return t.Members[val]
}

func (t *Tuple) empty(line string, args ...Object) Object {
	l := len(args)
	if l != 0 {
		return newError(line, ERR_ARGUMENT, "0", l)
	}

	if len(t.Members) == 0 {
		return TRUE
	}
	return FALSE
}

func (t *Tuple) HashKey() HashKey {
	// https://en.wikipedia.org/wiki/Jenkins_hash_function
	var hash uint64 = 0
	for _, v := range t.Members {
		hashable, ok := v.(Hashable)
		if !ok {
			errStr := fmt.Sprintf(ERR_KEY, v.Type())
			panic(errStr)
		}

		h := hashable.HashKey()

		hash += h.Value
		hash += hash << 10
		hash ^= hash >> 6
	}
	hash += hash << 3
	hash ^= hash >> 11
	hash += hash << 15

	return HashKey{Type: t.Type(), Value: hash}
}
```

在`CallMethod`方法中，我们没有提供类似数组的`push`、`pop`方法，因为元祖是不能更改元素数目的。

在前几节当中，我们说过，元组也可以用作`哈希(Hash)`的key，因此90-111行我们实现了`HashKey()`方法。这里使用的方法是`Jenkins`哈希算法。



## 解释器（Evaluator）的更改

对解释器的更改主要是以下几个方面：

1. `Eval()`函数需要增加一个处理`元祖表达式`的分支
2. __取元祖特定索引位置处的值__（更改`evalIndexExpression`函数）
3. 给元祖特定索引位置处的元素赋值（需要返回错误对象，我们不能对元祖的特定元素赋值）

现在让我们一个一个实现上面说的。



1.  给`Eval()`函数增加一个处理`元祖表达式`的分支

```go
func Eval(node ast.Node, scope *Scope) (val Object) {
	switch node := node.(type) {
	//...
	case *ast.ArrayLiteral:
		members := evalExpressions(node.Members, scope)
		if len(members) == 1 && isError(members[0]) {
			return members[0]
		}

		return &Array{Members: members}
	case *ast.TupleLiteral:
		members := evalExpressions(node.Members, scope)
		if len(members) == 1 && isError(members[0]) {
			return members[0]
		}

		return &Tuple{Members: members}
    }

	return nil
}
```

这个处理`元祖表达式`的分支（11-17行）的代码和处理`数组表达式`的分支（4-10行）几乎是一样的。



2. 取元祖特定索引位置处的值

```go
//eval.go

//处理索引表达式
func evalIndexExpression(node *ast.IndexExpression, left, index Object) Object {
	switch {
	case left.Type() == ARRAY_OBJ && index.Type() == NUMBER_OBJ:
		return evalArrayIndexExpression(node.Pos().Sline(), index)
	case left.Type() == TUPLE_OBJ:
		return evalTupleIndexExpression(node.Pos().Sline(), left, index)
	default:
		return newError(node.Pos().Sline(), ERR_NOINDEXABLE, left.Type())
	}
}

//处理元祖的索引表达式：tup[idx]
func evalTupleIndexExpression(line string, tuple, index Object) Object {
	tupleObject := tuple.(*Tuple)
	idx := int64(index.(*Number).Value)
	max := int64(len(tupleObject.Members) - 1)
	if idx < 0 || idx > max {
		//这里返回NIL。实际上可以根据情况返回错误：newError(xxx)
		//return newError(line, "index out of bound. index: %d, max: %d", idx, max)
		return NIL
	}

	return tupleObject.Members[idx]
}
```

代码的8-9行增加了一个`case`分支，用来处理元祖的索引表达式。16-27行是实际的处理`元祖索引表达式`的代码，几乎就是`evalArrayIndexExpression()`函数的克隆。

还有一个地方需要说明，我们的语言允许用户书写如下语句：

```javascript
tup = （1, "hello"）
if tup { //判断元祖长度
    println("tuple length is larger than zero")
}
```

因此，我们需要在`IsTrue`函数中加入相关的判断：

```go
//eval.go
func IsTrue(obj Object) bool {
	switch obj {
	//...
	default:
		switch obj.Type() {
		//...
		case ARRAY_OBJ:
			if len(obj.(*Array).Members) == 0 {
				return false
			}
		case TUPLE_OBJ:
			if len(obj.(*Tuple).Members) == 0 {
				return false
			}
		}
		return true
	}
}
```

代码的12-15行是新增的判断。没啥特殊的，几乎和数组对象的判断一样。



3.  给元祖特定索引位置处的元素赋值

上面也提到过，元祖的元素不能被赋值，所以这里我们直接返回一个错误对象：

```go
//eval.go
func evalAssignExpression(a *ast.AssignExpression, scope *Scope) Object {
	//...
	switch left.Type() {
	//...
	case ARRAY_OBJ:
		return evalArrayAssignExpression(a, name, left, scope, val)
	case TUPLE_OBJ:
		return evalTupleAssignExpression(a, name, left, scope, val)
	//...
	}
}

func evalTupleAssignExpression(a *ast.AssignExpression, name string, left Object, scope *Scope, val Object) (ret Object) {
	if a.Token.Literal == "=" { //tuple[idx] = item
		str := fmt.Sprintf("%s[IDX]", TUPLE_OBJ)
		return newError(a.Pos().Sline(), ERR_INFIXOP, str, a.Token.Literal, val.Type())
	}
	return newError(a.Pos().Sline(), ERR_INFIXOP, left.Type(), a.Token.Literal, val.Type())
}
```

代码9-11行增加了判断类型为`元祖`类型的分支，实际的逻辑在`evalTupleAssignExpression`函数中实现。在`evalTupleAssignExpression`函数中，我们仅仅返回一个错误对象。

最后还有一点，我们还需要给`len`内置函数加入取`Tuple`长度的分支：

```go
//builtin.go
func lenBuiltin() *Builtin {
	return &Builtin{
		Fn: func(line string, scope *Scope, args ...Object) Object {
			if len(args) != 1 {
				return newError(line, "wrong number of arguments. got %d, want=1", len(args))
			}

			switch arg := args[0].(type) {
			//...
			case *Tuple:
				return NewNumber(float64(len(arg.Members)))
			}
		},
	}
}
```

我们给`len`内置函数加入了取元祖长度的分支（代码11-12行）。



## 测试

下面我们写一个简单的程序测试一下：
```go
//main.go
func TestEval() {
	tests := []struct {
		input    string
		expected string
	}{
		{"let tup = (1, 10.5, \"Hello\", true); tup[0]", "1"},
		{"let tup = (1, 10.5, \"Hello\", true); tup[1]", "10.5"},
		{"let tup = (1, 10.5, \"Hello\", true); tup[2]", "Hello"},
		{"let tup = (1, 10.5, \"Hello\", true); tup[3]", "true"},
		{"let tup = (1, 10.5, \"Hello\", true); len(tup)", "4"},
		{"let tup = (1,); len(tup)", "1"},
		{"let tup = (); len(tup)", "0"},
		{`let tup = (1,); tup[0]=10`, "error"},
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

最后一个测试用例：

```javascript
let tup = (1,); tup[0]=10
```

会报告如下的错误：

```
Runtime Error at 1
        unsupported operator for infix expression: TUPLE[IDX] '=' NUMBER
```



下一节，我们会提供对`命名函数`的支持。

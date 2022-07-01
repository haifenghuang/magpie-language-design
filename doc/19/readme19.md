# `赋值表达式(=)`支持

`赋值（=）`应该是语言中最常见的一种语法。所以读者应该都是很熟悉了。

我们看一下形式：

```perl
a = 10
arr[5] = 12
hash["key"] = value
obj.xxx = 10
```

下面来看一下需要对代码做的更改：

1. 在抽象语法树（AST）的源码`ast.go`中加入`赋值表达式`对应的抽象语法表示。
3. 在语法解析器（Parser）的源码`parser.go`中加入对`赋值表达式`的语法解析。
4. 在解释器（Evaluator）的源码`eval.go`中加入对`赋值表达式`的解释。



## 抽象语法树（AST）的更改

`赋值表达式`的抽象语法表示比较简单，因为赋值的形式比较简单：

```perl
name = value
```

下面来看一下`赋值表达式`的代码：

```go
//ast.go
//赋值表达式： Name = Value
type AssignExpression struct {
	Token token.Token
	Name  Expression
	Value Expression
}

func (ae *AssignExpression) Pos() token.Position {
	return ae.Name.Pos()
}

func (ae *AssignExpression) End() token.Position {
	return ae.Value.End()
}

func (ae *AssignExpression) expressionNode()      {}
func (ae *AssignExpression) TokenLiteral() string { return ae.Token.Literal }

func (ae *AssignExpression) String() string {
	var out bytes.Buffer

	out.WriteString(ae.Name.String())
	out.WriteString(ae.Token.Literal)
	out.WriteString(ae.Value.String())

	return out.String()
}
```



## 语法解析器（Parser）的更改

对于语法解析器（Parser），我们需要做三处更改：

1. 对赋值词元类型（TOKEN_ASSIGN）注册中缀表达式回调函数
2. 新增解析`赋值表达式`的函数
2. 对赋值操作符`=`赋予优先级

我们先来看看赋值中缀操作符优先级。举个例子：

```go
x = a + 3
x = a * 3
x = arr[10]
arr[10] = 2
x = add(1,2)
obj.xxx = 12
x = 6 > 3 && 3 > 1
//...
```

所有的这些式子，赋值操作符的优先级都应该是最低的。也就是说我们需要计算完右边的`表达式`，然后再赋值。看一下代码：

```go
//parser.go
const (
	_ int = iota
	LOWEST
	ASSIGN      //=
	//...
)

var precedences = map[token.TokenType]int{
	token.TOKEN_ASSIGN: ASSIGN,
	//...
}
```

接下来，我们需要给`赋值操作符（=）`注册中缀回调函数：

```go
//parser.go
func (p *Parser) registerAction() {
	//...
	p.registerInfix(token.TOKEN_ASSIGN, p.parseAssignExpression)
}

func (p *Parser) parseAssignExpression(name ast.Expression) ast.Expression {
	a := &ast.AssignExpression{Token: p.curToken, Name: name}

	p.nextToken()
	a.Value = p.parseExpression(LOWEST) //处理value

	return a
}
```



## 解释器（Evaluator）的更改

我们需要在解释器（Evaluator）的`Eval`函数的`switch`分支中加入对`赋值表达式`的处理：

```go
//eval.go
func Eval(node ast.Node) (val Object) {
	switch node := node.(type) {
	//...
	case *ast.AssignExpression:
		return evalAssignExpression(node, scope)
	//...
	}
	return nil
}

//处理赋值表达式
func evalAssignExpression(a *ast.AssignExpression, scope *Scope) Object {
	val := Eval(a.Value, scope) //解释赋值语句右边的表达式
	if val.Type() == ERROR_OBJ {
		return val
	}

	var name string
	switch nodeType := a.Name.(type) {
	case *ast.Identifier: 	//a = 10
		name = nodeType.Value
	case *ast.IndexExpression:	//arr[idx] = "xxx"的情形, 这里a.Name等于'arr[idx]''
		switch nodeType.Left.(type) {
		case *ast.Identifier:
			name = nodeType.Left.(*ast.Identifier).Value //这里, name等于'arr'
		}
	case *ast.MethodCallExpression: //obj.xxx = 20
		name = nodeType.Object.String()
	}

	//这里读者可能有些奇怪，这个函数是处理赋值表达式的，为啥还要判断【a.Token.Literal == "="】这个条件呢？
	//是不是多此一举，其实赋值表达并不是只有"="，还有"+=", "-=", "*="。。。，这些之后我们会实现
	if a.Token.Literal == "=" {
		switch nodeType := a.Name.(type) {
		case *ast.Identifier: //e.g. name = "hello"
			scope.Set(nodeType.Value, val) //将值放入scope中
			return val
		}
	}

	//测试变量是否存在
	//到这里表明不是'name = value'这种简单的赋值，因为这个在上面的if语句中已经处理过了
	//这里的情形类似下面这几种：
    //  number += number (操作符不是'='的情况)
	//  obj.xxx = yyy
	//  arr[index] = xxx
	//  hash[key] = xxx
	//  str[index] = "xxx"
	var left Object
	var ok bool
	if left, ok = scope.Get(name); !ok { //不存在就报错
		return newError(a.Pos().Sline(), ERR_UNKNOWNIDENT, name)
	}

	switch left.Type() { //判断左边对象的类型
	case NUMBER_OBJ:
		return evalNumAssignExpression(a, name, left, scope, val)
	case STRING_OBJ:
		return evalStrAssignExpression(a, name, left, scope, val)
	case ARRAY_OBJ:
		return evalArrayAssignExpression(a, name, left, scope, val)
	case HASH_OBJ:
		return evalHashAssignExpression(a, name, left, scope, val)
	}

	return newError(a.Pos().Sline(), ERR_INFIXOP, left.Type(), a.Token.Literal, val.Type())
}

//这里没有实现。具体内容，之后介绍
// num += num
// num -= num
// etc...
func evalNumAssignExpression(a *ast.AssignExpression, name string, left Object, scope *Scope, val Object) (ret Object) {
	switch a.Token.Literal {
	case "+=":
	case "-=":
	}
	return newError(a.Pos().Sline(), ERR_INFIXOP, left.Type(), a.Token.Literal, val.Type())
}

//str[idx] = item
//str += item  ---> 没有实现。具体内容，之后介绍
func evalStrAssignExpression(a *ast.AssignExpression, name string, left Object, scope *Scope, val Object) (ret Object) {
	leftVal := left.(*String).String

	switch a.Token.Literal {
	case "=":
		switch nodeType := a.Name.(type) {
		case *ast.IndexExpression: //str[idx] = xxx
			//解释索引
			index := Eval(nodeType.Index, scope)
			if index == NIL {
				ret = NIL
				return
			}

			//取出索引的值，这里直接将索引转换成数字对象,如果转换失败，go语言会panic
			//其实可以更优雅的处理这个问题，比如不是数字，则返回一个错误对象
			idx := int64(index.(*Number).Value)

			//判断索引是否越界
			if idx < 0 || idx >= int64(len(leftVal)) {
				return newError(a.Pos().Sline(), ERR_INDEX, idx)
			}

			//获取新的字符串,并写入Scope
			ret = NewString(leftVal[:idx] + val.Inspect() + leftVal[idx+1:])
			scope.Set(name, ret) 
			return
		}
	}

	ret = NewString(leftVal+val.Inspect())
	ret, ok = scope.Set(name, ret)
	return
}

//array[idx] = item
//array += item ---> 没有实现。具体内容，留给读者
func evalArrayAssignExpression(a *ast.AssignExpression, name string, left Object, scope *Scope, val Object) (ret Object) {
	leftVals := left.(*Array).Members

	switch a.Token.Literal {
	case "=":
		switch nodeType := a.Name.(type) {
		case *ast.IndexExpression: //arr[idx] = xxx
			index := Eval(nodeType.Index, scope)
			if index == NIL {
				ret = NIL
				return
			}

			idx := int64(index.(*Number).Value)
			//判断索引是否越界
			if idx < 0 {
				return newError(a.Pos().Sline(), ERR_INDEX, idx)
			}

			if idx < int64(len(leftVals)) { //没有越界
				//给特定索引处的元素赋值后，再写回Scope
				leftVals[idx] = val
				ret = &Array{Members: leftVals}
				scope.Set(name, ret)
				return
			} else { //索引越界, 则自动扩展数组
				for i := int64(len(leftVals)); i < idx; i++ {
					leftVals = append(leftVals, NIL)
				}

				leftVals = append(leftVals, val)
				ret = &Array{Members: leftVals}
				scope.Set(name, ret)
				return
			}
		}

		return newError(a.Pos().Sline(), ERR_INFIXOP, left.Type(), a.Token.Literal, val.Type())
	}

	return newError(a.Pos().Sline(), ERR_INFIXOP, left.Type(), a.Token.Literal, val.Type())
}

//hash[key] = value
func evalHashAssignExpression(a *ast.AssignExpression, name string, left Object, scope *Scope, val Object) (ret Object) {
	leftHash := left.(*Hash)

	switch a.Token.Literal {
	case "=":
		switch nodeType := a.Name.(type) {
		case *ast.IndexExpression: //hashObj[key] = val
			key := Eval(nodeType.Index, scope)
			leftHash.push(a.Pos().Sline(), key, val)
			return leftHash
		case *ast.Identifier: //hashObj.key = val
			key := strings.Split(a.Name.String(), ".")[1]
			keyObj := NewString(key)
			leftHash.push(a.Pos().Sline(), keyObj, val)
			return leftHash
		}
		return newError(a.Pos().Sline(), ERR_INFIXOP, left.Type(), a.Token.Literal, val.Type())
	}

	return newError(a.Pos().Sline(), ERR_INFIXOP, left.Type(), a.Token.Literal, val.Type())
}
```

内容虽然有点多，但是并不复杂，关键地方我都写了注释。



## 测试

```go
//main.go
func TestEval() {
	tests := []struct {
		input    string
		expected string
	}{
		{`a = "hello world"; a`, "hello world"},
		{`a = "hello world"; a[2]="w"; a`, "hewlo world"},
		{`arr=[1, "hello", true]; arr[0] = "good"; arr[0]`, "good"},
		{`myHash={}; myHash["name"]="huanghaifeng"; myHash["name"]`, "huanghaifeng"},
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



下一节，我们将增加对`元祖（Tuple）字面量`的支持。

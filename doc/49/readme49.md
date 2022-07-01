# `增强哈希（Enhanced Hash）`支持

这一节中，我们会增强对哈希对象的支持。我们知道，声明哈希后，可以使用索引方式来访问哈希元素：

```go
h = {"a": 1, "b": 2}
println(h["a"]) //打印1

h["a"] = 10
println(h["a"]) //打印10

h["x"] = fn(x,y) { return x + y }
println(h["x"](4,3)) //打印7

```

但是有时候，使用这种索引的方式访问哈希元素不是很方便。我们希望可以使用下面的方式来访问哈希元素：

```go
h = {"a": 1, "b": 2}
println(h.a) //打印1

h.a = 10
println(h.a) //打印10

h.x = fn(x,y) { return x + y }
println(h.x(4,3)) //打印7
```

即使用`.`的方式来访问哈希元素。

但是这里有一点需要说明：使用`.`的方式访问哈希，它会把哈希的key当做字符串对象，不支持key为布尔对象或者数字对象。

我们知道，先前的文章介绍哈希实现的时候，哈希的key可以是布尔对象或者数字对象：

```go
h1 = { true: 1, false: 0}
println(h1[true]) //打印1

h2 = { 5: "five", 10: "ten"}
println(h2[5]) //打印"five"
```

如果使用`.`的方式来访问的话，就变成了下面这样：

```go
h1 = { true: 1, false: 0}
println(h1.true)

h2 = { 5: "five", 10: "ten"}
println(h2.5)
```

第2行变成了`h1.true`，第4行变成了`h2.5`。这样做的话，会让人感觉非常困惑。因此我们只支持`.`后面跟标识符的情况：

```go
h = {}
h.key = 12
```

因为我们没有增加新的词元，而且我们的语法解析器能够正确的识别`.`。因此，这里我们只需要更改解释器。



## 解释器（Evaluator）的更改

在更改之前，让我们先给`Hash`对象增加一个`get`方法，用来根据指定的key，得到key中存储的相应值（value）。

```go
//object.go
func (h *Hash) CallMethod(line string, scope *Scope, method string, args ...Object) Object {
	switch method {
	//...
	case "get":
		return h.get(line, args...)
	}

	return newError(line, ERR_NOMETHOD, method, h.Type())
}

func (h *Hash) get(line string, args ...Object) Object {
	if len(args) != 1 {
		return newError(line, ERR_ARGUMENT, "1", len(args))
	}
	hashable, ok := args[0].(Hashable) //判断对象是否可哈希化
	if !ok {
		return newError(line, ERR_KEY, args[0].Type())
	}
	if hashPair, ok := h.Pairs[hashable.HashKey()]; ok { //取得pair
		return hashPair.Value
	}
	return NIL //没找到key，返回NIL对象
}
```

方法比较简单，所以不多做介绍了。接下来我们开始介绍真正需要更改的地方。

对于`hash.key`的这种方式，在语法解析（Parser）阶段，我们是将其当做方法调用（method call）来处理的，因此我们需要更改`evalMethodCallExpression`方法。

```go
//eval.go
func evalMethodCallExpression(call *ast.MethodCallExpression, scope *Scope) Object {
	//...

	obj := Eval(call.Object, scope)
	if obj.Type() == ERROR_OBJ {
		return obj
	}

	switch m := obj.(type) {
	//...
	//这里假设为'hs.xxx'或者hs.xxx(param1, param2, ...)
	case *Hash:
		switch o := call.Call.(type) { //判断'xxx'的类型
		case *ast.Identifier:
			//将'xxx'作为字符串，生成一个新的字符串对象，表示索引，模仿hs["xxx"]
			index := NewString(call.Call.String())
			//调用已知函数'evalHashIndexExpression'
			return evalHashIndexExpression(call.Call.Pos().Sline(), m, index)
		case *ast.CallExpression:
			//o.Function.String()即'xxx'
			funcObj := m.get(call.Call.Pos().Sline(), NewString(o.Function.String()))
			if isError(funcObj) {
				return funcObj
			}
			return evalCallExpression(o, funcObj, scope)
		}
	//...
	}

	return newError(call.Call.Pos().Sline(), ERR_NOMETHOD, call.String(), obj.Type())
}
```

代码还是很好理解的。需要注意的是第26行的`evalCallExpression`这个函数，我们给它增加了一个参数（即第二个参数）。我们从哈希中取得函数对象后，将它传给了`evalCallExpression`函数。`evalCallExpression`函数更改前的原型是这样的：

```go
//eval.go
func evalCallExpression(node *ast.CallExpression, scope *Scope) Object {
}
```

更改后，函数原型变成了下面这样：

```go
//eval.go
func evalCallExpression(node *ast.CallExpression, funcObj Object, scope *Scope) Object {
}
```

更改了这个函数原型后，`Eval`函数的`case *ast.CallExpression`分支调用的地方就需要做相应的更改：

```go
//eval.go
func Eval(node ast.Node, scope *Scope) (val Object) {
	case *ast.CallExpression:
		return evalCallExpression(node, nil, scope)
}
```

第4行，我们给`evalCallExpression`传递的第二个参数是`nil`。因为这里对`evalCallExpression`函数的调用，我们还没有解释`*ast.CallExpression`中的函数。

来看一下`evalCallExpression`的变动：

```go
//eval.go

func evalCallExpression(node *ast.CallExpression, funcObj Object, scope *Scope) Object {
	//...

	var function Object
	if funcObj != nil { //如果已经有函数对象了
		function = funcObj
	} else {
		function = Eval(node.Function, scope) //解释*ast.CallExpression中的Function
		if isError(function) {
			return function
		}
	}

	//...
	return applyFunction(node.Pos().Sline(), scope, function, args)
}
```

其中10-12行是原有的逻辑代码。7-9行的判断是新增的。

我们还需要修改赋值语句的逻辑，以支持`h.key=xxx`的形式：

```go
//eval.go
func _evalAssignExpression(a *ast.AssignExpression, val Object, scope *Scope) Object {
	//...
	if strings.Contains(a.Name.String(), ".") {
		switch o := a.Name.(type) {
		case *Hash: //h.a = xxx
			key := NewString(o.Call.String())
			m.push(a.Pos().Sline(), key, val) //调用哈希对象的'push'方法
			return NIL
		}
}
```



## 其它内置对象的增强支持

虽然本节主要讲解的是对于哈希的增强支持。但是这里也简单的提及一下对于字符串、数组和元组的相应增强。

看几个简单的例子：

```go
a = [1, "hello world"]  // 数组
println(a.1)            //a.1 等价于 a[1]
a.1 = "hello"
println(a.1)

t = (1, "hello world")  //元组
println(t.1)            // t.1 等价于 t[1]


s = "Hello World"  //字符串
println(s.6)       //s.6 等价于 s[6]
s.6 = "myz"
println(s.6)
println(s)

```

这里不会详细介绍，只是列出相应的代码，并在代码中做一些注释，以便读者理解。



下面是给数组和字符串对象增加给特定索引设置值的函数`set`:

```go
//object.go
func (s *String) CallMethod(line string, scope *Scope, method string, args ...Object) Object {
	switch method {
	//...
	case "set": //给String对象增加给特定索引设置值的函数
		return s.set(line, args...)
	}
	return newError(line, ERR_NOMETHOD, method, s.Type())
}

func (s *String) set(line string, args ...Object) Object {
	argLen := len(args)
	if argLen != 2 { //判断参数个数
		return newError(line, ERR_ARGUMENT, "2", argLen)
	}

	idxObj, ok := args[0].(*Number) //判断参数类型
	if !ok {
		return newError(line, ERR_PARAMTYPE, "first", "set", "*Number", args[0].Type())
	}

	idx := int64(idxObj.Value) //取出索引并判断范围
	if idx < 0 || idx > int64(len(s.String)) {
		return newError(line, ERR_INDEX, idx)
	}

	replaceObj, ok := args[1].(*String) //取出需要替换的值
	if !ok {
		return newError(line, ERR_PARAMTYPE, "second", "set", "*String", args[1].Type())
	}

	//实际的替换操作
	var out bytes.Buffer
	runes := []rune(s.String)
	for index, rune := range runes {
		if int64(index) == idx {
			out.WriteString(replaceObj.String)
		} else {
			out.WriteString(string(rune))
		}
	}

	s.String = out.String() //将变化后的值重新写回字符串对象
	return s
}
```



```go
//object.go
func (a *Array) CallMethod(line string, scope *Scope, method string, args ...Object) Object {
	switch method {
	//...
	case "set": //给数组对象增加给特定索引设置值的函数
		return a.set(line, args...)
	}
	return newError(line, ERR_NOMETHOD, method, a.Type())
}

func (a *Array) set(line string, args ...Object) Object {
	if len(args) != 2 { //判断参数个数
		return newError(line, ERR_ARGUMENT, "2", len(args))
	}

	idxObj, ok := args[0].(*Number) //判断参数类型
	if !ok {
		return newError(line, ERR_PARAMTYPE, "first", "set", "*Number", args[0].Type())
	}

	idx := int64(idxObj.Value) //取出索引
	if idx < 0 || idx >= int64(len(a.Members)) {
		oldLen := int64(len(a.Members))
		for i := oldLen; i <= idx; i++ {
			a.Members = append(a.Members, NIL)
		}
	}

	a.Members[idx] = args[1] //替换特定索引处的值
	return NIL
}
```



下面是赋值函数，例如`arr_obj.1 = xxx`、`str_obj.1 = xxx`。

```go
//eval.go
func _evalAssignExpression(a *ast.AssignExpression, val Object, scope *Scope) Object {
	//...
	if strings.Contains(a.Name.String(), ".") {
		switch o := a.Name.(type) {
		//...
		case *Array: //a.1 = xxx
			switch o.Call.(type) {
			case *ast.NumberLiteral:
				index := Eval(o.Call, scope) //解释索引
				m.set(o.Call.Pos().Sline(), index, val) //调用Array对象的'set'方法
			}
			return NIL
		case *String: //s.1 = xxx
			switch o.Call.(type) {
			case *ast.NumberLiteral:
				index := Eval(o.Call, scope) //解释索引
				m.set(o.Call.Pos().Sline(), index, val) //调用String对象的'set'方法
			}
			return NIL
		}
}
```



下面是取值函数，例如`println(arr_obj.1)`、`println(str_obj.1)`和`println(tup_obj.1)`。

```go
//eval.go
func evalMethodCallExpression(call *ast.MethodCallExpression, scope *Scope) Object {
	//...

	obj := Eval(call.Object, scope)
	if obj.Type() == ERROR_OBJ {
		return obj
	}

	switch m := obj.(type) {
	//...
	default:
		if obj.Type() == ARRAY_OBJ { //数组
			switch call.Call.(type) {
			case *ast.NumberLiteral:
				index := Eval(call.Call, scope)  //解释索引
				//调用既有函数evalArrayIndexExpression
				return evalArrayIndexExpression(call.Call.Pos().Sline(), obj, index)
			}
		} else if obj.Type() == TUPLE_OBJ { //元祖
			switch call.Call.(type) {
			case *ast.NumberLiteral:
				index := Eval(call.Call, scope)  //解释索引
				//调用既有函数evalTupleIndexExpression
				return evalTupleIndexExpression(call.Call.Pos().Sline(), m, index)
			}
		} else if obj.Type() == STRING_OBJ { //字符串
			switch call.Call.(type) {
			case *ast.NumberLiteral:
				index := Eval(call.Call, scope)  //解释索引
				//调用既有函数evalStringIndex
				return evalStringIndex(call.Call.Pos().Sline(), m, index)
			}
		}
		//...
	}

	return newError(call.Call.Pos().Sline(), ERR_NOMETHOD, call.String(), obj.Type())
}
```



## 测试

### 哈希

```javascript
fn demo() {
  h = {}
  h.a = 10
  h.b = 2
  h.c = fn(x,y) { return x + y }
  return h
}

hs = demo()
printf("hs.a=%d\n", hs.a)
println(hs.c(2,3))
println(hs.b)
```



### 数组、元组和字符串

```javascript
a = [1, "hello world"]           // 数组
println(a.1)
a.1 = "hello"
println(a.1)

t = (1, "hello world")           // 元组
println(t.1)
//t.1 = "hello" //错误


s = "Hello World"               //字符串
println(s.6)
s.6 = "myz"
println(s.6)
println(s)
```



下一节，我们将介绍Pipe操作符`|>`。




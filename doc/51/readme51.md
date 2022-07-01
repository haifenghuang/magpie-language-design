# `有序哈希（Ordered Hash）`支持

这一节中，我们会增加对`有序哈希`的支持。为什么要提供对`有序哈希`的支持呢？大家知道，当我们取哈希键值对、打印哈希、遍历哈希、或者取哈希的所有`key或者value`的时候，你每次看到的顺序可能都是不一样的。但是对于一些特殊场合，例如哈希数据在网络上传播的时候，你可能就很希望数据是有序的。



那么我们怎么解决无序哈希的问题呢？我们的语言已经有了对无序哈希的支持，怎么样让我们的程序变动最少？这个都是应该解决的问题。对于这个问题，我的思路是在哈希的抽象语法表示中，加入两个新的字段:

1. `IsOrdered`表示哈希是否是有序哈希
2. Order数组：用来存放哈希的key。这样遍历的时候，我们就不是使用原有的方式来遍历哈希了，而是使用这个Order数组。

那么我们的脚本语言如何表示有序哈希呢？先来看一下无序哈希的例子：

```go
hs = {"a": 1, "name":"hhf"}
```

我想到的解决办法是在声明哈希的时候，在`{`的前面加入一个`@`字符来表示有序哈希（当然这种方法可能并不是最完美的）：

```go
ordered_hs = @{"a": 1, "name":"hhf"}
```

因为这个`@`字符是我们在`函数装饰器`一节加入的，所以我们无需新增词元，也无需更改词法解析器。



有了上面的简短介绍，接下来来看我们需要做哪些更改：

1. 在抽象语法树（AST）的源码`ast.go`中，给`HashLiteral`结构新增字段。
2. 在语法解析器（Parser）的源码`parser.go`中，增加对`有序哈希`的语法解析。
3. 在对象系统的源码`object.go`中，修改哈希对象的相关逻辑。
4. 在解释器（Evaluator）的源码`eval.go`中修改对哈希字面量的解释及哈希遍历的解释。



## 抽象语法树（AST）的更改

前面的讲述中，我们提到，需要给哈希加入两个字段：`IsOrdered`和`Order数组`：

```go
//ast.go
type HashLiteral struct {
	Token       token.Token
	Pairs       map[Expression]Expression
	RBraceToken token.Token
	IsOrdered   bool         //哈希是否是有序的
	Order       []Expression //为了保持哈希的key的顺序
}
```

有了这个更改后，对于哈希字面量的字符串表示，也需要更改一下：

```go
//ast.go
//哈希字面量的字符串表示
func (h *HashLiteral) String() string {
	var out bytes.Buffer

	pairs := []string{}
	if h.IsOrdered {
		for _, key := range h.Order { //如果是有序哈希，则通过Order数组遍历哈希的key
			value, _ := h.Pairs[key]
			pairs = append(pairs, key.String()+": "+value.String())
		}
	} else {
		for key, value := range h.Pairs {
			pairs = append(pairs, key.String()+":"+value.String())
		}
	}

	out.WriteString("{")
	out.WriteString(strings.Join(pairs, ", "))
	out.WriteString("}")

	return out.String()
}
```

第7-12行的`if`判断是新增的逻辑。如果是个有序哈希，我们使用`Order数组`来遍历获取哈希的`key/value`。



## 语法解析器（Parser）的更改

如果读者还有印象的话，在实现`函数装饰器`的时候，我们注册了`@`词元的前缀回调函数来处理`函数装饰器`：

```go
//parser.go
func (p *Parser) registerAction() {
	//...
	p.registerPrefix(token.TOKEN_AT, p.parseDecorator)

}

func (p *Parser) parseDecorator() ast.Expression {
	//...
}
```

我们需要更改一下`parseDecorator`这个函数的逻辑，使其增加对有序哈希的支持：

```go
//parser.go
func (p *Parser) parseDecorator() ast.Expression {
    if p.peekTokenIs(token.TOKEN_LBRACE) { //如果'@'后是一个'{'的话，表示是一个有序哈希(ordered hash)
		p.nextToken() //skip the '@'
		result := p.parseHashLiteral().(*ast.HashLiteral) //调用既有的`parseHashLiteral`函数来解析哈希字面量
		result.IsOrdered = true //将其`IsOrdered`设置为true
		return result
	}

	//既有逻辑
}
```

第3行的`if`判断就是新增的逻辑，还是比较简单的。

接下来，我们还需要更改`parseHashLiteral`这个函数，对于解析出来的每个哈希的key，我们需要将其加入新增的`Order数组`这个字段中：

```go
//parser.go
func (p *Parser) parseHashLiteral() ast.Expression {
	hash := &ast.HashLiteral{Token: p.curToken, Order: []ast.Expression{}}
	hash.Pairs = make(map[ast.Expression]ast.Expression)
	for !p.peekTokenIs(token.TOKEN_RBRACE) {
		p.nextToken()
		key := p.parseExpression(LOWEST)
		if !p.expectPeek(token.TOKEN_COLON) {
			return nil
		}

		p.nextToken()
		value := p.parseExpression(LOWEST)
		hash.Pairs[key] = value
		hash.Order = append(hash.Order, key) //将遍历的key加入Order数组中
		if !p.peekTokenIs(token.TOKEN_RBRACE) && !p.expectPeek(token.TOKEN_COMMA) {
			return nil
		}
	}

	if !p.expectPeek(token.TOKEN_RBRACE) {
		return nil
	}

	return hash
}
```

这里有两处更改。第2行，我们在生成哈希字面量结构的时候，使用`Order: []ast.Expression{}`方式，将`Order`字段初始化成一个空的数组。第15行，我们解析完一个哈希的`key`后，将其加入了`Order`数组中。



## 哈希对象（Object）的更改

有了有序哈希后，我们的哈希对象结构，也需要做相应的更改。直接看代码：

```go
//object.go
type Hash struct {
	Pairs     map[HashKey]HashPair
	IsOrdered bool
	Order     []HashKey
}
```

和抽象语法树小节的更改很像，我们在第4-5行加入了两个字段`IsOrdered`和`Order`数组。

不知道读者是否还有印象，我们有一个`NewHash`这个工具函数：

```go
//object.go
func NewHash() *Hash {
	return &Hash{Pairs: make(map[HashKey]HashPair)}
}
```

我们也需要更改一下这个函数，让它能够将我们新增的这个`Order`字段初始化一下：

```go
//object.go
func NewHash() *Hash {
	return &Hash{Pairs: make(map[HashKey]HashPair), Order: []HashKey{}}
}
```

我们还需要更改哈希对象的`Inspect()`函数的逻辑，以反映有序哈希的情况：

```go
//object.go
func (h *Hash) Inspect() string {
	var out bytes.Buffer
	pairs := []string{}
	if h.IsOrdered { //如果是有序哈希
		//hk:hash key
		for _, hk := range h.Order { //通过遍历`Order`数组的方式遍历哈希的key/value
			var key, val string

			pair, _ := h.Pairs[hk]
			if pair.Key.Type() == STRING_OBJ {
				key = "\"" + pair.Key.Inspect() + "\""
			} else {
				key = pair.Key.Inspect()
			}

			if pair.Value.Type() == STRING_OBJ {
				val = "\"" + pair.Value.Inspect() + "\""
			} else {
				val = pair.Value.Inspect()
			}

			pairs = append(pairs, fmt.Sprintf("%s:%s", key, val))
		}
	} else {
		//既有逻辑...
	}
	//...
}
```

第5行的`if`判断是新增的逻辑，用来反映`有序哈希`的情况。我们的哈希对象，还提供了一些内置的方法，以方便用户脚本来使用：

```go
//object.go
func (h *Hash) CallMethod(line string, scope *Scope, method string, args ...Object) Object {
	switch method {
	case "keys":
		return h.keys(line, args...)
	case "values":
		return h.values(line, args...)
	case "pop", "delete", "remove":
		return h.pop(line, args...)
	case "push", "set":
		return h.push(line, args...)
	case "get":
		return h.get(line, args...)
	}

	return newError(line, ERR_NOMETHOD, method, h.Type())
}
```

上面的代码是既有的逻辑。对于上面的这些哈希对象的方法，除了`get`函数不需要更改（它只是根据key取相应的value），其它的都需要作出相应的更改，来反映有序哈希的情况。下面分别列出代码：

```go
//object.go
//取得哈希的所有key
func (h *Hash) keys(line string, args ...Object) Object {
	keys := &Array{}
	if h.IsOrdered { //如果是有序哈希
		 //hk:hash key
		for _, hk := range h.Order { //通过遍历`Order`数组来遍历哈希
			pair, _ := h.Pairs[hk]
			keys.Members = append(keys.Members, pair.Key)
		}
	} else { //这个else中是既有逻辑
		for _, pair := range h.Pairs {
			keys.Members = append(keys.Members, pair.Key)
		}
	}

	return keys
}
```

第5-10行的`if`判断是新增的逻辑。

```go
//object.go
//取得哈希的所有value
func (h *Hash) values(line string, args ...Object) Object {
	values := &Array{}
	if h.IsOrdered { //如果是有序哈希
		//hk:hash key
		for _, hk := range h.Order {  //通过遍历`Order`数组来遍历哈希
			pair, _ := h.Pairs[hk]
			values.Members = append(values.Members, pair.Value)
		}
	} else { //这个else中是既有逻辑
		for _, pair := range h.Pairs {
			values.Members = append(values.Members, pair.Value)
		}
	}
	return values
}
```

第5-10行的`if`判断是新增的逻辑。

```go
//object.go
func (h *Hash) pop(line string, args ...Object) Object {
	if len(args) != 1 {
		return newError(line, ERR_ARGUMENT, "1", len(args))
	}
	hashable, ok := args[0].(Hashable)
	if !ok {
		return newError(line, ERR_KEY, args[0].Type())
	}

	hk := hashable.HashKey()
	if hashPair, ok := h.Pairs[hashable.HashKey()]; ok {
		for idx, k := range h.Order { // 将'key'从'Order'数组中移除
			r := reflect.DeepEqual(hk, k)
			if r {
				h.Order = append(h.Order[:idx], h.Order[idx+1:]...)
				break
			}
		}

		delete(h.Pairs, hk)
		return hashPair.Value
	}

	return NIL
}
```

第13-19行是新增的逻辑。移除哈希key的时候，我们还需要将其从`Order`数组中移除。这里的比较使用的是`reflect.DeepEqual`的方式，可能不是很高效。

```go
//object.go
func (h *Hash) push(line string, args ...Object) Object {
	if len(args) != 2 {
		return newError(line, ERR_ARGUMENT, "2", len(args))
	}
	if hashable, ok := args[0].(Hashable); ok {
		hk := hashable.HashKey()
		if _, exists := h.Pairs[hk]; !exists { //如果key不存在，我们将其加入'Order'数组
			h.Order = append(h.Order, hk)
		}

		h.Pairs[hashable.HashKey()] = HashPair{Key: args[0], Value: args[1]}
	} else {
		return newError(line, ERR_KEY, args[0].Type())
	}

	return h
}
```

第7-10行是新增的代码，当我们给哈希添加key/value的时候，我们需要根据情况将`key`加入哈希对象的`Order`这个数组字段中。



## 解释器（Evaluator）的更改

对于解释器，我们需要更改两个地方。一个是更改哈希字面量的解释函数`evalHashLiteral`。另一个是哈希遍历函数`evalForEachMapExpression`。先来看一下`evalHashLiteral`这个函数，由于这个函数的变动比较大，所以我将既有的函数给注释掉了，重新提供了一个实现：

```go
//eval.go
func evalHashLiteral(node *ast.HashLiteral, scope *Scope) Object {
	hash := NewHash() //使用工具函数`NewHash`生成一个新的哈希对象
	hash.IsOrdered = node.IsOrdered

	for _, key := range node.Order { //遍历哈希字面量节点的`Order`数组
		k := Eval(key, scope) //解释key
		if k.Type() == ERROR_OBJ {
			return k
		}

		if _, ok := k.(Hashable); !ok { //key是否是可遍历的
			return newError(node.Pos().Sline(), ERR_KEY, k.Type())
		}

		value, _ := node.Pairs[key] //解释value
		v := Eval(value, scope)
		if v.Type() == ERROR_OBJ {
			return v
		}

		 //这里使用的是哈希对象的'push'方法，它会负责处理哈希key的有序/无序的插入。
		hash.push(node.Pos().Sline(), k, v)
	}
	return hash
}
```

虽然重写了相关的代码，但是逻辑也并不是很复杂。接下来让我们看一下`evalForEachMapExpression`函数的变更：

```go
//eval.go
func evalForEachMapExpression(fml *ast.ForEachMapLoop, scope *Scope) Object { //fml:For Map Loop
	//...

	//for _, pair := range hash.Pairs { //既有的旧代码
	for _, hk := range hash.Order {
		pair, _ := hash.Pairs[hk]
		//...
	}

	return arr
}
```

我们将第5行的原有的遍历注释掉，更改成了第6行使用哈希对象的`Order`数组来遍历。其实对于这里的变更来说，并不完美。为什么呢？因为我这里没有区分有序哈希和无序哈希的情况。实际的更改应该可能像下面这样：

```go
//eval.go
func evalForEachMapExpression(fml *ast.ForEachMapLoop, scope *Scope) Object { //fml:For Map Loop
	//...

    if hash.IsOrdered {
	for _, hk := range hash.Order {
		pair, _ := hash.Pairs[hk]
		//...
    } else {
		//既有逻辑
	}

	retur
```

因为重复代码太多，我也懒得写一个新的函数，所以蒙混一下过关了:smile:。



## 测试

```javascript
// ordered hash
println("ordered hash")
h1 = @{"a": 1,
       "b": 2,
       "c": 3,
       "d": 4,
       "e": 5,
       "f": 6,
       "g": 7,
}
println(h1)

println()
println("un-ordered hash")
// un-ordered hash
h2 = {"a": 1,
      "b": 2,
      "c": 3,
      "d": 4,
      "e": 5,
      "f": 6,
      "g": 7,
}
println(h2)
```



下一节，我们讨论单一执行文件（Single Standalone executable）支持。




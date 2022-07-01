# 哈希支持

在这一篇文章中，我们要加入对于`哈希`的支持。哈希是由零个或者多个键值对组成的。

> 对于哈希，不同的语言可能有不同的称谓。有的称为map， 有的称为dictionary，还有的称为hash map。



我们先来看一下哈希的例子：

```go
let emptyHash = {} //空哈希
let myHash = {"name": "huanghaifeng", "height": 165} //定义哈希
println(myHash["name"] //取哈希值
println(myHash["height"])
```

从上面的例子中，我们可以得出如下信息：

哈希是以花括弧`{}`作为界定符的，这和数组以方括号`[]`作为界定符有所区别。哈希的key和value使用`:`作为分隔符。取哈希的值使用的是和数组一样的方式，都是使用`[]`。



下面看一下我们需要做的更改：

1. 在词元（Token）源码`token.go`中加入新的词元（Token）类型(`:`)
2. 在词法分析器（Lexer）源码`lexer.go`中加入对`:`的识别
3. 在抽象语法树（AST）的源码`ast.go`中加入`哈希`对应的抽象语法表示。
4. 在语法解析器（Parser）的源码`parser.go`中加入对`哈希`的语法解析。
4. 在对象（Object）系统中的源码`object.go`中加入新的`哈希对象(Hash Object)`。
5. 在解释器（Evaluator）的源码`eval.go`中加入对`哈希`的解释。



## 词元（Token）更改

```go
//token.go
const (
	//...
	TOKEN_COLON     //:
)

//词元类型的字符串表示
func (tt TokenType) String() string {
	switch tt {
	//...
	case TOKEN_COLON:
        return ":"
	}
}
```



## 词法分析器（Lexer）的更改

我们需要在词法分析器（Lexer）的`NextToken()`函数中加入对`[`和`]`的识别（9-12行）：

```go
//lexer.go

//获取下一个词元（Token）
func (l *Lexer) NextToken() token.Token {
	//...

	switch l.ch {
	//...
	case ':':
		tok = newToken(token.TOKEN_COLON, l.ch)
	}
	//...
}

```



## 抽象语法树（AST）的更改

从文章开头的例子中，我们大致可以了解到`哈希`的一般表达形式：

```go
hash = {
	<key-expr1> : <value-expr1>,
	<key-expr2> : <value-expr2>,
	<key-expr3> : <value-expr3>,
	//...
}
```

从上面的形式我们可以得出`哈希`的抽象语法表示：

```go
//ast.go
//哈希字面量
type HashLiteral struct {
	Token       token.Token
	Pairs       map[Expression]Expression //键值对
    RBraceToken token.Token //用在"End()"方法中
}

func (h *HashLiteral) Pos() token.Position {
	return h.Token.Pos
}

func (h *HashLiteral) End() token.Position {
	return token.Position{Filename: h.Token.Pos.Filename, Line: h.RBraceToken.Pos.Line, 
                          Col: h.RBraceToken.Pos.Col + 1}
}

func (h *HashLiteral) expressionNode()      {}
func (h *HashLiteral) TokenLiteral() string { return h.Token.Literal }

func (h *HashLiteral) String() string { //哈希的字符串表示
	var out bytes.Buffer

	pairs := []string{}
	for key, value := range h.Pairs {
		pairs = append(pairs, key.String()+":"+value.String())
	}

	out.WriteString("{")
	out.WriteString(strings.Join(pairs, ", "))
	out.WriteString("}")

	return out.String()
}
```

这里的代码也是大家非常熟悉的。

取哈希值的方式如下：

```go
hash[<key-expression>]
```

由于这里使用的是和取数组元素相同的`[]`。所以我们不用为取哈希元素再新建一个抽象语法表示，直接使用上一节中的`IndexExpression（索引表达式）`即可。



## 语法解析器（Parser）的更改

我们需要给词元类型`TOKEN_LBRACE（即'{'）`注册前缀表达式回调函数。

来看一下代码：

```go
//parser.go
func (p *Parser) registerAction() {
	//...
	p.registerPrefix(token.TOKEN_LBRACE, p.parseHashLiteral) //注册前缀表达式
}

```

`parseHashLiteral`函数实现如下：

```go
//parser.go
func (p *Parser) parseHashLiteral() ast.Expression {
    hash := &ast.HashLiteral{Token: p.curToken, Pairs: make(map[ast.Expression]ast.Expression)}

	for !p.peekTokenIs(token.TOKEN_RBRACE) {
		p.nextToken()
		key := p.parseExpression(LOWEST) //解析key表达式
		if !p.expectPeek(token.TOKEN_COLON) {
			return nil
		}

		p.nextToken()
		value := p.parseExpression(LOWEST) //解析value表达式
		hash.Pairs[key] = value

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

代码不多，理解起来也比较简单。



## 对象（Object）系统的更改

我们这里建立的哈希的key可以是一个字符串对象，可以是一个数字对象，也可以是一个布尔型或者元祖对象。但不能是数组或者另外的哈希。例如：

```javascript
let h = {
	"key": "hhf",   //ok
	10: 10,         //ok
	true: "yes",    //ok
	(1，2): "xxx",  //ok
	[1,2]: [1,2]   //error,数组不能作为哈希key
}
```

我们需要往对象系统中增加一个`哈希对象(Hash Object)`。哈希对象里面存放的是什么呢？读者可能已经想到了，我们可以使用go语言的map作为哈希对象的内部表示：

```go
//object.go
type Hash struct {
    Pairs: map[Object]Object
}
```

这个定义是非常自然且明显的。但是这个定义有个问题：我们怎么填充这个`Pairs`呢？还有我们怎么取得哈希的值？让我们看一个例子：

```perl
hash = {"language": "Magpie"}
hash["language"]
```

第一行我们定义了一个哈希字面量，我们的`Hash`对象的Pairs的内容如下：

```javascript
Pairs["language"] = "Magpie"
```

这个`Pairs`map的key是一个字符串对象，value也是一个字符串对象。好像没有什么问题。但是问题出在第二行`hash["language"]`，我们使用索引表达式来取得key为"language"的哈希值，我们希望取得结果是"Magpie"这个值。有的读者就会问了，这也没啥问题啊。让我来说明一下。

对于字符串对象，让我们来看一下当解释器（Evaluator）遇到字符串字面量的时候，是如何解释（Evaluating）的：

```go
//eval.go
func evalStringLiteral(s *ast.StringLiteral, scope *Scope) Object {
	return NewString(s.Value) //返回一个新的字符串对象
}

//object.go
func NewString(s string) *String {
	return &String{String: s} //生成一个新的字符串对象
}
```

看到了吗？如果我们遇到了字符串，我们是生成了一个<span style="color:blue">__新的字符串对象__</span> 。说到这里，有的读者可能已经看出问题来了吧。我们在刚才的例子中的第一行使用`hash = {"language": "Magpie"}`创建哈希字面量对象的时候，这个key是一个字符串。而当我们在第二行使用`hash["language"]`取得其值的时候，这里的`key`是一个新的字符串，和哈希创建时候的字符串不是一个对象（它们的地址是不同的，虽然它们中保存的值都是"language"）。 所以第二行使用`hash["language"]`取值的时候，返回的不是`"Magpie"`，而是`nil`。

我们用代码来说明一下可能更容易理解：

```go
hash = {"language": "Magpie"} 
/*
这时候的Pairs map如下：
    Pairs[&String{String:"language"}] = &String{String:"Magpie"}
*/

hash["language"]
/*
这时候的Pairs map如下：
    Pairs[&String{String:"language"}]
*/

```

第4行和第10行中，`Pairs`中的Key是两个完全不同的`String`对象（虽然存储的内容都是"language"）。这不可能取到，是吧。



还不明白？好吧，让我们写一个`go`程序来验证一下：

```go
language1 := &String{String: "language"} //字符串对象
magpie := &String{String: "Magpie"}      //字符串对象

pairs := map[Object]Object{}
pairs[language1] = magpie
fmt.Printf("pairs[language1]=%+v\n", pairs[language1])
// => pairs[language1]=&{String:Magpie}


language2 := &String{String: "language"} //字符串对象(存储的内容和language1变量是一样的，都是"language")
fmt.Printf("pairs[language2]=%+v\n", pairs[language2])
// => pairs[language2]=<nil>

fmt.Printf("(language1 == language2)=%t\n", language1 == language2)
// => (language1 == language2)=false
```

仔细看第6行和第11行，打印的结果完全不同。同时从第13行的结果也可以看到，`language1`和`language2`不相等（虽然里面保存的内容都是"language"）。



那怎么解决上面说的问题呢？我们可以遍历这个`Pairs`map，取出每个对象，检查其是否是一个`String`对象，并比较`String`对象中存储的值是否是"language"，如果是，就得到相应的值。这种方法不是不可以，但是这样的查找方法，使得哈希获取值的时间不是`O(1)`了，这并不是我们期望的。

> 关于O(1)是啥意思，读者可以到网上查找相关资料。对于这里的哈希key的查找来说，就是我们希望无论它的key的长短或者存储的内容多与少，找到key所对应的值，使用的时间都是固定的。

我们需要的是一种方法，对于给定的`对象(Object)`，它能够生成一个哈希值，这样我们就可以很容易的比较它们，并作为哈希的key。例如，对于字符串对象，我们需要一种方法，让它能够生成一个哈希值，用来和别的有相同内容的字符串对象进行比较。但是这个哈希key绝不能和另外一个数字对象或者布尔对象的哈希key值相同，就是说对于不同的类型，它们的哈希的key值必须不同。举个例子说明一下：

```go
name1 := &String{Value: "HHF"}
name2 := &String{Value: "HHF"}
name3 := &String{Value: "XXX"}

//我们期待如下的结果
//name1.HashKey() == name2.HashKey()
//name1.HashKey() != name3.HashKey()
//name2.HashKey() != name3.HashKey()

num1   := &Number{Value: 10}
num2   := &Number{Value: 10}
num3   := &Number{Value: 35}
//我们期待如下的结果
//num1.HashKey() == num2.HashKey()
//num1.HashKey() != num3.HashKey()

//num1.HashKey() != name1.HashKey()
```

上面的`name1`和`name2`是两个不同的字符串，但是值都是`HHF`, 我们希望它们的`HashKey()`方法返回的值是相同的。`name1`和`name3`两个字符串存储的值不同，它们的`HashKey()`方法返回的值当然需要是不同的。



为了节省篇幅，不让这个问题占用太多的时间，让我们来给数字对象、布尔对象和字符串对象定义`HashKey()`方法，它需要满足我们上面例子的期待结果：

1. 如果__对象类型相同，且对象存储的内容相同__，则比较它们的`HashKey()`方法得到的结果应该是__相等__的
2. 如果__对象类型不相同__，则比较它们的`HashKey()`方法得到的结果应该是__不相等__的



```go
//object.go
import "hash/fnv" //引入计算哈希值的包，用来计算哈希值

type HashKey struct {
	Type  ObjectType //对象类型
	Value uint64     //哈希函数计算后得到的哈希值
}

//数字对象
func (i *Number) HashKey() HashKey {
	return HashKey{Type: i.Type(), Value: uint64(i.Value)}
}

//布尔对象
func (b *Boolean) HashKey() HashKey {
	var value uint64
	if b.Bool {
		value = 1
	} else {
		value = 0
	}

	return HashKey{Type: b.Type(), Value: value}
}

//字符串对象
func (s *String) HashKey() HashKey {
	h := fnv.New64a() //返回一个64bit的FNV-1a哈希
	h.Write([]byte(s.String))
    return HashKey{Type: s.Type(), Value: h.Sum64()} //h.Sum64():取得哈希值(uint64类型)
}

//元组：由于还没有讲到，所以这里暂时不列出实现

```

所有的`HashKey()`方法都返回一个`HashKey`结构（不是`&HashKey结构`）。`HashKey`结构也比较简单，其中的`Type`字段表示对象类型，`Value`字段表示计算的哈希值。这个`HashKey`结构可以作为`go语言`map的key。这样，我们根据哈希key去取对应的值的时候，就可以用下面的代码（伪代码）：

```go
if pair, ok := hashObj.Pairs[key.HashKey()]; !ok {
	return NIL //没找到返回nil
}

return pair.Value //找到则返回对应的值
```

让我们写几个简单的小程序测试一下：

__数字对象__:

```go
i1 := &Number{Value: 10}
i2 := &Number{Value: 10}
i3 := &Number{Value: 25}

if i1.HashKey() == i2.HashKey() {
    fmt.Println("i1的HashKey等于i2的HashKey")
}

if i1.HashKey() != i3.HashKey() {
    fmt.Println("i1的HashKey不等于i3的HashKey")
}
```

结果:

```
i1的HashKey等于i2的HashKey
i1的HashKey不等于i3的HashKey
```



__布尔对象__:

```go
b1 := &Boolean{Bool: true}
b2 := &Boolean{Bool: true}
b3 := &Boolean{Bool: false}

if b1.HashKey() == b2.HashKey() {
    fmt.Println("b1的HashKey等于b2的HashKey")
}

if b1.HashKey() != b3.HashKey() {
    fmt.Println("b1的HashKey不等于b3的HashKey")
}
```

结果:

```
b1的HashKey等于b2的HashKey
b1的HashKey不等于b3的HashKey
```



__字符串对象__:

```go
s1 := &String{String: "Hello"}
s2 := &String{String: "Hello"}
s3 := &String{String: "Welcome"}

if s1.HashKey() == s2.HashKey() {
    fmt.Println("s1的HashKey等于s2的HashKey")
}

if s1.HashKey() != s3.HashKey() {
    fmt.Println("s1的HashKey不等于s3的HashKey")
}
```

结果:

```
s1的HashKey等于s2的HashKey
s1的HashKey不等于s3的HashKey
```



__不同对象__:

```go
str1 := &String{String: "10"}
num1 := &Number{Value: 10}

if str1.HashKey() != num1.HashKey() {
    fmt.Println("str1的HashKey不等于num1的HashKey")
}
```

结果:

```
str1的HashKey不等于num1的HashKey
```

测试结果表明，这和我们期待的是一致的。



有了上面的说明，我们来看一下`Hash对象`的代码：

```go
//object.go
var (
	HASH_OBJ = "HASH"
)

type HashPair struct { //key-value键值对
	Key   Object
	Value Object
}

type Hash struct {
	Pairs map[HashKey]HashPair //以HashKey结构作为map的key
}

func (h *Hash) Type() ObjectType { return HASH_OBJ }
func (h *Hash) Inspect() string {
	var out bytes.Buffer
	pairs := []string{}
	for _, pair := range h.Pairs {
		var key, val string
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

	out.WriteString("{")
	out.WriteString(strings.Join(pairs, ", "))
	out.WriteString("}")
	return out.String()
}

//工具(utility)函数
func NewHash() *Hash { //创建一个空的哈希对象
	return &Hash{Pairs: make(map[HashKey]HashPair)}
}
```

第11-13行定义了咱们的`Hash`对象。第12行我们定义的`Pairs`这个map变量，它的值为`HashPair`结构，这个结构保存了原始的key-value(键值对)对象。同时，我们在43-45行定义了一个工具函数`NewHash()`，它创建一个空的哈希对象，这个函数我们将来会用到。



最后，还有一个需要说明的，我们需要加入一个`Hashable`接口：

```go
//object.go
type Hashable interface {
	HashKey() HashKey
}
```

在我们的解释器中，当解释哈希字面量或者哈希索引表达式的时候，可以使用这个接口来判断一个对象是否可以作为哈希的key。由于我们的数字对象，布尔对象和字符串对象都实现了这个`HashKey()`方法，因此它们都可以作为哈希的key。



## 解释器（Evaluator）的更改

我们需要在解释器（Evaluator）的`Eval`函数的`switch`分支中加入对`哈希字面量表达式`的处理：

```go
//eval.go

func Eval(node ast.Node, scope *Scope) (val Object) {
	switch node := node.(type) {
	//...

	case *ast.HashLiteral: //处理哈希字面量
		return evalHashLiteral(node, scope)
	}

	return nil
}

//解释哈希字面量
func evalHashLiteral(node *ast.HashLiteral, scope *Scope) Object {
	pairs := make(map[HashKey]HashPair) //创建一个key为'HashKey', Value为HashPair的map
	for keyNode, valueNode := range node.Pairs { //遍历哈希字面量表达式的每个键值对
		keyObj := Eval(keyNode, scope) //解释key
		if isError(keyObj) {
			return keyObj
		}

		hashKey, ok := keyObj.(Hashable) //keyObj是否可哈希化
		if !ok {
			return newError(node.Pos().Sline(), ERR_KEY, keyObj.Type())
		}

		valueObj := Eval(valueNode, scope) //解释value
		if isError(valuObj) {
			return valueObj
		}

        hashed := hashKey.HashKey() //取得计算后的哈希值
		pairs[hashed] = HashPair{Key: keyObj, Value: valueObj}
	}

	return &Hash{Pairs: pairs} //返回哈希对象
}

//errors.go
var (
	//...
	ERR_KEY = "key error: type %s is not hashable"
)

```

我们在第7行增加了一个处理哈希表达式的`case`分支。实际的代码在`evalHashLiteral`函数中。

我们还需要更改`evalIndexExpression`这个函数，用来处理哈希索引表达式求值：

```go
//eval.go

//处理索引表达式
func evalIndexExpression(node *ast.IndexExpression, left, index Object) Object {
	switch {
	//...
	case left.Type() == HASH_OBJ:
		return evalHashIndexExpression(node.Pos().Sline(), left, index)
	}
}

//处理哈希的索引表达式：hash[idx]
func evalHashIndexExpression(line string, hash, index Object) Object {
	hashObject := hash.(*Hash)
	key, ok := index.(Hashable) //key是否可哈希化
	if !ok {
		return newError(node.Pos().Sline(), ERR_KEY, index.Type())
	}

	pair, ok := hashObject.Pairs[key.HashKey()] //根据key取得相应的pair
	if !ok {
		return NIL //没找到，返回nil
	}

	return pair.Value //返回对应的哈希值
}
```

还有一个地方需要说明，我们的语言允许用户书写如下语句：

```javascript
hash = {"name": "hhf"]
if hash { //判断哈希是否存在键值对
    println("hash key-value pair is larger than zero")
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
		case HASH_OBJ:
			if len(obj.(*Hash).Pairs) == 0 {
				return false
			}
		} //end switch

		return true
	}
}
```

代码的8-11行是新增的代码。



当然别忘记修改我们的老朋友`len内置函数`。对于哈希来说，它返回哈希对象的键值对的个数：

```go
//builtin.go

func lenBuiltin() *Builtin {
	return &Builtin{
		Fn: func(line string, scope *Scope, args ...Object) Object {
			//...
			switch arg := args[0].(type) {
			//...
			case *Hash:
				return NewNumber(float64(len(arg.Pairs)))
			}
		},
	}
}
```

9-10行是新增加的`case`分支。

## 测试

下面我们写一个简单的程序测试一下：
```go
//main.go
func TestEval() {
	tests := []struct {
		input    string
		expected string
	}{
		{`let h = {"name":"hhf","height":165}; println(h["name"], h["height"])`, "nil"},
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

运行结果如下：

```
huanghaifeng
165
```



下一节，我们会提供对`方法调用(method-call)`的支持：`obj.method(param1, param2, ...)`


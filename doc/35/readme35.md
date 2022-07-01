# `struct`支持

这一节中，我们将给`magpie语言`提供一个相对高级的支持：`struct语句`支持。我们先给个使用`struct`的例子，让读者能够有所了解：

```javascript
struct math
{
     let name = "math"
     let dummy = 0

    // 构造函数
    fn init(x, y) {
        self.x = x
        self.y = y
        printf("Hello %s\n", self.name)
    }

    // unexported，外部调用会出错
    fn add() {
        return self.x + self.y
    }

    fn Add() {
        return self.add()
    }

    fn Sub() {
        return self.x - self.y
    }

    fn Print(msg) {
        printf("Hello %s, self.dummy = %d\n", msg, self.dummy)
    }

    fn Add_then_sub() {
        add_result = self.Add()
        sub_result = self.Sub()
        printf("In add_then_sub: add_result=%d, sub_result=%d\n", add_result, sub_result)
    }
}

m1 = math(10, 12)  // 会调用math结构的'init'构造方法，并传递参数
printf("add result=%g\n", m1.Add())
printf("sub result=%g\n", m1.Sub())
m1.Add_then_sub()
m1.Print("hhf")

// 会报告如下错误：cannot refer to unexported name m1.add
printf("add result=%g\n", m1.add())
```

这里有几点需要注意：

1. `struct`中所有声明的变量和函数，在内部使用的时候都必须加上`self`，否则会报错
2. 所有的方法，如果希望在外部调用的话（即公有方法），那么方法的首字母必须大写
3. `struct`的构造函数是名为`init`的方法
4. 实例化`struct`使用类似函数调用的方式：`structName(param1, param2, ...)`



现在让我们看一下需要做哪些更改：

1. 在词元（Token）源码`token.go`中加入新的词元（Token）类型
2. 在抽象语法树（AST）的源码`ast.go`中加入`struct`对应的抽象语法表示
3. 在语法解析器（Parser）的源码`parser.go`中加入对`struct`的语法解析
4. 在对象（Object）系统源码`object.go`中，新增一个`结构对象(Struct Object)`
5. 在解释器（Evaluator）的源码`eval.go`，加入对`struct`的解释



## 词元（Token）的更改

因为都是读者再熟悉不过的内容，不做解释，直接看代码：

```go
//token.go
const (
	//...
	TOKEN_STRUCT   //struct

)

//词元的字符串表示
func (tt TokenType) String() string {
	switch tt {
	//...
	case TOKEN_STRUCT:
		return "struct"
	}
}

var keywords = map[string]TokenType{
	//...
	"struct":   TOKEN_STRUCT,
}
```



## 抽象语法树（AST）的更改

我们先来想一下，对于`struct语句`的抽象语法表示，需要什么样的信息？

从最开始的`struct`使用例，我们可以抽象出`struct`的一般表示：

```c
struct structName { block }
```

1. 词元信息（是的，这个是所有的抽象语法表示都有的）
2. 结构名称（Name）
3. block块

有了上面的分析，让我们看一下代码：

```go
//ast.go
// 结构(struct)的抽象语法表示
type StructStatement struct {
	Token token.Token
	Name  string //结构名
	Block       *BlockStatement
	RBraceToken token.Token     //右花括弧，仅用在'End()'方法中
}

func (s *StructStatement) Pos() token.Position {
	return s.Token.Pos
}

//结构(struct)的终了位置
func (s *StructStatement) End() token.Position {
	return s.RBraceToken.Pos
}

//结构(struct)是一个语句
func (s *StructStatement) statementNode()       {}

func (s *StructStatement) TokenLiteral() string { return s.Token.Literal }

//结构(struct)的字符串表示
func (s *StructStatement) String() string {
	var out bytes.Buffer

	out.WriteString(s.Token.Literal + " ")
	out.WriteString(s.Name)

	out.WriteString("{ ")
	out.WriteString(s.Block.String()) //结构(struct)的块语句的字符串表示
	out.WriteString(" }")

	return out.String()
}
```



## 语法解析器（Parser）的更改

对于`struct语句(struct-statement)`的语法解析，我们需要在`parseStatement`函数中加入一个`case`分支：

```go
//parser.go

func (p *Parser) parseStatement() ast.Statement {
	switch p.curToken.Type {
	//...
	case token.TOKEN_STRUCT:
		return p.parseStructStatement()
	}
}

//解析`struct语句`
func (p *Parser) parseStructStatement() ast.Statement {
	st := &ast.StructStatement{Token: p.curToken}

	p.nextToken()
	st.Name = p.curToken.Literal //获取结构名

	if !p.expectPeek(token.TOKEN_LBRACE) { //结构名后面必须紧跟一个`{`
		return nil
	}

	st.Block = p.parseBlockStatement() //调用块语句解析函数`parseBlockStatement`
	st.RBraceToken = p.curToken //记录右花括弧的位置，即struct语句的结束位置

	return st
}
```

第6行的`case`分支用来处理`struct`语句，实际的代码在`parseStructStatement`函数中。这个`parseStructStatement`函数是不是比想象中简单得多？



## 对象（Object）系统的更改

我们需要创建一个`Struct对象`。这个`Struct对象`中应该包含什么样的信息呢？

`Struct结构`里面有字段和方法，它们的解释都是在`Struct对象`的`作用域（Scope）`中执行的，所以我们需要一个Scope字段。还有别的需要吗？不需要了。下面来看看代码：

```go
//object.go
const (
	//...
	STRUCT_OBJ       = "STRUCT"
)

//结构对象
type Struct struct {
	Scope   *Scope //结构的Scope
}
```

既然我们要创建一个`Struct对象(Struct Object)`，所以这个`Struct结构`需要实现`Object`接口的所有方法：

```go
//object.go

type Struct struct {
	Scope *Scope //struct's scope
}

func (s *Struct) Inspect() string {
	var out bytes.Buffer
	out.WriteString("( ")
	for k, v := range s.Scope.store {
		out.WriteString(k)
		out.WriteString("->")
		out.WriteString(v.Inspect())
		out.WriteString(" ")
	}
	out.WriteString(" )")

	return out.String()
}

func (s *Struct) Type() ObjectType { return STRUCT_OBJ }
func (s *Struct) CallMethod(line string, scope *Scope, method string, args ...Object) Object {
	var fn2 Object
	var fn *Function
	var ok bool

	if fn2, ok = s.Scope.Get(method); !ok { //在Struct的作用域中查找
		return newError(line, ERR_NOMETHOD, method, s.Type()) //找不到，则返回错误对象
	}

	fn = fn2.(*Function)
	extendedScope := extendFunctionScope(fn, args)
	extendedScope.Set("self", s)
	obj := Eval(fn.Literal.Body, extendedScope)
	return unwrapReturnValue(obj)
}
```

这里我们主要讲解一下`CallMethod`方法。它首先从`Struct`对象的`Scope`中查找函数名，找不到则报错。找到的话，将其对象转换为`函数对象(Function Object)`（31行）.

再来看`Struct对象`的第33行代码`extendedScope.Set("self", s)`。这里我们在解释（Evaluating）`Struct`的`body`之前（代码第34行），将`Struct对象`自身放入key为`self`的函数作用域中。这样，就强迫函数体内部，使用`结构对象`的字段或者方法的时候，都必须加入`self`前缀，例如：`self.fields`或者`self.methods(xxx)`。不加`self`前缀，解释器就会报告【找不到方法或者字段】的运行期错误。

我们假设没有33行的代码，那么我们新建的`extendedScope`中则只包含函数的参数，在函数体的内部你要是使用了`结构`作用域中的字段，当然是找不到了。要想找到`结构`中定义的方法或者字段，必须要使用`结构`的作用域。传入`结构对象`自身，我们就可以从这个`结构`的`Scope`中获得定义的方法和字段。

## 解释器（Evaluator）的更改

首先，我们需要在`Eval()`函数的`switch`语句中增加一个`case`分支：

```go
//eval.go

func Eval(node ast.Node, scope *Scope) (val Object) {
    //...

	switch node := node.(type) {
	//...
	case *ast.StructStatement:
		return evalStructStatement(node, scope)

	}

	return nil
}

//解释Struct
func evalStructStatement(structStmt *ast.StructStatement, scope *Scope) Object {
	scope.SetStruct(structStmt) //save to scope
	return NIL
}
```

代码第8行，我们增加了一个`case`分支用来解释`Struct语句`。实际的解释代码放在`evalStructStatement`函数中，在这个函数内部我们只是简单的将`结构语句(*ast.StructStatement)`放入作用域，key为结构（struct）名。简单来说就是遇到`结构声明`，就将`结构语句(*ast.StructStaetment)`保存起来。这个`SetStruct`是一个新增的函数，我们来看一下它的实现：

```go
//scope.go
type Scope struct {
	//...
	structStore map[string]*ast.StructStatement //新增字段
}

func NewScope(p *Scope, w io.Writer) *Scope {
	s := make(map[string]Object)
	ss := make(map[string]*ast.StructStatement) //创建map
	ret := &Scope{store: s, parentScope: p, structStore: ss} //创建一个Scope对象
	if p == nil {
		ret.Writer = w
	} else {
		ret.Writer = p.Writer
	}

	return ret
}

//根据参数name，从Scope中取得存储的StructStatement
func (s *Scope) GetStruct(name string) (*ast.StructStatement, bool) {
	obj, ok := s.structStore[name]
	if !ok && s.parentScope != nil {
		obj, ok = s.parentScope.GetStruct(name)
	}
	return obj, ok
}

//将结构语句(StructStatement)存入Scope。key为结构名
func (s *Scope) SetStruct(structStmt *ast.StructStatement) *ast.StructStatement {
	s.structStore[structStmt.Name] = structStmt //key是结构名称
	return structStmt
}
```

代码第4行，我们给`Scope`结构新增了一个`structStore`map，这个map的key是`结构名`，value是一个`结构语句(Struct Statement)`。有的读者会问了，这个`Scope`结构里`store`变量存储的value是一个`Object`。这里为啥存储的是一个`结构语句（*ast.StructStatement）`，而不是一个`结构对象(Struct Object)`呢？ 我们来看一个例子：

```perl
struct math {
    fn init(x, y) { # 构造方法
        self.x = x
        self.y = y
    }
    
    #...
}

#调用
m1 = math(2,3) #调用math结构的init(2,3)构造函数
m2 = math(3,7) #调用math结构的init(3,7)构造函数
```

当我们在代码中遇到一个`struct math`结构声明的时候，我们只需要将这个`struct math`结构保存到`Scope`的`structStore`这个map中，并不需要创建这个`struct math`对象。真正创建这个`struct math`对象是在调用构造函数的时候（即代码的第11和12行），我们才会分别创建两个`struct math`对象。

当代码的11、12行以`math(xxx, xxx)`的方式调用构造函数的时候，我们怎么知道这是个__结构调用__还是个__普通函数调用__呢？解决方法就是到刚才保存的`structStore`这个map中去寻找。如果以__`伪代码`__来描述的话，大致如下：

```go
//遇到了第1-8行的`struct math`结构
scoope.structStore["math"] = *ast.StructStatemment

//上面例子的11行的函数调用
if exists(scoope.structStore["math"]) { //存在
    let structStmt = scoope.structStore["math"] //取出保存的*ast.StructStatement
    //根据这个structStmt来创建我们的'math'结构对象
    createStructObject(structStmt)
}

//12行的函数调用,执行的是和上面相同的逻辑
```

我们知道，`Struct的创建`使用的是类似于函数调用的方式，那么我们就需要更改函数调用时的解释代码，来判断是否是一个__普通函数调用__还是一个__结构调用__：

```go
//eval.go
func evalCallExpression(node *ast.CallExpression, scope *Scope) Object {
	args := evalExpressions(node.Arguments, scope) //解释函数参数
	if len(args) == 1 && isError(args[0]) {
		return args[0]
	}

	//判断这个函数调用是否是一个结构调用
	if structStmt, ok := scope.GetStruct(node.Function.String()); ok {
		structObj := createStructObj(structStmt, scope) //创建结构对象
		//判读是否有'init'构造函数
		if _, ok := structObj.Scope.Get("init"); !ok { //ok为false表示没有'init'构造函数
			if len(args) > 0 { //创建的时候提供了参数
				return newError(node.Pos().Sline(), ERR_NOCONSTRUCTOR, len(args))
			}
			return structObj
		}
		//调用结构对象的`init`构造方法
		r := structObj.CallMethod(node.Pos().Sline(), scope, "init", args...)
		if r.Type() == ERROR_OBJ {
			return r
		}
		return structObj //返回结构对象
	}

	//解释函数（node.Function可能为一个标识符或者函数字面量）
	function := Eval(node.Function, scope)
	if isError(function) {
		return function
	}

	return applyFunction(node.Pos().Sline(), scope, function, args)
}
```

这里重点讨论一下第9-24行新增加的代码：

我们会将方法名作为key，传递给scope的`structStore`这个map，检查一下结构中是否存在这个方法名，如果存在的话，则创建一个结构对象，然后判断这个结构对象是否提供了`init`构造函数。没有提供的话，我们再检查用户创建结构对象的时候，是否提供了参数，如果提供了参数，而结构中却没有`init`构造函数则报错。使用例：

```perl
#声明
struct math { #c没有提供'init'构造函数
    fn Add(x, y) { return x + y }
}

#调用
m1 = math（1,2） #我们传递了两个参数，但'math'结构却没有提供'init'构造函数

#运行结果：
#  【got 2 parameters, but the struct has no 'init' method supplied】
```

如果找到了构造函数，我们就会调用上面新增的`结构对象(Struct Object)`的`CallMethod`方法，传递的方法名是`init`。

第10行代码中的`createStructObj`函数的实现如下：

```go
//eval.go
func createStructObj(structStmt *ast.StructStatement, scope *Scope) *Struct {
	structObj := &Struct{  //创建结构对象（Struct Object）
		Scope: NewScope(scope, nil), //创建一个新的Scope
	}

	//解释结构中的Block块
	Eval(structStmt.Block, structObj.Scope)
	scope.Set(structStmt.Name, structObj)

	return structObj
}
```

代码应该比较好理解。需要重点说明的就是代码的第8行，这里传递的是`结构对象的Scope`，即在代码第4行创建的`Scope`，而不是`createStructObj`函数中的`scope`参数。原因：结构中的`字段`或者`方法`都是在结构的作用域中进行解释处理的，而不是外部的某个作用域。举个例子：

```c
struct {
	let x = 10
	fn getX() { return self.x }
}
```

代码的第二行的变量`x`和第三行的`getX`函数都是在`结构的作用域`中进行解释处理的。



下一个需要更改的地方是结构方法的调用（`structObj.method(xxx)`）：

```perl
struct math {
    fn Add(x,y) { return x + y }
}

m1 = math()
m1.Add(2,3)  # 结果是5
```

第6行的`m1.Add(2,3)`是结构方法调用。所以我们需要修改`evalMethodCallExpression`方法，新增分支来处理结构类型：

```go
//eval.go
func evalMethodCallExpression(call *ast.MethodCallExpression, scope *Scope) Object {
	//...

	switch m := obj.(type) {
	//...

	case *Struct: //结构
		switch o := call.Call.(type) {
		case *ast.Identifier: //structObj.filed1
			if i, ok := m.Scope.Get(call.Call.String()); ok {
				return i
			}
		case *ast.CallExpression://structObj.method1()
			funcName := o.Function.String()

			if !unicode.IsUpper(rune(funcName[0])) && && str != "self"  {
				return newError(call.Call.Pos().Sline(), ERR_NAMENOTEXPORTED, 
								call.Object.String(), funcName)
			}

			args := evalExpressions(o.Arguments, scope) //解释函数的参数
			if len(args) == 1 && isError(args[0]) {
				return args[0]
			}

			//调用结构对象的'CallMethod'方法
			r := obj.CallMethod(call.Call.Pos().Sline(), scope, funcName, args...)
			return r
		case *ast.IndexExpression: //例如： math.xxx[i] (假设'math'是一个结构变量)
			//left := Eval(o.Left, scope)
			//index := Eval(o.Index, scope)
			//return evalIndexExpression(o, left, index)
			//直接调用Eval方法来处理索引表达式，它会负责调用`evalIndexExpression`方法
			return Eval(o, m.Scope)
		}
	//...
	}

	return newError(call.Call.Pos().Sline(), ERR_NOMETHOD, call.String(), obj.Type())
}


```

代码第8行我们加入了一个判断`结构对象`的分支。接着第9行，我们判断结构对象后面跟着的是一个标识符还是函数调用，如果是标识符（第10行），则从结构的`Scope`中取出这个变量的值并返回。如果是一个函数调用（第14行），则调用`结构对象`的`CallMethod`方法（第28行）。同时用户的脚本代码中也可能会有类似`math.arr[i]`这样的访问例子，因此第30行的`case`分支是用来处理这种情况的。

下面来看一个例子：

```perl
struct math
{
	let Name = "MATH"
	let arr = [1,2,3]
    fn Add(x,y) { return x + y }
	fn Demo() { println(self.arr[2]) } #这里的'self.Arr[2]'就是索引表达式
}

m1 = math()
printf("name=%s\n", m1.Name)   # 这里m1.Name中的Name就是一个Identifier
printf("2 + 3 = %d\n", m1.Add(2,3)) # 这里m1.Add()中的Add()就是一个函数调用
m1.Demo() # 这里会打印'3'
```

第17行用来判断函数名的首字符是否是大写的，如果不是大写的，且不是以`self.field1`或者`self.method1()`的方式调用的则报错。具体来说就是在`结构对象`的内部，结构的方法之间是可以互相调用的，且不区分大小写的：

```perl
struct math
{
    fn add(x,y) { return x + y }
    fn Add(x,y) {
       return self.add(x,y)
    }
}

m1 = math()
println(m1.Add(2,3)) # 正确
println(m1.add(2,3)) # 错误
```

在`Add()`方法的内部，我们调用了结构的另一个方法`add()`（第5行）。虽然这个`add()`的首字母是小写的，但是程序没有问题（只不过你不能把这个add()方法给结构外部使用）。代码第11行，我们在结构对象的外部通过`m1.add()`的方式调用，这个是会报告错误的：

```
Runtime Error at <examples/struct4.mp:11>
        cannot refer to unexported name m1.add
```

有的读者可能会说了，你这个判断有问题。例如下面的代码：
```perl
struct math
{
    fn add(x,y) { return x + y }
    fn Add(x,y) {
       return self.add(x,y)
    }
}

self = math() #将结果赋值给self变量
println(self.Add(2,3))
println(self.add(2,3))
```

上面的代码确实不会报错。好吧，我承认！！那怎么解决这个问题呢？你可以在语言的说明文档中明确的记载：
    【不允许`self`作为变量名】，或者
    【`self`是将来预留的保留字(Reserved keywords)】（我承认我说谎了:smile:）

或者更直接点：

【`self`只能在结构体内使用，不能在任何其它地方使用，且不能被赋值】

吹毛求疵的读者可能会说，这样不太好吧，怎么用代码来防止这种情况出现呢？其实若真的要防止`self`被赋值，我们可以在`evalAssignExpression`方法中，加一个简单的判断即可：

```go
//errors.go
var（
	ERR_SELFASSIGN = "'self' cannot be assigned"
）

//eval.go
func evalAssignExpression(a *ast.AssignExpression, scope *Scope) Object {
	if a.Name.String() == "self" {
		return newError(a.Pos().Sline(), ERR_SELFASSIGN)
	}
	//....
}
```

如果读者再次运行上面的例子，就会报告如下运行期错误：

```
Runtime Error at <examples/struct4.mp:9>
        'self' cannot be assigned
```

甚至这个错误，在语法解析阶段就可以判断出来：

```go
//parser.go
func (p *Parser) parseAssignExpression(name ast.Expression) ast.Expression {
	if name.String() == "self" {
		//报错
	}
	//...
}
```

具体采用哪种方法都可以。这里我们暂且采用第二种方法（即在语法解析阶段）。

这里的解决方法还不太完善，如果读者写的代码如下：

```javascript
a, self, c = 1, true, "hello"
let a, self, c = 1, true, "hello"
```

将`self`放到`let`语句中赋值。为了防止这种代码出现，我们需要修改`parseLetStatement`函数：

```go
//parser.go
func (p *Parser) parseLetStatement(nextFlag bool) *ast.LetStatement {
	//...

	//name部分
	for {
		//...
		name := &ast.Identifier{Token: p.curToken, Value: p.curToken.Literal}
		if p.curToken.Literal == "self" { //标识符的字面量如果是"self"则报错
			msg := fmt.Sprintf("Syntax Error:%v- 'self' can not be assigned", 
								p.curToken.Pos)
			p.errors = append(p.errors, msg)
			p.errorLines = append(p.errorLines, p.curToken.Pos.Sline())
			return nil
		}
		stmt.Names = append(stmt.Names, name)
		//...
	}

	//...
}
```

代码第9行的`if`判断是新增的内容。

有了上面的对于解析器的更改，我们再运行上面对`self`赋值的例子，结果如下（这次报告的是语法错误）：

```
Syntax Error: <examples/struct4.mp:9:7> - 'self' can not be assigned
```




我们还可以给结构变量赋值，就像下面这样（第8-9行）：

```perl
struct math {
    fn Add() {
        return self.x + self.y
    }
}

m1 = math()
m1.x = 2  # 给结构变量赋值
m1.y = 3  # 给结构变量赋值
println(m1.Add())  # 结果是5
```

所以，我们还需要修改`evalAssignExpression`函数：

```go
//eval.go
func evalAssignExpression(a *ast.AssignExpression, scope *Scope) Object {
	val := Eval(a.Value, scope)
	if val.Type() == ERROR_OBJ {
		return val
	}

	if strings.Contains(a.Name.String(), ".") {
		switch o := a.Name.(type) {
		case *ast.MethodCallExpression: //如果是方法调用，例如：xxx.name = 10
			obj := Eval(o.Object, scope)
			if obj.Type() == ERROR_OBJ {
				return obj
			}
			switch m := obj.(type) {
			case *Struct: //如果是结构对象, 例如: structObj.name = 10
				switch c := o.Call.(type) {
				case *ast.Identifier:
					m.Scope.Set(c.Value, val) //将值放入结构对象的Scope中
					return val
				case *ast.IndexExpression: //如果是索引表达式，例如：structObj.xxx[idx]
					var left Object
					var ok bool

					name := c.Left.(*ast.Identifier).Value //取得structObj.xxx[idx]中的'xxx'
					if left, ok = m.Scope.Get(name); !ok { //判断变量'xxx'是否存在于结构的scope中(即m.Scope)
						return newError(a.Pos().Sline(), ERR_UNKNOWNIDENT, name)
					}

					//为了调用原有的写好的函数，这里我们构造一个新的'AssignExpression'表达式,
					//其中的Name即为'strctObj.xxx[idx]'中的'xxx[idx]'这样的索引表达式
					//注意：下面调用的函数是在结构的scope中解释的（即m.Scope）
					b := &ast.AssignExpression{Token: a.Token, Name: c}
					switch left.Type() { //判断'structObj.xxx[idx]'中的'xxx'变量的类型
					case STRING_OBJ:
						return evalStrAssignExpression(b, name, left, m.Scope, val)
					case ARRAY_OBJ:
						return evalArrayAssignExpression(b, name, left, m.Scope, val)
					case TUPLE_OBJ:
						return evalTupleAssignExpression(b, name, left, m.Scope, val)
					case HASH_OBJ:
						return evalHashAssignExpression(b, name, left, m.Scope, val)
					}
				default:
					//error
				}
			} //end inner switch

		} //end outer switch
	} //end if

	//...
}
```

第8行的`if`判断是新追加的。我在代码中加入了足够多的注释，以便于读者理解。

最后，还有一点需要考虑到。我们的语言是支持`import`的，如果有下面的代码：

```perl
# calc.mp
struct Math {
    # 构造函数
    fn init(x, y) {
        self.x = x
        self.y = y
    }

    # unexported
    fn add() { return self.x + self.y }
    fn Add() { return self.add() }
    fn Sub() { return self.x - self.y }
}

# import.mp
import sub_package.calc

m = Math(2,3) # 调用calc.mp文件中的Math结构的构造函数
printf("Math.add()=%d\n", m.Add())
```

第18行我们调用了导入模块`calc`的`Math结构`的构造方法，第19行调用了结构对象的`Add`方法。

本节前面提到，我们给`Scope`结构增加了一个`structStore`变量，用来保存结构信息。就是说Scope中的`structStore`专门用来存储结构信息。因此当你在18行调用`Math`结构的构造函数的时候，这个`Math`结构在Scope的`store`变量中是不存在的。因此第18行会报告变量`Math`找不到的错误：

```sh
unknown identifier: 'Math' is not defined
```

为了解决这个问题，我们需要修改`GetAllExported`这个函数：

```go
//scope.go
func (s *Scope) GetAllExported(anotherScope *Scope) {
	for key, value := range s.store {
		if unicode.IsUpper(rune(key[0])) { //仅大写字母开头的函数/变量被导出
			anotherScope.Set(key, value)
		}
	}

	for key, value := range s.structStore {
		if unicode.IsUpper(rune(key[0])) { //仅大写字母开头的结构被导出
			anotherScope.SetStruct(value)
		}
	}
}
```

第9-13行是新增的代码。这里我们把`scope`自身保存的所有的结构信息拷贝到当前`anotherScope`中（只拷贝大写字母开头的结构）。



## 测试代码

```perl
struct Person { # person结构
    fn init(name, score) { # 构造函数
        self.name = name
        self.score = score
    }

    fn GetInfo() {
        return self.name, self.score
    }
}

struct Class { # 班级结构
    fn init(name) {  # 构造函数
        self.clsName = name
    }

    fn SetPersons(persons) {
        self.persons = persons
    }

    fn GetClassInfo() {
        printf("%s:\n", self.clsName)
        for person in self.persons {
            name, score = person.GetInfo()
            printf("\tname = %s, score = %.2f\n", name, score)
        } 
    }
}


cls = Class("五年一班")
cls.SetPersons([Person("小李", 89.2), Person("小王", 93.5), Person("小张", 70.8)])
cls.GetClassInfo()

```

运行结果：

```
五年一班:
        name = 小李, score = 89.20
        name = 小王, score = 93.50
        name = 小张, score = 70.80
```

这篇非常有挑战性的文章终于完成了。



下一节，我们将加入`switch-case`语句的支持。




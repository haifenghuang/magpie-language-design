# for循环支持

在这一节中，我们要加入对于`for`循环的支持。

对于循环，我们需要有控制循环退出或者继续的机制，这是大家都熟悉的`break`和`continue`。

所以在讲解`for`循环的支持前，让我们先加入对`break`和`continue`的支持。

## `Break`和`Continue`支持

对于`Break`和`Continue`，修改的内容都是我们熟悉的，所以这里直接就上代码，我会在需要的地方写上注释。

### 词元（Token）的更改

```go
//token.go
const (
	//...
	TOKEN_BREAK    //break
	TOKEN_CONTINUE //continue
)

//词元类型的字符串表示
func (tt TokenType) String() string {
	switch tt {
	//...
	case TOKEN_BREAK:
		return "BREAK"
	case TOKEN_CONTINUE:
		return "CONTINUE"
	}
}

var keywords = map[string]TokenType{
	//...
	"break":    TOKEN_BREAK,
	"continue": TOKEN_CONTINUE,
}
```

### 抽象语法树（AST）的更改

```go
//ast.go
// BREAK
type BreakExpression struct {
	Token token.Token
}

func (be *BreakExpression) Pos() token.Position {
	return be.Token.Pos
}

func (be *BreakExpression) End() token.Position {
	length := utf8.RuneCountInString(be.Token.Literal)
	pos := be.Token.Pos
	return token.Position{Filename: pos.Filename, Line: pos.Line, Col: pos.Col + length}
}

func (be *BreakExpression) expressionNode()      {}
func (be *BreakExpression) TokenLiteral() string { return be.Token.Literal }

func (be *BreakExpression) String() string { return be.Token.Literal }


// CONTINUE
type ContinueExpression struct {
	Token token.Token
}

func (ce *ContinueExpression) Pos() token.Position {
	return ce.Token.Pos
}

func (ce *ContinueExpression) End() token.Position {
	length := utf8.RuneCountInString(ce.Token.Literal)
	pos := ce.Token.Pos
	return token.Position{Filename: pos.Filename, Line: pos.Line, Col: pos.Col + length}
}

func (ce *ContinueExpression) expressionNode()      {}
func (ce *ContinueExpression) TokenLiteral() string { return ce.Token.Literal }

func (ce *ContinueExpression) String() string { return ce.Token.Literal }
```

### 语法解析器（Parser）的更改

对于`break`和`continue`，我们是不允许在循环外部使用的。那么怎么防止这种情况呢？我们可以在语法解析阶段防止这种情况，也可以在解释阶段（Evaluating phase）防止这种情况。这里我们采用第一种方式，因为在语法解析阶段实现起来相对来说容易一些。

让我们来看一个循环的例子：

```javascript
let arr = [1,2,3]
for item in arr {
	println(item)

	let arr_sub = [4,5,6]
	for item_sub in arr {
		println(item_sub)	
	}
}
```

对于上面的例子，我们可以使用一个变量`loopDepth（默认为0）`，当进入循环的时候，我们将这个变量加1。当退出循环的时候，将这个变量减1。这样，当我们遇到`break`或者`continue`的时候，我们首先检查这个变量的值，如果值为0，说明不在循环中，这时候就可以报错。还是以上面的代码为例子（伪代码）：

```javascript
loopDepth = 0

let arr = [1,2,3]
loopDepth++
for item in arr {
	println(item)

	let arr_sub = [4,5,6]
	loopDepth++
	for item_sub in arr {
		println(item_sub)	
	}
	loopDepth--
}
loopDepth--

```

首先我们设置`loopDepth`变量的值伪0（第1行）。当我们进入第5行的外层for循环前，我们将`    loopDepth`的值加1。当退出外层的for循环后，我们将这个`loopDepth`的值减1（第15行）。同样的道理，进入内层的for循环前，我们继续将`    loopDepth`的值加1（第9行），退出内层的for循环后，将这个`loopDepth`的值减1（第13行）。

这样下来，只有在for循环里面的时候，这个`loopDepth`变量的值才会大于0。当`loopDepth`变量的值为0的时候，说明我们已经退出了循环（或者说不在循环语句中）。

有了上面的说明，让我们看一下实现：

```go
//parser.go

type Parser struct {
	//...
	loopDepth int // 当前循环深度(depth), 0表示不在循环中
}

func (p *Parser) registerAction() {
	//...
	//注册break和continue词元的前缀表达式回调函数
	p.registerPrefix(token.TOKEN_BREAK, p.parseBreakExpression)
	p.registerPrefix(token.TOKEN_CONTINUE, p.parseContinueExpression)

	//...
}

func (p *Parser) parseBreakExpression() ast.Expression {
	if p.loopDepth == 0 { //如果loopDepth变量值为0，说明不在循环中，则报错
		msg := fmt.Sprintf("Syntax Error:%v- 'break' outside of loop context", 
				p.curToken.Pos)
		p.errors = append(p.errors, msg)
		p.errorLines = append(p.errorLines, p.curToken.Pos.Sline())

		return nil
	}

	return &ast.BreakExpression{Token: p.curToken}

}

func (p *Parser) parseContinueExpression() ast.Expression {
	if p.loopDepth == 0 { //如果loopDepth变量值为0，说明不在循环中，则报错
		msg := fmt.Sprintf("Syntax Error:%v- 'continue' outside of loop context", 
				p.curToken.Pos)
		p.errors = append(p.errors, msg)
		p.errorLines = append(p.errorLines, p.curToken.Pos.Sline())

		return nil
	}

	return &ast.ContinueExpression{Token: p.curToken}
}
```



### 对象系统（Object System）的更改

```go
//object.go
const (
	//...
	BREAK_OBJ        = "BREAK"
	CONTINUE_OBJ     = "CONTINUE"
)

var (
	//...
	BREAK    = &Break{}    //所有的break都是一样的，所以这里我们定义了一个唯一的`Break对象`
	CONTINUE = &Continue{} //所有的continue都是一样的，所以这里我们定义了一个唯一的`Continue对象`
)

//Break对象
type Break struct{}

func (b *Break) Inspect() string  { return "break" }
func (b *Break) Type() ObjectType { return BREAK_OBJ }
func (b *Break) CallMethod(line string, scope *Scope, method string, args ...Object) Object {
	return newError(line, ERR_NOMETHOD, method, b.Type())
}

//Continue对象
type Continue struct{}

func (c *Continue)Inspect() string  { return "continue" }
func (c *Continue)Type() ObjectType { return CONTINUE_OBJ }
func (c *Continue)CallMethod(line string,scope *Scope,method string,args ...Object)Object {
	return newError(line, ERR_NOMETHOD, method, c.Type())
}
```



### 解释器（Evaluator）的更改

```go
//eavl.go
func Eval(node ast.Node, scope *Scope) (val Object) {
	switch node := node.(type) {
	//...
	case *ast.BreakExpression:
		return BREAK
	case *ast.ContinueExpression:
		return CONTINUE
	//...
	}

	return nil
}

func evalBlockStatement(block *ast.BlockStatement, scope *Scope) Object {
	var result Object
	for _, statement := range block.Statements {
		result = Eval(statement, scope)
		//...
		if _, ok := result.(*Break); ok {
			return result
		}
		if _, ok := result.(*Continue); ok {
			return result
		}
	}
	return result
}
```

我们首先在`Eval()`函数中加入了判断`break`和`continue`的分支（5-8行）。

其次，在`evalBlockStatement()`函数中，在for循环中处理每一个`Statement(语句)`的时候，如果遇到`break`或者`continue`也会返回相应的`Break`和`Continue`对象（20-25行）。有的读者会问，为什么要在这个`evalBlockStatement`函数中处理`break`和`continue`?不应该在`for循环`的内部处理中处理这个`break`和`continue`吗？我们来看一下下面这个例子：

```javascript
for i in [1,2,3] {
	if i == 2 {
		continue  //或者'break'
	}
	println(i)
}
```

当解释器遇到`if表达式`的时候，它就会调用`evalIfExpression`函数。我们再来看一下它的实现：

```go
//eval.go
func evalIfExpression(ie *ast.IfExpression, scope *Scope) Object {
	for _, c := range ie.Conditions {
		condition := Eval(c.Cond, scope)
		if condition.Type() == ERROR_OBJ {
			return condition
		}

		if IsTrue(condition) {
			return evalBlockStatement(c.Body, scope)
		}
	}

	//eval "else" part
	if ie.Alternative != nil {
		return evalBlockStatement(ie.Alternative, scope)
	}

	return NIL
}
```

注意第9行的判断，如果条件为真的话，它就会调用`evalBlockStatement`函数处理`if表达式`中的块语句(Block Statement)。对于这个`evalBlockStatemnt`函数，解释器是区分不出来到底是在处理`for循环`的块语句还是在处理`if表达式`的块语句。但是这无关紧要，只要是在块循环中，我们遇到`break`或`continue`的时候都会进行处理。有的读者就会问了，那要是在这里面处理`break`或者`continue`的话，下面的例子不会出错吗？

```go
i = 10
{
	if i == 10 { continue }
}
```

上面的例子当然会出错，而且是在语法解析阶段。前面我们已经分析过，对于这种情况，`语法解析器（Parser）`会捕捉到`continue`不在循环中，会直接报告如下错误：

```
"Syntax Error:<3:17>- 'continue' outside of loop context
```

讲解完`break`和`continue`后，我们来进入正题。



## `for循环`支持

先来看一下我们语言的`for`的例子，让读者有个初步的了解：

```javascript
//for var in value
for var in value { block } //这里value可以是数组，元祖和字符串

//for k,v in X
for k, v in X { block } //这里X可以是数组，元祖，哈希和字符串。对于哈希以外的类型，这个key存放的是索引值
for _, v in X { block }
for k, _ in X { block }
for _, _ in X { block } //错误，不允许

//无限循环loop
for { block }

//c's for-loop
i = 0
for (i = 0; i < 5; i++) { block }  //类似c语言的for循环， '()'必须要有

i = 0
for (; i < 5; i++) { block }  // 无初期化语句

i = 0
for (; i < 5;;) { block } // 无初期化和更新语句

i = 0
for (;;;) { block } // 等价于'for { block }'语句，即无限循环
```

从上面可以看到，总共有四种类型的循环

* for var in value
* for k, v in X
* for {}无限循环
* 类似c语言的for循环



下面来看一下需要更改的内容。



### 词元（Token）的更改

我们增加了两个新的关键字：`for`和`in`。

```go
//token.go
package token

import (
	"fmt"
)

// token
type TokenType int

const (
	//...
	TOKEN_FOR      //for
	TOKEN_IN       //in
)

//词元类型的字符串表示
func (tt TokenType) String() string {
	switch tt {

	//...
	case TOKEN_FOR:
		return "FOR"
	case TOKEN_IN:
		return "IN"
	}
}

var keywords = map[string]TokenType{
	//...
	"for":      TOKEN_FOR,
	"in":       TOKEN_IN,
}
```



### 抽象语法树的（AST）更改

#### C语言的for loop

我们先来看一下`c`语言的for循环：

```c
//c for loop
int i = 0;
for (i = 0; i < 10; i++) { block}
```

如果我们抽象一下的话，可以得到如下的表示：

```c
for (init; condition; updater) { block }
```

其中包含三个部分，即：

* 初始化部分(init)
* 条件部分(condition)
* 条件更新部分(updater)

这三个部分的任何一个都可以省略。甚至三个全部都可以省略，变成如下的形式：

```c
for(;;;) { block }
```

有了上面的分析，我们的`c`循环的抽象语法表示就比较清楚了：

```go
//ast.go

//c language like for loop
type CForLoop struct {
	Token  token.Token      //'for' token
	Init   Expression       //初始化部分（可为nil）
	Cond   Expression       //条件部分(可为nil)
	Update Expression       //条件更新部分(可为nil)
	Block  *BlockStatement  //块语句
}

func (fl *CForLoop) Pos() token.Position {
	return fl.Token.Pos
}

func (fl *CForLoop) End() token.Position {
	return fl.Block.End()
}

func (fl *CForLoop) expressionNode()      {}
func (fl *CForLoop) TokenLiteral() string { return fl.Token.Literal }

func (fl *CForLoop) String() string {
	var out bytes.Buffer

	out.WriteString("for")
	out.WriteString(" ( ")

	if fl.Init != nil {
		out.WriteString(fl.Init.String())
	}
	out.WriteString(" ; ")

	if fl.Cond != nil {
		out.WriteString(fl.Cond.String())
	}
	out.WriteString(" ; ")

	if fl.Update != nil {
		out.WriteString(fl.Update.String())
	}
	out.WriteString(" ) ")
	out.WriteString(" { ")
	out.WriteString(fl.Block.String())
	out.WriteString(" }")

	return out.String()
}
```

内容还是比较简单的，没什么需要太多说明的。



#### `for var in value`循环

下面我们来看`for var in value`这个循环。它的格式如下：

```perl
for var in value { block }
```

它的抽象语法表示也比较简单：

```go
//ast.go

//for variable in value { block }
type ForEachArrayLoop struct {
	Token token.Token
	Var   string     //变量
	Value Expression //可遍历的值（如数组，元祖及字符串）
	Block *BlockStatement //块语句
}

func (fal *ForEachArrayLoop) Pos() token.Position {
	return fal.Token.Pos
}

func (fal *ForEachArrayLoop) End() token.Position {
	return fal.Block.End()
}

func (fal *ForEachArrayLoop) expressionNode()      {}
func (fal *ForEachArrayLoop) TokenLiteral() string { return fal.Token.Literal }

func (fal *ForEachArrayLoop) String() string {
	var out bytes.Buffer

	out.WriteString("for ")
	out.WriteString(fal.Var)
	out.WriteString(" in ")
	out.WriteString(fal.Value.String())
	out.WriteString(" { ")
	out.WriteString(fal.Block.String())
	out.WriteString(" }")

	return out.String()
}
```



#### `for k, v in X`循环

接下来我们来看一下`for k,v in X`这个循环。它的格式如下：

```perl
for k, v in X { block }
```

下面是它的抽象语法表示：

```go
//ast.go

//for k, v in X { block }
type ForEachMapLoop struct {
	Token token.Token
	Key   string          //key
	Value string          //value
	X     Expression      //可遍历的值（如数组，元祖，哈希及字符串）
	Block *BlockStatement //块语句
}

func (fml *ForEachMapLoop) Pos() token.Position {
	return fml.Token.Pos
}

func (fml *ForEachMapLoop) End() token.Position {
	return fml.Block.End()
}

func (fml *ForEachMapLoop) expressionNode()      {}
func (fml *ForEachMapLoop) TokenLiteral() string { return fml.Token.Literal }

func (fml *ForEachMapLoop) String() string {
	var out bytes.Buffer

	out.WriteString("for ")
	out.WriteString(fml.Key + ", " + fml.Value)
	out.WriteString(" in ")
	out.WriteString(fml.X.String())
	out.WriteString(" { ")
	out.WriteString(fml.Block.String())
	out.WriteString(" }")

	return out.String()
}
```



#### `for {}`无限循环

它的格式就更简单了：

```go
for { block }
```

它的抽象语法表示如下：

```go
//ast.go

//for { block }
type ForEverLoop struct {
	Token token.Token
	Block *BlockStatement
}

func (fel *ForEverLoop) Pos() token.Position {
	return fel.Token.Pos
}

func (fel *ForEverLoop) End() token.Position {
	return fel.Block.End()
}

func (fel *ForEverLoop) expressionNode()      {}
func (fel *ForEverLoop) TokenLiteral() string { return fel.Token.Literal }

func (fel *ForEverLoop) String() string {
	var out bytes.Buffer

	out.WriteString("for ")
	out.WriteString(" { ")
	out.WriteString(fel.Block.String())
	out.WriteString(" }")

	return out.String()
}
```



### 语法解析器（Parser）的更改

首先我们需要对`for`关键字注册一个前缀表达式回调函数：

```go
//parser.go
func (p *Parser) registerAction() {
	//...
	p.registerPrefix(token.TOKEN_FOR, p.parseForLoopExpression)

}
```

我们给`TOKEN_FOR`词元类型注册了一个前缀表达式回调函数。接下来看一下这个`parseForLoopExpression`函数的实现：

```go
//parser.go
//处理for循环
func (p *Parser) parseForLoopExpression() ast.Expression {
	p.loopDepth++
	curToken := p.curToken //保存当前词元

	var r ast.Expression
	if p.peekTokenIs(token.TOKEN_LBRACE) { //for { block }
		r = p.parseForEverLoopExpression(curToken)
		p.loopDepth--
		return r
	}

	if p.peekTokenIs(token.TOKEN_LPAREN) { //for (init; cond; updater) { block }
		r = p.parseCForLoopExpression(curToken)
		p.loopDepth--
		return r
	}

	p.nextToken()                             //skip 'for'
	if p.curToken.Literal == "_" { //for _, value in xxx { block }
		r = p.parseForEachMapExpression(curToken, p.curToken.Literal)
	} else if p.curTokenIs(token.TOKEN_IDENTIFIER) {
        if p.peekTokenIs(token.TOKEN_COMMA) { //for k, value in xxx { block }
			r = p.parseForEachMapExpression(curToken, p.curToken.Literal)
        } else { //for item in xxx { block }
			r = p.parseForEachArrayExpression(curToken, p.curToken.Literal)
		}
	} else {
		msg := fmt.Sprintf("Syntax Error:%v- for loop must be followed by an underscore or identifier. got %s", p.curToken.Pos, p.curToken.Literal)
		p.errors = append(p.errors, msg)
		p.errorLines = append(p.errorLines, p.curToken.Pos.Sline())
		return nil
	}

	p.loopDepth--
	return r
}

//for (init; condition; update) {}
//for (; condition; update) {}  --- init is empty
//for (; condition;;) {}  --- init & update both empty
// for (;;;) {} --- init/condition/update all empty
func (p *Parser) parseCForLoopExpression(curToken token.Token) ast.Expression {
	var result ast.Expression

	if !p.expectPeek(token.TOKEN_LPAREN) {
		return nil
	}

	var init ast.Expression
	var cond ast.Expression
	var update ast.Expression

	//init部分
	p.nextToken()
	if !p.curTokenIs(token.TOKEN_SEMICOLON) {
		init = p.parseExpression(LOWEST)
		p.nextToken()
	}

	//condition部分
	p.nextToken() //skip ';'
	if !p.curTokenIs(token.TOKEN_SEMICOLON) {
		cond = p.parseExpression(LOWEST)
		p.nextToken()
	}

	//update部分
	p.nextToken()
	if !p.curTokenIs(token.TOKEN_SEMICOLON) {
		update = p.parseExpression(LOWEST)
	}

	if !p.expectPeek(token.TOKEN_RPAREN) {
		return nil
	}

	if !p.peekTokenIs(token.TOKEN_LBRACE) {
		msg := fmt.Sprintf("Syntax Error:%v- for loop must be followed by a '{'.", p.curToken.Pos)
		p.errors = append(p.errors, msg)
		p.errorLines = append(p.errorLines, p.curToken.Pos.Sline())
		return nil
	}

	p.nextToken()

	//如果三个部分都是nil的化，我们将其当作无线循环来处理
	if init == nil && cond == nil && update == nil {
		loop := &ast.ForEverLoop{Token: curToken}
		loop.Block = p.parseBlockStatement()
		result = loop
	} else {
		loop := &ast.CForLoop{Token: curToken, Init: init, Cond: cond, Update: update}
		loop.Block = p.parseBlockStatement()
		result = loop
	}

	return result
}

//for item in xxx {}
func (p *Parser) parseForEachArrayExpression(curToken token.Token, variable string) ast.Expression {
	if !p.expectPeek(token.TOKEN_IN) {
		return nil
	}
	p.nextToken()

	value := p.parseExpression(LOWEST)

	var block *ast.BlockStatement
	if p.peekTokenIs(token.TOKEN_LBRACE) {
		p.nextToken()
		block = p.parseBlockStatement()
	} else {
		msg := fmt.Sprintf("Syntax Error:%v- for loop must be followed by a '{' ", p.curToken.Pos)
		p.errors = append(p.errors, msg)
		p.errorLines = append(p.errorLines, p.curToken.Pos.Sline())
		return nil
	}

	result := &ast.ForEachArrayLoop{Token: curToken, Var: variable, 
									Value: value, Block: block}
	return result
}

//for key, value in xxx {}
//key和value的任意一个都可以是"_"，但不能都是"_"
func (p *Parser) parseForEachMapExpression(curToken token.Token, key string) ast.Expression {
	loop := &ast.ForEachMapLoop{Token: curToken}
	loop.Key = key

	if !p.expectPeek(token.TOKEN_COMMA) {
		return nil
	}

	p.nextToken() //skip ','
	if p.curToken.Literal == "_" {
		//do nothing
	} else if !p.curTokenIs(token.TOKEN_IDENTIFIER) {
		msg := fmt.Sprintf("Syntax Error:%v- for loop must be followed by an identifier. got %s",
                           p.curToken.Pos, p.curToken.Literal)
		p.errors = append(p.errors, msg)
		p.errorLines = append(p.errorLines, p.curToken.Pos.Sline())
		return nil
	}
	loop.Value = p.curToken.Literal

    //key, value全部是"_"就报错
	if loop.Key == "_" && loop.Value == "_" { //for _, _ in xxx { block }
		msg := fmt.Sprintf("Syntax Error:%v- foreach map's key & map are both '_'", p.curToken.Pos)
		p.errors = append(p.errors, msg)
		p.errorLines = append(p.errorLines, p.curToken.Pos.Sline())
		return nil
	}

	if !p.expectPeek(token.TOKEN_IN) {
		return nil
	}

	p.nextToken()
	loop.X = p.parseExpression(LOWEST)

	if p.peekTokenIs(token.TOKEN_LBRACE) {
		p.nextToken()
		loop.Block = p.parseBlockStatement()
	} else {
		msg := fmt.Sprintf("Syntax Error:%v- for loop must be followed by a '{'.", p.curToken.Pos)
		p.errors = append(p.errors, msg)
		p.errorLines = append(p.errorLines, p.curToken.Pos.Sline())
		return nil
	}

	return loop
}

//for { block }无限循环
func (p *Parser) parseForEverLoopExpression(curToken token.Token) ast.Expression {
	loop := &ast.ForEverLoop{Token: curToken}

	p.expectPeek(token.TOKEN_LBRACE)
	loop.Block = p.parseBlockStatement()

	return loop
}
```

代码虽然很长，但是也不是很难懂。在`parseForLoopExpression`函数的开始处，我们将`loopDepth`变量加1， 在`return`前将`loopDepth`变量减1。这样，只要在loop循环中，这个`loopDepth`变量的值就大于1。如果这个`loopDepth`的值为0，则表明我们已经退出了循环，或者我们根本就不在循环中。

`parseForLoopExpression`函数的另一个功能是：根据遇到的不同词元类型来判断到底是上面说的四种`for`循环类型的哪一种，然后调用不同的方法来处理。



### 解释器(Evaluator)的更改

同样，我们直接来看代码：

```go
//eval.go
func Eval(node ast.Node, scope *Scope) (val Object) {
	switch node := node.(type) {
	//...
	case *ast.CForLoop:  //处理C的for循环: for (init; condition; updater) { block }
		return evalCForLoopExpression(node, scope)
	case *ast.ForEverLoop: //处理无限循环: for { block }
		return evalForEverLoopExpression(node, scope)
	case *ast.ForEachArrayLoop: //处理: for var in Value { block }
		return evalForEachArrayExpression(node, scope)
	case *ast.ForEachMapLoop: //处理：for k, v in X { block }
		return evalForEachMapExpression(node, scope)
	}

	return nil
}
```

下面我们分别来看一下上面的四个函数的实现。

#### `evalCForLoopExpression`函数

这个是处理C的for循环的函数。

```go
//eval.go
//for (init; condition; update) { block }
// 返回最后一个表达式的值或者NIL
func evalCForLoopExpression(fl *ast.CForLoop, scope *Scope) Object { //fl:For Loop
	if fl.Init != nil {
		init := Eval(fl.Init, scope)
		if init.Type() == ERROR_OBJ {
			return init
		}
	}

	var result Object = NIL
	for {
		//condition
		var condition Object = NIL
		if fl.Cond != nil {
			condition = Eval(fl.Cond, scope)
			if condition.Type() == ERROR_OBJ {
				return condition
			}
			if !IsTrue(condition) {
				break
			}
		}

		//body
		result = Eval(fl.Block, scope)
		if result.Type() == ERROR_OBJ {
			return result
		}

		if _, ok := result.(*Break); ok { //如果是break
			break
		}
		if _, ok := result.(*Continue); ok { //如果是continue
			if fl.Update != nil {
				newVal := Eval(fl.Update, scope) //继续之前，我们需要调用一下'Update'
				if newVal.Type() == ERROR_OBJ {
					return newVal
				}
			}

			continue
		}
		if v, ok := result.(*ReturnValue); ok { //如果是return
			return v
		}

		if fl.Update != nil {
			newVal := Eval(fl.Update, scope)
			if newVal.Type() == ERROR_OBJ {
				return newVal
			}
		}
	}

	if result == nil || result.Type() == BREAK_OBJ || result.Type() == CONTINUE_OBJ {
		return NIL
	}

	return result
}
```

代码中我给出了比较详细的说明。其中需要关注的是35-44行的处理：遇到`continue`语句，我们需要更新`update`部分(36-41行)，然后再`continue`(43行)。

还有需要注意的就是57-59行的判断语句。为啥要有这个判断呢？举个例子：

```c
i = 0
for (; i < 5; i++) {  // 无初期化语句
    if (i > 4) { break }
    if (i == 2) { continue }
    println("i=", i)
}
```

因为我们的语言中`for`是个表达式，可以返回值。它的返回值为最后一个执行的表达式的值。从这个例子中我们可以知道，执行的最后一个表达式是第3行的`if`判断中的`break`。那么如果我们把上面的`for`循环的返回值赋给一个变量，如下所示：

```c
i = 0;
x = for (; i < 5; i++) {  # 无初期化语句
    if (i > 4) { break }
    if (i == 2) { continue }
    println("i=", i)
}
println(x)
```

那么`x`变量中存放的就是`BREAK`对象。第7行就会打印`break`这几个字符，这当然不是我们期望的。这种情况下我们返回NIL对象。

当然，这种用法一般是不会将其`for`循环的返回值赋给变量的，而只是想执行一段`for`代码。



#### `evalForEverLoopExpression`函数

这个函数处理`for { block }`无限循环，代码非常简单：

```go
//eval.go
// for { block }
// 返回最后一个表达式的值或者NIL
func evalForEverLoopExpression(fel *ast.ForEverLoop, scope *Scope) Object {
	var e Object = NIL
	for {
		e = Eval(fel.Block, scope)
		if e.Type() == ERROR_OBJ {
			return e
		}

		if _, ok := e.(*Break); ok {
			break
		}
		if _, ok := e.(*Continue); ok {
			continue
		}
		if v, ok := e.(*ReturnValue); ok {
			return v
		}
	}

	if e == nil || e.Type() == BREAK_OBJ || e.Type() == CONTINUE_OBJ {
		return NIL
	}

	return e
}
```

23-25行的`if`判断和上面是一样的逻辑。



#### `evalForEachArrayExpression`函数

这个函数处理`for var in Value { block }`

```go
//eval.go
//for item in array
//for item in string
//for item in tuple
//返回数组对象或者是返回值对象(Return-object)
func evalForEachArrayExpression(fal *ast.ForEachArrayLoop, scope *Scope) Object { //fal:For Array Loop
	aValue := Eval(fal.Value, scope)
	if aValue.Type() == ERROR_OBJ {
		return &Array{Members: []Object{aValue}}
	}

	if aValue.Type() == NIL_OBJ {
		return &Array{Members: []Object{}} //如果是NIL的话，返回空数组
	}

	//判断是否可遍历
	iterObj, ok := aValue.(Iterable)
	if !ok {
		errObj := newError(fal.Pos().Sline(), ERR_NOTITERABLE)
		return &Array{Members: []Object{errObj}}
	}
	if !iterObj.iter() {
		errObj := newError(fal.Pos().Sline(), ERR_NOTITERABLE)
		return &Array{Members: []Object{errObj}}
	}

	//判断类型
	var members []Object
	if aValue.Type() == STRING_OBJ { //字符串
		aStr, _ := aValue.(*String)
		runes := []rune(aStr.String)
		for _, rune := range runes {
			members = append(members, NewString(string(rune)))
		}
	} else if aValue.Type() == ARRAY_OBJ { //数组
		arr, _ := aValue.(*Array)
		members = arr.Members
	} else if aValue.Type() == TUPLE_OBJ { //元组
		tuple, _ := aValue.(*Tuple)
		members = tuple.Members
	}

	if len(members) == 0 {
		return &Array{Members: []Object{}} //返回空数组
	}

	arr := &Array{}
	defer func() { //下面52行代码中将变量加入了scope，这里需要删除掉，因为这个变量只在循环内部使用
		scope.Del(fal.Var)
	}()
	for _, value := range members {
		scope.Set(fal.Var, value) //将每次循环中取得的值放入scope中

		result := Eval(fal.Block, scope) //解释块语句
		if result.Type() == ERROR_OBJ {
			arr.Members = append(arr.Members, result)
			return arr
		}

		if _, ok := result.(*Break); ok {
			break
		}
		if _, ok := result.(*Continue); ok {
			continue
		}
		if v, ok := result.(*ReturnValue); ok {
			return v
		} else {
			arr.Members = append(arr.Members, result)
		}
	}

	return arr
}
```

这里需要注意的是17-21行的判断。`for var in Value {block}`这个式子中的Value必须是一个可遍历对象，如果不是的话，我们就会报错。例如：

```perl
for item in 1 {
    # do something 
}
```

这个例子是会报错的。到目前为止，我们支持的可遍历对象有：数组，元祖和字符串和哈希。为了识别某个对象是否的可遍历（iterable），我们新加入了一个`Iterable`接口：

```go
//可遍历接口
type Iterable interface {
	iter() bool //所有可遍历对象必须实现这个方法，并且将这个方法返回true
}

```

同时，数组，元祖，字符串和哈希实现了这个接口，并且`iter()`方法都返回true：

```go
//object.go
func (s *String) iter() bool { return true }
func (a *Array) iter() bool  { return true }
func (h *Hash) iter() bool   { return true }
func (t *Tuple) iter() bool  { return true }
```

举个使用这个`for`循环的例子：

```go
arr = [1, 2, 3]
arr2 = for item in arr { item * 2 }
println(arr2) //打印: [2, 4, 6]
```

这里我们将`for`循环的每个元素的值乘以了2。



#### `evalForEachMapExpression`函数

这个函数处理`for k, v in X { block}`形式。继续来看代码：

```go
//eval.go
//for k, v in X { block }
//返回数组对象或者是返回值对象(Return-object)
func evalForEachMapExpression(fml *ast.ForEachMapLoop, scope *Scope) Object { //fml:For Map Loop
	aValue := Eval(fml.X, scope)
	if aValue.Type() == ERROR_OBJ {
		return &Array{Members: []Object{aValue}}
	}

	if aValue.Type() == NIL_OBJ {
		return &Array{Members: []Object{}} //返回空数组
	}

	//判断对象是否可遍历
	iterObj, ok := aValue.(Iterable)
	if !ok {
		errObj := newError(fml.Pos().Sline(), ERR_NOTITERABLE)
		return &Array{Members: []Object{errObj}}
	}
	if !iterObj.iter() {
		errObj := newError(fml.Pos().Sline(), ERR_NOTITERABLE)
		return &Array{Members: []Object{errObj}}
	}

	//for index, value in arr
	//for index, value in string
	//for index, value in tuple
	if aValue.Type() == STRING_OBJ || aValue.Type() == ARRAY_OBJ || aValue.Type() == TUPLE_OBJ {
		return evalForEachArrayWithIndex(fml, aValue, scope)
	}

	//'for k, v in hash'的情况
	hash, _ := aValue.(*Hash)
	if len(hash.Pairs) == 0 { //hash is empty
		return &Array{Members: []Object{}} //哈希空的话，返回空数组
	}

	arr := &Array{}
	//下面代码的52、55行向scope中加入了两个新的变量（而这两个变量仅在for循环内部使用），
    //因此循环结束后，我们需要将其从scope中删除
	defer func() {
		if fml.Key != "_" {
			scope.Del(fml.Key)
		}
		if fml.Value != "_" {
			scope.Del(fml.Value)
		}
	}()

	for _, pair := range hash.Pairs {
		if fml.Key != "_" {
			scope.Set(fml.Key, pair.Key) //将每次循环中取到的key放入scope中
		}
		if fml.Value != "_" {
			scope.Set(fml.Value, pair.Value) //将每次循环中取到的value放入scope中
		}

		result := Eval(fml.Block, scope)
		if result.Type() == ERROR_OBJ {
			arr.Members = append(arr.Members, result)
			return arr
		}

		if _, ok := result.(*Break); ok {
			break
		}
		if _, ok := result.(*Continue); ok {
			continue
		}
		if v, ok := result.(*ReturnValue); ok {
			return v
		} else {
			arr.Members = append(arr.Members, result)
		}
	}

	return arr
}

//for index, value in string
//for index, value in array
//for index, value in tuple
//返回数组对象或者是返回值对象(Return-object)
func evalForEachArrayWithIndex(fml *ast.ForEachMapLoop, val Object, scope *Scope) Object {
	var members []Object
	if val.Type() == STRING_OBJ { //字符串
		aStr, _ := val.(*String)
		runes := []rune(aStr.String)
		for _, rune := range runes {
			members = append(members, NewString(string(rune)))
		}
	} else if val.Type() == ARRAY_OBJ { //数组
		arr, _ := val.(*Array)
		members = arr.Members
	} else if val.Type() == TUPLE_OBJ { //元组
		tuple, _ := val.(*Tuple)
		members = tuple.Members
	}

	if len(members) == 0 {
		return &Array{Members: []Object{}} //如果数组为空，返回空对象
	}

	arr := &Array{}
	//下面代码的118、112行向scope中加入了两个新的变量（而这两个新的变量仅在for循环中使用），
	//运行完for循环后需要将其从scope中删除
	defer func() {
		if fml.Key != "_" {
			scope.Del(fml.Key)
		}
		if fml.Value != "_" {
			scope.Del(fml.Value)
		}
	}()
	for idx, value := range members {
		if fml.Key != "_" {
			//将每次循环中取到的索引值放入scope中
			scope.Set(fml.Key, NewNumber(float64(idx)))
		}
		if fml.Value != "_" {
			//将每次循环中取到的value放入scope中
			scope.Set(fml.Value, value)
		}

		result := Eval(fml.Block, scope)
		if result.Type() == ERROR_OBJ {
			arr.Members = append(arr.Members, result)
			return arr
		}

		if _, ok := result.(*Break); ok {
			break
		}
		if _, ok := result.(*Continue); ok {
			continue
		}
		if v, ok := result.(*ReturnValue); ok {
			return v
		} else {
			arr.Members = append(arr.Members, result)
		}
	}

	return arr
}
```

本章的代码比较多，但是相对来说理解起来不算困难。当然也希望读者好好理解其中的逻辑。



## 测试

终于到了测试环节了。

```go
//main.go

func TestEval() {
	tests := []struct {
		input    string
		expected string
	}{
		//array, tuple, string
		{`for item in [1, true, "Hello"] { println(item) } println()`, "nil"},
		{`for idx,v in [1,true,"Hello"]{if idx==2{break} println(v) } println()`, "nil"},
		{`for c in "Hello" { println(c) } println()`, "nil"},
		{`for item in (1, true, "Hello") { println(item) } println()`, "nil"},
		{`for idx,v in (1,true,"Hi") {if idx==2 {break} println(v) } println()`, "nil"},
        
		//c-for
		{`for(i=0;i<5;i++){if(i>4){break} if (i==2){continue} println("i=",i)}`, "nil"},
		{`i=0 for(;i<5;i++){if(i>4){break} if(i==2){continue} println("i=",i)}`, "nil"},
		{`i = 0 for (;;;) { if (i > 4) { break } println("i=", i) i++}`, "nil"},

		//hash
		{`for k, v in {"name":"hhf", "age":40} { println(k, " = ", v) }`, "[nil, nil]"},
		{`for k, _ in {"name":"hhf", "age":40} { println("key = ", k) }`, "[nil, nil]"},
		{`for _, v in {"name":"hhf", "age":40} { println("val = ", v) }`, "[nil, nil]"},
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



下一节，我们将加入对`while和do`循环的支持。

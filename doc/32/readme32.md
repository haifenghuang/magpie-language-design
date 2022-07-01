# Panic、Panic、Panic

程序中的致命错误一般有两种：

1. 编译期致命错误（语法解析器解析时候产生的致命错误）
2. 运行期致命错误（解释器在解释代码的时候产生的致命错误）

下面我们分别来讲解一下。



## 语法解析时的致命错误处理

我们知道，在语法解析阶段，`语法解析器（Parser）`发现错误的时候，会报告错误信息。而且我们也有相应的实现。但是，对于语法解析阶段的致命错误（panic），我们的解析器还没有处理这种情况。比如下面的例子（这里假设我们的语法解析器还不能识别`#`）：

```go
fn Add(x, y) # {
    return _add(x, y)
}
```

我们的语法解析器在处理`命名函数(Named Function)`的时候，函数的参数列表处理完成后，应该期待的是一个`{`，但是却遇到了`#`。对于这种情况，我们确实会捕获这个语法错误，但是语法解析器解析语法的时候也会`panic`。产生`panic`的原因，请看代码：

```go
//parser.go
func (p *Parser) parseFunctionStatement() ast.Statement {
	FnStmt := &ast.FunctionStatement{Token: p.curToken}

	p.nextToken()
	FnStmt.Name = &ast.Identifier{Token: p.curToken, Value: p.curToken.Literal}

	FnStmt.FunctionLiteral = p.parseFunctionLiteral().(*ast.FunctionLiteral)

	if p.peekTokenIs(token.TOKEN_SEMICOLON) {
		p.nextToken()
	}

	return FnStmt
}

func (p *Parser) parseFunctionLiteral() ast.Expression {
	lit := &ast.FunctionLiteral{Token: p.curToken}
	if !p.expectPeek(token.TOKEN_LPAREN) {
		return nil
	}
	lit.Parameters = p.parseFunctionParameters()
	if !p.expectPeek(token.TOKEN_LBRACE) {
		return nil
	}
	lit.Body = p.parseBlockStatement()
	return lit
}
```

`parseFunctionStatement()`函数在解析`函数语句(Function Statement)`的时候，会调用`parseFunctionLiteral`函数（第8行）。在`parseFunctionLiteral`函数内部，处理完函数参数后，它会期待下一个词元类型为`TOKEN_LBRACE(即左花括弧)`（23行的`if`判断），如果不是的话，会返回`nil`。问题来了。让我们再看一下`parseFunctionStatement()`函数调用`parseFunctionLiteral`函数那一行：

```go
FnStmt.FunctionLiteral = p.parseFunctionLiteral().(*ast.FunctionLiteral)
```

现在`parseFunctionLiteral()`函数返回`nil`了，而这一行会将其`nil`返回值强制转换成`*ast.FunctionLiteral`。这就会导致致命错误。因此你会看到运行那个例子后，程序的输出大致如下：

```
panic: interface conversion: ast.Expression is nil, not *ast.FunctionLiteral

goroutine 1 [running]:
magpie/parser.(*Parser).parseFunctionStatement(0xc000144ea0, 0xc000160460, 0xc0001579a0)
        D:/HHF/MyWriting/interpreter_new/28/src/magpie/parser/parser.go:355 +0x466
magpie/parser.(*Parser).parseStatement(0xc000144ea0, 0x0, 0x0)
        D:/HHF/MyWriting/interpreter_new/28/src/magpie/parser/parser.go:183 +0x12f
//...

```

输出的第一行非常清楚的打印了相关的错误。为了处理解析器的致命错误，其实方法非常简单，来看代码：

```go
//parser.go
func (p *Parser) ParseProgram() *ast.Program {
	defer func() {
		if r := recover(); r != nil {
			fmt.Printf("Syntax Error:%v- %s\n", p.curToken.Pos, r)
		}
	}()

	program := &ast.Program{}
	//...
}
```

我们只需在`ParseProgram`（即解析代码的入口处）加入相关的`recover`逻辑即可。当解析器产生类似刚才那种致命错误的时候，`defer`函数就会被调用。在`defer`函数内部，我们使用`recover`函数判断是否有`panic`，如果有的话，`recover()`函数的返回值就不为`nil`，这时候我们就将`recover()`的返回值和当前解析器的位置打印出来。

加入了上面的代码后，我们再运行上面的举的例子，输出如下：

```
Syntax Error: <1:12> - interface conversion: ast.Expression is nil, not *ast.FunctionLiteral
Syntax Error: <1:14> - expected next token to be {, got ILLEGAL(#) instead
```



## 解释代码时的致命错误处理

在我们的解释器（Evaluator）调用`Eval()`函数解释代码的过程中，如果程序出现致命错误，该如何处理呢？

举个例子，读者可能知道，我们的程序现在只支持`a++`这种形式，不知道`++a`这种形式。现在假设我们写了一个简单的程序：

```javascript
let x = ++2;
```

运行一下解释器，看看会得到什么结果。在我的机器上，结果输出如下：

```
panic: runtime error: invalid memory address or nil pointer dereference
[signal 0xc0000005 code=0x0 addr=0x28 pc=0x8de080]

goroutine 1 [running]:
magpie/eval.evalLetStatement(0xc000118000, 0xc00004c400)
        C:/work/interpreter_new/31/src/magpie/eval/eval.go:450 +0xe0
magpie/eval.Eval({0x967d40, 0xc000118000}, 0xc00004c400)
        C:/work/interpreter_new/31/src/magpie/eval/eval.go:105 +0x987
magpie/eval.evalProgram(0xc00004c3c0, 0x2030000)
        C:/work/interpreter_new/31/src/magpie/eval/eval.go:149 +0x125
magpie/eval.Eval({0x967e90, 0xc00004c3c0}, 0xc00004c400)
        C:/work/interpreter_new/31/src/magpie/eval/eval.go:24 +0xc5
main.TestEval()
        C:/work/interpreter_new/31/main.go:63 +0x16f
main.main()
        C:/work/interpreter_new/31/main.go:133 +0x7f
```

出来的是`go`语言的错误堆栈。No，No，No！！我们希望的不是看到`go`语言的堆栈，而是`magpie`语言报告的错误信息。

这一节，我们就来实现这个想法。由于我们希望捕获的是运行期的错误，所以只需要修改`解释器(Evaluator)`的代码即可。但是，我们应该在什么地方加入捕获错误信息的代码呢？

大家知道`解释器(Evaluator)`的入口函数是`Eval()`。解释代码的工作就是从这个函数开始的，所以在这个函数中加入错误处理是最理想的。

那么我们怎么捕获运行期的致命错误呢？我们知道`go`语言有一个`recover()`内置函数，用来捕获致命错误并从致命错误中恢复。我们可以利用这个特性来处理程序程序中的致命错误，捕获到致命错误后，将致命错误的信息取出，放入我们的`错误对象(Error Object)`。

> 注意：并不是所有的致命错误都能够恢复！比如段错误，`StackOverflow`等就无法恢复。严格来说，`recover()`函数只能捕捉`panic()`异常。

下面来看一下具体的实现：

```go
//eval.go
func Eval(node ast.Node, scope *Scope) (val Object) {
	defer func() {
		if r := recover(); r != nil {
			val = panicToError(r, node)
		}
	}()

    switch node := node.(type) {
        //...
    }
}

//将panic转换为我们的Error对象
func panicToError(p interface{}, node ast.Node) *Error {
	errLine := node.Pos().Sline()
	switch e := p.(type) {
	case *Error:
		return e
	case error, string, fmt.Stringer:
		return newError(errLine, "%s", e)
	default:
		return newError(errLine, "unknown error:%s", e)
	}
}
```

我们在代码的3-7行中加入了相应的捕获致命错误的逻辑。捕获到致命错误后，使用`recover()`函数得到捕获的错误值，然后将错误值转换成我们的错误对象(Error对象)。第4行使用`panciToError()`函数将`错误对象(Error Object)`赋值给返回值变量`val`。

第15-25行的代码就是将`panic`转换为`Error对象`。如果是Error对象直接返回，如果是`error`、`string`、`fmt.Stringer`则将错误内容取出，使用`newError`将错误内容包装成错误对象(Error Object)。否则就认为是`unknown error`。

更改后，我们再来看一下，运行下面的代码产生的错误是什么。

```javascript
let x = ++2;
```

下面是我本机的运行结果：

```
Runtime Error at 1
        runtime error: invalid memory address or nil pointer dereference
```

看起来是不是好一些了，至少不会再打印出`go`语言的错误堆栈了。



## 总结

通过上面的介绍，我们实现了对于`panic`错误的捕获处理，但是并不代表就万事大吉了。上面的实现只能够捕获`go`语言的`panic`异常，而一些非`panic`的致命错误是无法捕获的。例如：如果出现`Stack Overflow`错误，我们的实现是无法捕获到的。遇到这种错误，还是会出现本节最开始看到的`go`抛出的错误堆栈信息。对于这种情况，我找了很多的资料，至少目前为止，还没有找到更好的解决方案。



下一节，我们会添加`file`内置对象支持。

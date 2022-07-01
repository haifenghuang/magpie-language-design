# `命令执行（Command Execution）`支持

这一节中，我们会给`magpie`语言加入执行外部命令的功能。先看一下使用例子：

```go
date = `date /t` //windows
if !date.ok() {
  printf("An error occurred: %s\n", res)
} else {
  printf("date: %s\n", date)
}
```

执行外部命令使用\`cmd\`的形式（第1行），执行是否成功或者失败，使用`ok()`方法来判断（第2行）。



我们先来看一下我们需要做哪些更改：

1. 在词元（Token）源码`token.go`加入新增的词元类型（\`）。
2. 在词法分析器（Lexer）源码`lexer.go`加入对（\`）的词法分析。
3. 在抽象语法树(AST)源码`ast.go`中加入`命令执行`的抽象语法表示。
4. 在语法解析器（Parser）的源码`parser.go`中增加`命令执行`语法解析。
5. 在解释器（Evaluator）的源码`eval.go`中加入对`命令执行`的解释。



## 词元（Token）的更改

```go
//token.go

const (
	//...
	TOKEN_CMD // `
)

//词元的字符串表示
func (tt TokenType) String() string {
	switch tt {
	//...
	case TOKEN_CMD:
		return "``"
	}
}
```



## 词法解析器（Lexer）的更改

```go
//lexer.go
func (l *Lexer) NextToken() token.Token {
	//...

	switch l.ch {
	//...
	default:
		//...
		} else if l.ch == 34 { //double quotes
			//...
		} else if l.ch == '`' {
			if s, err := l.readCommand(l.ch); err == nil {
				tok.Type = token.TOKEN_CMD
				tok.Literal = s
				return tok
			} else {
				tok.Type = token.TOKEN_ILLEGAL
				tok.Pos = pos
				tok.Literal = err.Error()
				return tok
			}
		} else {
			tok = newToken(token.TOKEN_ILLEGAL, l.ch)
		}
	}
	//...
}
```

11-21行的`else if`是新增加的代码。12行的`readCommand`是实际的读取命令的函数，来看一下其实现：

```go
//lexer.go
func (l *Lexer) readCommand(r rune) (string, error) {
	var ret []rune
eoc:
	for {
		l.readNext()
		switch l.ch {
		case '\r':
		case '\n':
			//遇到回车换行就继续（即无视回车换行）
			continue
		case 0: //如果遇到结束标记，还没有找到`字符，则报错
			return "", errors.New("unexpected EOF")
		case r: //遇到`字符
			l.readNext()
			break eoc //eoc:end of command
		case '\\': //遇到特殊转义字符
			l.readNext()
			switch l.ch {
			case '$': // \$
				ret = append(ret, '\\')
				ret = append(ret, '$')
				continue
			case '`': // \`
				ret = append(ret, '`')
				continue
			}
			ret = append(ret, l.ch)
			continue
		default:
			ret = append(ret, l.ch)
		}
	}

	return string(ret), nil
}
```

我在函数的实现中，加入了比较详细的注释，应该不难理解



## 抽象语法树（AST）的更改

聪明的读者可能很快就能想到，`命令执行`的抽象语法表示，就是一个包含命令字符串的结构。没错，非常简单：

```go
//ast.go
type CmdExpression struct {
	Token token.Token
	Value string
}

func (c *CmdExpression) Pos() token.Position {
	return c.Token.Pos
}

//结束位置 = 开始位置 + 命令的长度
func (c *CmdExpression) End() token.Position {
	length := utf8.RuneCountInString(c.Value)
	return token.Position{Filename: c.Token.Pos.Filename, Line: c.Token.Pos.Line, 
						  Col: c.Token.Pos.Col + length}
}

func (c *CmdExpression) expressionNode()      {}
func (c *CmdExpression) TokenLiteral() string { return c.Token.Literal }
func (c *CmdExpression) String() string       { return c.Value }
```



## 语法解析器（Parser）的更改

对于`命令执行`, 我们只需要给\`符号增加一个前缀回调函数即可：

```go
//parser.go
func (p *Parser) registerAction() {
	//...

	p.registerPrefix(token.TOKEN_CMD, p.parseCommand)
}

// `cmd option1 option2 ...`
func (p *Parser) parseCommand() ast.Expression {
	return &ast.CmdExpression{Token: p.curToken, Value: p.curToken.Literal}
}
```

代码比较简单，就不多做解释了。



## 解释器（Evaluator）的更改

因为我们新增了一个`命令执行`的抽象语法表示`CmdExpression`，因此我们需要在`Eval`函数的`switch`语句中追加一个`case`分支：

```go
//eval.go
func Eval(node ast.Node, scope *Scope) (val Object) {

	switch node := node.(type) {
	//..
	case *ast.CmdExpression:
		return evalCmdExpression(node, scope)
	}

	return nil
}
```

在实现这个`evalCmdExpression`之前，请读者想一下，对于`命令执行`，我们是不是只需要捕获命令的标准输出(stdout)和标准错误(stderr)即可？是的，没错。那么在解释器内部，是不是就可以使用我们的字符串对象（`String`对象）来保存标准输出和标准错误就可以了？想法没错，但是，使用字符串有个问题，如何知道命令执行是成功还是失败呢？其实有两种解决方案：

1. 新增一个`命令执行`对象（Command Object）
2. 返回一个元祖（`tuple`）

下面我先给读者讲解第一种，之后再讲解第二种，具体使用哪种，读者可以自行决定。当然，实现方法不局限于我说的这两种。

### 使用新增命令对象方式

这个新增的`命令对象`包含哪些信息呢？它需要保存命令的标准输出和标准错误。还有，它需要一个`ok`方法，来表示命令是否执行成功：

```go
//object.go
const (
	//...
	CMD_OBJ          = "CMD_OBJ"
)

type Command struct {
	stdout string //保存标准输出信息
	stderr string //保存标准错误信息
	err    bool   //内部用来表示成功或者失败的flag
}

/* 下面是`Object`接口所定义的方法 */
func (c *Command) Inspect() string {
	if c.err {
		return c.stderr
	} else {
		return c.stdout
	}
}

func (c *Command) Type() ObjectType { return CMD_OBJ }
func (c *Command) CallMethod(line string, scope *Scope, method string, args ...Object) Object {
	switch method {
	case "ok": //仅有一个ok方法，用来判断命令运行成功还是失败
		return c.ok(line, args...)
	}

	return newError(line, ERR_NOMETHOD, method, c.Type())
}

func (c *Command) ok(line string, args ...Object) Object {
	if c.err == false { //如果没有错误，返回TRUE对象
		return TRUE
	}
	return FALSE //错误，返回FALSE对象
}
```

有了这个`Command`对象，现在来看一下`evalCmdExpression`的实现：

```go
//eval.go
func evalCmdExpression(t *ast.CmdExpression, scope *Scope) Object {
	cmd := strings.Trim(t.Value, " ") //将命令两边的空格去掉

	// 如果命令中有$var之类的变量，则使用46节的介绍的函数将其替换掉。
	cmd = InterpolateString(cmd, scope)

	var commands []string
	var executor string
	if runtime.GOOS == "windows" { //如果是windows系统，使用'cmd.exe /C'来运行我们的命令
		commands = []string{"/C", cmd}
		executor = "cmd.exe"
	} else { //如果是非windows系统，使用'bash -c'来运行我们的命令
		commands = []string{"-c", cmd}
		executor = "bash"
	}

	var stdout bytes.Buffer
	var stderr bytes.Buffer

	c := exec.Command(executor, commands...) //使用go语言的`exec.Command`来创建一个命令
	c.Env = os.Environ() //命令的环境变量取操作系统的环境变量
	c.Stdin = os.Stdin  //命令的标准输入取操作系统的标准输入
	c.Stdout = &stdout  //命令的标准输出保存到stdout这个变量中
	c.Stderr = &stderr //命令的标准错误保存到stderr这个变量中

	err := c.Run() //运行命令
	if err != nil {
		return &Command{stderr: stderr.String(), err: true} //失败，设置命令对象的stderr和err
	}

	return &Command{stdout: stdout.String(), err: false} //成功，设置命令对象的stdout和err
}
```

代码虽然有点多，但是内容还是比较简单的。细心的读者可能会说了，其实我们的`Command`对象，不需要标准输出和标准错误两个变量，只需要一个即可。是的，实际上只需要一个变量（用来保存命令的标准错误或者标准输出），类似下面这样：

```go
//object.go
type Command struct {
	out string    //保存命令的输出信息（标准输出或者标准错误）
	err    bool   //内部用来表示成功或者失败的flag
}

func (c *Command) Inspect() string {
	return c.out
}
//其它方法不变


//eval.go
func evalCmdExpression(t *ast.CmdExpression, scope *Scope) Object {
	//...

	err := c.Run() //运行命令
	if err != nil {
		return &Command{out: stderr.String(), err: true} //失败，设置命令对象的stderr和err
	}

	return &Command{out: stdout.String(), err: false} //成功，设置命令对象的stdout和err
}
```



### 使用元祖方式

接下来我们来看一下第二种实现，使用元祖（tuple）的方式。读者已经知道，我们的函数可以返回多个值。对于返回多个值的函数，解释器内部实际上是使用的元祖。如果读者有点忘记了，可以参照第25节介绍的知识。这里只列出代码，以供读者参考：

```go
//eval.go
func evalCmdExpression(t *ast.CmdExpression, scope *Scope) Object {
	//...

	tup := NewTuple(true) //新建一个元祖，作为返回对象，这里的true表示函数有多个返回值。
	err := c.Run()
	if err != nil {
		tup.Members[0] = NewString(stderr.String()) //第一个返回值是标准错误信息
		tup.Members[1] = FALSE //第二个返回值是FALSE对象（表示执行失败）
		return tup
	}

	tup.Members[0] = NewString(stdout.String()) //第一个返回值是标准输出信息
	tup.Members[1] = TRUE//第二个返回值是TRUE对象（表示执行成功）
	return tup
}
```

通过上面的介绍，我们实现了两种方式，孰优孰劣，每个人都有不同的见解。我们没有必要为了这个而争论不休。当然，可能还有更好的实现方式。这就验证了我们经常说的`条条大路通罗马`。

> 注：这里的实现使用的是`Run`方法，我们来看一下`Run`方法的官方文档：
>
> ```go
> func (c *Cmd) Run() error
> //Run starts the specified command and waits for it to complete.
> ```
>
> 就是说这里的实现是等待命令的执行完成。如果你不希望等待命令执行完，例如让命令在后台运行，那么你应该使用`Start`方法：
>
> ```go
> func (c *Cmd) Start() error
> //Start starts the specified command but does not wait for it to complete.
> ```
>
> 当然，如果你使用`Start`方法的话，需要做一些额外的工作，例如使用`Wait`函数：
>
> ```go
> func (c *Cmd) Wait() error
> //Wait waits for the command to exit and waits for any copying to stdin or copying from stdout or stderr to complete.
> ```
>
> 不仅是这样，你还需要别的操作，例如`kill`后台命令、互斥锁（mutex）等。
>
> 为了简便起见，这里就不做介绍了。



## 测试

对于测试代码来说，如果我们的实现使用的是新增`命令对象`的方式，则脚本代码如下：

```javascript
res = `curl.exe -s https://api.ipify.org?format=json`
if !res.ok() {
  printf("An error occurred: %s\n", res)
} else {
  printf("res: %s\n", res)
}

res = `date /t`
if !date.ok() {
  printf("An error occurred: %s\n", res)
} else {
  printf("date: %s\n", res)
}
```

如果使用的是返回元祖的形式，则脚本代码如下：

```javascript
res, ok = `curl.exe -s https://api.ipify.org?format=json`
if !ok {
  printf("An error occurred: %s\n", res)
} else {
  printf("res: %s\n", res)
}

res, ok = `date /t`
if !ok {
  printf("An error occurred: %s\n", res)
} else {
  printf("date: %s\n", res)
}

```

一点题外话，运行`date /t`命令的时候，我发现DOS终端显示有乱码，于是上网查了一下解决办法，这里为了方便读者，直接给出解决办法：

1. 在CMD终端，输入`chcp 65001`命令，将代码页更改为UTF-8。（UTF-8的代码页为65001）
2. 在命令行标题栏上点击右键，选择"属性"->"字体"，将字体修改为True Type字体"Lucida Console"，然后点击确定将属性应用到当前窗口。



下一节，我们讲解`增强哈希`支持。




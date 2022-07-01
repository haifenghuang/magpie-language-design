# 重写main函数

到现在为止，我们的执行程序还有两个主要的问题：

* 只能够从字符串中读取输入数据
* 没有注释支持

这一节，就让我们来完善一下吧。

对于注释，我们支持如下三种：

1. `#`开头的是单行注释（Perl、Python、Shell等脚本语言使用的是这种注释）
2. `//`开头的是单行注释（C、Java等使用的是这种注释）
3. `/* */`是多行注释（C、Java等使用的是这种注释）

由于加入注释只涉及新词元（Token）的追加和`词法分析器(Lexer)`的识别，所以比较简单。下面直接列出相应的更改：

```go
//token.go

const (
	//...
	TOKEN_COMMENT    // #, //, /**/

)

//词元类型的字符串表示
func (tt TokenType) String() string {
	switch tt {
	//...
	case TOKEN_COMMENT:
		return "#" //这里使用'#'其实有点不妥，因为我们还支持'//'和'/* */'这两种注释
	}
}
```



```go
//lexer.go
func (l *Lexer) NextToken() token.Token {
	//...
	switch l.ch {
    case '#': //comment
		l.skipComment()
		return l.NextToken()
	}
	//...
	case '/':
		if l.peek() == '/' { //单行注释：//
			l.readNext()
			l.skipComment()
			return l.NextToken()
		} else if l.peek() == '*' { //单行注释：/* */
			l.readNext()
			err := l.skipMultilineComment()
			if err == nil {
				return l.NextToken()
			} else {
				tok.Type = token.TOKEN_ILLEGAL
				tok.Pos = pos
				tok.Literal = err.Error()
				return tok
			}
		} else {
			tok = newToken(token.TOKEN_DIVIDE, l.ch)
		}
}

//忽略注释（单行）
func (l *Lexer) skipComment() {
	for l.ch != '\n' && l.ch != 0 { //一直读到行尾
		l.readNext()
	}
}

//忽略注释（多行）
func (l *Lexer) skipMultilineComment() error {
	var err error = nil
loop:
	for {
		l.readNext()
		switch l.ch {
		case '*':
			switch l.peek() {
			case '/': // 遇到'*/'，则表明遇到了注释结束符
				l.readNext() //skip the '*'
				l.readNext() //skip the '/'
				break loop
			}
		case 0: // 文件结束（这表示多行注释没有结束符）
			err = errors.New("Unterminated multiline comment, GOT EOF!")
			break loop
		}
	}
	return err
}

```

代码虽然有点多，但是应该都不算难懂。对于`skipComment`函数，有的读者可能会问，为什么第33行的代码要判断`l.ch != 0`，到行尾只需要`l.ch !=\n`这个判断就够了。其实有一种情况，就是正好注释是程序的最后一行，且最后一行的后面没有回车换行。

同时注意`skipMultilineComment`函数，因为是多行注释，所以可能会导致没有结束符的问题。就是说读到文件尾，也没有读到`*/`结束符，那么就需要报错（代码52-55行）。



从文件中而不是从字符串中读取源码的实现也是比较简单的，请看代码：

```go
//lexer.go
//从文件中读取程序
func NewFileLexer(filename string) (*Lexer, error) {
	f, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	l := NewLexer(string(f))
	l.filename = filename
	return l, nil
}
```

我们这里做的仅仅是使用`ioutil.ReadFile`读取整个文件的内容，然后调用既有的`NewLexer`函数。

有了这两个实现后，我们就可以把源码写在脚本文件中了，同时文件中也可以有相应的注释。

举个例子：

```javascript
//demo.mp
s1 = "hello, 黄"  // strings are UTF-8 encoded
println(s1)

三 = 3       // UTF-8 identifier
println(三)

i = 20000000         // int
println(i)

b = true               // bool
println(b)

a = [1, "2"]           // array
println(a)

h = {"a": 1, "b": 2}   // hash
println(h)

t = (1,2,3)            // tuple
println(t)

n = nil               // nil
println(n)

# printf builtin
printf("2**3=%g, 2.34.floor=%.0f\n", 2.pow(3), 2.34.floor())

/* below is 
  multiple assignment
*/
a, b, c = 2, false, ["x", "y", "z"]
printf("a=%d,b=%t, c=%v\n", a, b, c)
```

然后执行:

```
magpie demo.mp  //假设执行程序叫`magpie`
```



下一节，我们将给语言增加内置的`正则表达式`支持。


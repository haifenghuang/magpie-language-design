# `正则表达式`支持

这一节中，我们将完成一个相对来说比较有挑战性的内容：正则表达式的支持。

> 先声明一下，对于正则表达式不太熟悉的读者，请先找相关的资料学些一下。这样就会对本节下面的讲解有更好的理解。



我一直非常羡慕`perl`语言内置对正则表达式的支持。而且说实话，我是`perl`语言的忠实用户，我喜欢用`perl`语言做一些分析日志的小脚本，扯得有点远了:smile:。咱们来看`perl`语言的一个例子：

```perl
$bar = "This is foo and again foo"; # 定义一个字符串变量$bar(perl的标量要求用$来开头)
if ($bar =~ /foo/) { # ’=~‘表示匹配的意思。如果$bar变量的内容匹配'foo'
   print "matching\n";
} else {
   print "not matching\n";
}
```

将正则表达式集成到语言层面，非常优雅。我一直期待在自己写的语言中能够实现这种功能。

这一节，我将带领小伙伴们，一起实现这个期待已久的功能。来看看我们的语言，实现正则表达式支持后的例子：

```perl
# 第一个例子
name = "Huang HaiFeng"
if name =~ /Huang/ { # ’=~‘ 匹配的意思
    println("matched 'Huang'")
}
if name !~ /foo/ { # '!~' 不匹配的意思
    println("not matched 'foo'")
}

#第二个例子
name = "Huang HaiFeng"
# ’=~‘ 匹配的意思, 'i'是正则表达式的修饰符(flag或modifier)，表示不区分大小写(ignore-case)
if name =~ /huang/i {
    println("matched 'huang'")
}
```

从上面的例子中，我们可以了解到，我们这里新增了：

1. `=~ `操作符表示匹配
2. `!~`操作符表示不匹配
3. /regex/表示正则表达式的pattern（即匹配模式）



好了，有了上面的介绍，就让我们开始一步一步实现对正则表达式的支持吧。

和以往的文章一样，让我们看一下需要做哪些变更：

1. 在词元（Token）源码`token.go`中加入新的词元（Token）类型

2. 在词法解析器（Lexer）源码`lexer.go`中，增加对新的词元（Token）类型的解析

3. 在抽象语法树（AST）的源码`ast.go`中加入`正则表达式字面量`对应的抽象语法表示

4. 在语法解析器（Parser）的源码`parser.go`中加入对`正则表达式字面量`的语法解析

   > 还需要对`=~`和`!~`这两个中缀操作符赋予优先级

5. 在对象系统的源码`object.go`中，增加一个`正则表达式`对象。

6. 在解释器（Evaluator）的源码`eval.go`中加入对`正则表达式字面量`的解释及`=~`和`!~`操作符的解释。



## 词元(Token)的更改

```go
//token.go
const (
	//...
	TOKEN_MATCH    // =~ 匹配
	TOKEN_NOTMATCH // !~ 不匹配
	TOKEN_REGEX // 正则表达式（regular expression）
)

//词元类型的字符串表示
func (tt TokenType) String() string {
	switch tt {
	//...
	case TOKEN_MATCH:
		return "=~"
	case TOKEN_NOTMATCH:
		return "!~"
	case TOKEN_REGEX:
		return "<REGEX>"
	}
}
```



## 词法解析器(Lexer)的更改

```go
//lexer.go

var prevToken token.Token //前一个词元
func (l *Lexer) NextToken() token.Token {
	//...

	switch l.ch {
	//...
	case '/':
		if l.peek() == '/' { //单行注释
			l.readNext()
			l.skipComment()
			return l.NextToken()
		}

		// '/'通常表示除法，但也可能是一个正则表达式的开头
		if prevToken.Type == token.TOKEN_RPAREN || // (a+c) / b
			prevToken.Type == token.TOKEN_RBRACKET || // a[3] / b
			prevToken.Type == token.TOKEN_IDENTIFIER || // a / b
			prevToken.Type == token.TOKEN_NUMBER { // 3 / b,  3.5 / b
			tok = newToken(token.TOKEN_DIVIDE, l.ch)
		} else { //正则表达式（regexp）
			if regStr, err := l.readRegExLiteral(); err == nil {
				tok.Literal = regStr
				tok.Type = token.TOKEN_REGEX
				tok.Pos = pos
				return tok
			} else {
				tok.Type = token.TOKEN_ILLEGAL
				tok.Pos = pos
				tok.Literal = err.Error()
				return tok
			}
		}
	case '=':
		if l.peek() == '=' {
			tok = token.Token{Type: token.TOKEN_EQ, 
                              Literal: string(l.ch) + string(l.peek())}
			l.readNext()
		} else if l.peek() == '~' { //匹配
			tok = token.Token{Type: token.TOKEN_MATCH, 
                              Literal: string(l.ch) + string(l.peek())}
			l.readNext()
		} else {
			tok = newToken(token.TOKEN_ASSIGN, l.ch)
		}
	case '!':
		if l.peek() == '=' {
			tok = token.Token{Type: token.TOKEN_NEQ, 
                              Literal: string(l.ch) + string(l.peek())}
			l.readNext()
		} else if l.peek() == '~' { //不匹配
			tok = token.Token{Type: token.TOKEN_NOTMATCH, 
                              Literal: string(l.ch) + string(l.peek())}
			l.readNext()
		} else {
			tok = newToken(token.TOKEN_BANG, l.ch)
		}
		//...
	default:
		//这个`default`分支中的所有`return tok`的语句之前加入`prevToken = tok`这句，
        //用来记住之前的token
	}

	tok.Pos = pos
	l.readNext()
	prevToken = tok //返回前记住当前的token
	return tok
}
```

40-43行我们追加了`=~`的判断。52-55行追加了`!~`的判断。

这里重点说一下第9行的`case`分支。大家知道`/`这个操作符，在之前我们的代码中是表示除法的中缀操作符（比如`a/b`）。现在有了`正则表达式字面量`的支持，我们就要区分这个`/`是正则表达式的开始，还是一个中缀的除法符号。

因此我们在第3行增加了一个`prevToken`这个变量，用来记录前一个词元。当遇到`/`的时候，我们需要判断前一个词元的类型，如果前一个词元的类型是下面的几种，则我们就认为这个`/`符号是代表除法：

1. 右括号。       比如:  (a+c) / b
2. 右方括号。   比如：a[3] / b
3. 标识符。       比如：a / b
4. 数字。            比如：3 / b,  3.5 / b

如果不是上面的情况，我们就认为这个`/`是一个正则表达式的开头。

下面让我们来看一下如果`/`表示正则表达式的时候，它的词法分析代码：

```go
//lexer.go

// 读取正则表达式，包含其修饰符(flag或modifier)
//形式: /regexp/imsU: 这个imsU的含义见下面代码的注释
func (l *Lexer) readRegExLiteral() (string, error) {
	out := ""

	for {
		l.readNext()

		if l.ch == 0 { //读到最后,还没有遇到下一个'/'的情况下，那么说明正则表达式不完整
			return "unterminated regular expression", fmt.Errorf("unterminated regular expression")
		}

		if l.ch == '/' { // 最后的 "/".
			l.readNext()
            //分析是否有正则表达式的flag
			flags := ""

            //下面是'go'语言支持的修饰符(flag or modifier):
			//   i -> case-insensitive （不区分大小写）
            //   m -> multi-line mode (多行模式)
            //   s -> let . match \n ('.'匹配回车。通常我们使用正则表达式中的'.'是不匹配回车的)
			//   U -> ungreedy （非贪婪模式。通常我们使用正则表达式是贪婪模式的，就是尽可能多的匹配）
			for l.ch == 'i' || l.ch == 'm' || l.ch == 's' || l.ch == 'U' {
				// 如果'/'后面跟着'imsU'这几个字符，则先保存这个字符，如果用户重复提供
                // 某个flag，比如'/regext/ii',我们需要去重，即只保留一个'i'
				if !strings.Contains(flags, string(l.ch)) {
					tmp := strings.Split(flags, "")
					tmp = append(tmp, string(l.ch))
					flags = strings.Join(tmp, "")
				}

				l.readNext()
			}

			// 在'go'语言中，正则表达式的修饰符是放在前面的。所以我们需要转换成'go'语言
			// 能够正确识别的正则表达式，例如：
			//     | --------------------------- |
			//     |    mapgie   |      go       |
			//     | --------------------------- |
            //     | /regexp/im  |  (?im)regexp  |
			//     | --------------------------- |
			if len(flags) > 0 {
				out = "(?" + flags + ")" + out
			}
			break
		}
		out = out + string(l.ch)
	}

	return out, nil
}
```

上面的代码，我在关键的地方都加了注释，请读者好好理解一下。



### 抽象语法树（AST）的更改

我们需要加入一个`正则表达式字面量`对应的抽象语法表示：

```go
//ast.go
//正则表达式字面量:(?flag)pattern
type RegExLiteral struct {
	Token token.Token
	Value string //(?flag)pattern
}

func (rel *RegExLiteral) Pos() token.Position {
	return rel.Token.Pos
}

func (rel *RegExLiteral) End() token.Position {
	length := utf8.RuneCountInString(rel.Value)
	pos := rel.Token.Pos
	return token.Position{Filename: pos.Filename, Line: pos.Line, Col: pos.Col + length}
}

//正则表达式字面量是一个表达式
func (rel *RegExLiteral) expressionNode()      {}
func (rel *RegExLiteral) TokenLiteral() string { return rel.Token.Literal }

// '正则表达式字面量'的字符串表示
// 这里主要是将'(?flag)pattern' 转换成`/pattern/flag`的形式
func (rel *RegExLiteral) String() string {
    	reg := rel.Value
	begin := strings.Index(reg, "(?")
	if begin == -1 { //没找到，表明没有修饰符
		return fmt.Sprintf("/%s/", reg)
	}
	end := strings.Index(reg, ")")
	val := reg[end+1:]
	flag := reg[begin+2 : end]
	return fmt.Sprintf("/%s/%s", val, flag)

}
```

这里需要注意的是`正则表达式字面量`的字符串表示，因为我们存储的是`(?flag)pattern`这种`go`语言形式，所以需要转换成`magpie`语言的`/pattern/flag`的形式。



## 语法解析器（Parser）的更改

对于语法解析器（Parser）来说，我们需要做如下更改：

1. 给`TOKEN_MATCH`和`TOKEN_NOTMATCH`两个词元类型注册中缀表达式回调函数
2. 给`TOKEN_MATCH`和`TOKEN_NOTMATCH`两个词元类型增加优先级
3. 给`TOKEN_REGEX`词元类型注册前缀表达式回调函数

下面来看实际代码：

```go
//parser.go
func (p *Parser) registerAction() {
    //...
	p.registerPrefix(token.TOKEN_REGEX, p.parseRegexpLiteral) //给'/pattern/'注册前缀表达式回调函数

	//...
	p.registerInfix(token.TOKEN_MATCH, p.parseInfixExpression)    // 给'=~'注册中缀表达式回调函数
	p.registerInfix(token.TOKEN_NOTMATCH, p.parseInfixExpression) // 给'!~'注册中缀表达式回调函数
}
```

下面我们来看一下`TOKEN_MATCH`和`TOKEN_NOTMATCH`这两个词元类型的优先级：

```go
//parser.go
const (
	//...
	PRODUCT      //*, /, **
	REGEXP_MATCH // !~, ~=
	PREFIX       //!true, -10
)

var precedences = map[token.TokenType]int{
	//...
	token.TOKEN_MATCH:    REGEXP_MATCH,
	token.TOKEN_NOTMATCH: REGEXP_MATCH,
}
```

我们给`匹配(=~)`和`不匹配(!~)`这两个中缀操作符指定了优先级（第5行）。它比`乘除及乘方`高一级，比`前缀操作符`低一级。这个是参考`perl`语言而给的优先级。然后给`TOKEN_MATCH`和`TOKEN_UNMATCH`赋予了`REGEXP_MATCH`优先级（第11-12行）。

> 这里说一下优先级的问题。其实有个很简单的方法，就是参照流行语言。

最后来看一下我们还没有实现的`parseRegexpLiteral`前缀回调函数：

```go
//parser.go
// parses a regular-expression
func (p *Parser) parseRegexpLiteral() ast.Expression {
	return &ast.RegExLiteral{Token: p.curToken, Value: p.curToken.Literal}
}
```

直接返回`正则表达式字面量`的抽象语法表示结构。



## 对象系统(Object System)的更改

对于`正则表达式字面量`这个抽象语法表示，我们需要新建一个`正则表达式对象(regexp object)`。

最简单的方式就是看代码：

```go
//object.go
const (
	//...
	REGEX_OBJ        = "REGEX" //新增一个'正则表达式'对象类型
)
```

现在让我们来想想，我们的`正则表达式对象`需要什么样的内容呢？首先，我们需要包含一个`模式（Patern）`字符串，这个是正则表达式需要匹配的模式（pattern）。其次，因为我们需要处理正则表达式的解析，匹配之类的操作，所以还需要包含一个`go`语言中处理正则表达式的对象：`*regexp.Regexp`：

```go
//object.go
type RegEx struct {
	RegExp *regexp.Regexp //保存go语言处理正则表达式的对象指针
    Value  string //正则表达式的pattern: (?flag)pattern
}
```

我们的`RegEx`结构还是比较简单的。现在我们要让其实现`Object`接口：

```go
//object.go

//实现Object接口的方法
func (re *RegEx) Inspect() string  { return re.Value }
func (re *RegEx) Type() ObjectType { return REGEX_OBJ }

func (re *RegEx) CallMethod(line string, scope *Scope, method string, args ...Object) Object {
	switch method {
	case "match":
		return re.match(line, args...)
	case "replace":
		return re.replace(line, args...)
	case "split":
		return re.split(line, args...)
	}
	return newError(line, ERR_NOMETHOD, method, re.Type())
}

func (re *RegEx) match(line string, args ...Object) Object {
	if len(args) != 1 {
		return newError(line, ERR_ARGUMENT, "1", len(args))
	}

	if args[0].Type() != STRING_OBJ {
		return newError(line, ERR_PARAMTYPE, "first", "match", "*String", args[0].Type())
	}

	str := args[0].(*String)
	matched := re.RegExp.MatchString(str.String) //调用'regexp’包的'MatchString'方法
	if matched {
		return TRUE
	}
	return FALSE
}

func (re *RegEx) replace(line string, args ...Object) Object {
	if len(args) != 2 {
		return newError(line, ERR_ARGUMENT, "2", len(args))
	}

	if args[0].Type() != STRING_OBJ {
		return newError(line, ERR_PARAMTYPE, "first", "replace", "*String", args[0].Type())
	}

	if args[1].Type() != STRING_OBJ {
		return newError(line, ERR_PARAMTYPE, "second", "replace", "*String", args[1].Type())
	}

	str := args[0].(*String)
	repl := args[1].(*String)
	result := re.RegExp.ReplaceAllString(str.String, repl.String) //调用'regexp’包的'ReplaceAllString'方法
	return NewString(result)
}

func (re *RegEx) split(line string, args ...Object) Object {
	if len(args) != 1 {
		return newError(line, ERR_ARGUMENT, "1", len(args))
	}

	if args[0].Type() != STRING_OBJ {
		return newError(line, ERR_PARAMTYPE, "first", "split", "*String", args[0].Type())
	}

	str := args[0].(*String)
	splitResult := re.RegExp.Split(str.String, -1) //调用'regexp’包的'Split'方法

	a := &Array{}
	for i := 0; i < len(splitResult); i++ {
		a.Members = append(a.Members, NewString(splitResult[i]))
	}
	return a
}
```

我们给`正则表达式`对象提供了简单的几个函数供用户调用，它们分别是`match`、`replace`和`split`函数。当然可以根据需要增加更多的方法。这几个函数内部主要就是做了以下几个动作：

1. 判断参数个数
2. 判断参数类型
3. 调用相应的go语言的`Regexp`包中的相应方法



## 解释器（Evaluator）的更改

对于解释器（Evaluator）来说，第一，我们需要在`Eval()`函数的`switch`中增加一个新的`case`分支：

```go
//eval.go
func Eval(node ast.Node, scope *Scope) (val Object) {
	switch node := node.(type) {
	//...
	case *ast.RegExLiteral:
		return evalRegExLiteral(node, scope)
	}

	return nil
}

//解释`正则表达式字面量`
func evalRegExLiteral(node *ast.RegExLiteral, scope *Scope) Object {
    //调用regexp包的`Compile`方法，传入pattern(这里是node.Value)
	regExp, err := regexp.Compile(node.Value)
	if err != nil {
		return newError(node.Pos().Sline(), ERR_INVALIDARG)
	}

    //返回正则表达式对象
	return &RegEx{RegExp: regExp, Value: node.Value}
}

//errors.go
var (
	ERR_INVALIDARG      = "invalid argument supplied"
)
```

我们在代码的第5行新增了一个`case`分支，用来处理`正则表达式字面量`。

代码13-22行是处理正则表达式字面量的实际代码。

由于第17行我们多了一个`ERR_INVALIDARG`变量，所以我们在`errors.go`文件加入了这个变量(代码26行)。

对于`匹配(=~)`和`非匹配(!~)`这两个中缀操作符，我们还需要增加相关的处理代码：

```go
//eval.go
//<left-expression> operator <right-expression>
func evalInfixExpression(node *ast.InfixExpression, left, right Object, scope *Scope) Object {
	operator := node.Operator
	switch {
    case operator == "=~", operator == "!~": //如果是匹配(=~)或者非匹配(!~)操作符
		if right.Type() != REGEX_OBJ {
			return newError(node.Pos().Sline(), ERR_NOTREGEXP, right.Type())
		}

        //left =~ /pattern/
        //left !~ /pattern/
        //调用MatchString方法判断左边的对象是否匹配右边的模式(pattern)
		matched := right.(*RegEx).RegExp.MatchString(left.Inspect())
		if matched { //如果匹配的话
			if operator == "=~" { //如果是匹配操作符,则返回`TRUE`对象
				return TRUE
			}
			return FALSE
		} else { //不匹配
			if operator == "=~" { //如果是匹配操作符,则返回`FALSE`对象
				return FALSE
			}
			return TRUE
		}
    }
}

//errors.go
var (
	ERR_NOTREGEXP       = "right type is not a regexp object, got %s"
)
```

我们给解释中缀表达式的函数`evalInfixExpression`的`switch`增加了一个`case`分支（第6行）。第7行我们判断右面的对象（Object）是否是一个`正则表达式`对象，如果不是则报错。这里我们新增了一个`ERR_NOTREGEXP`错误变量，在第31行进行的赋值。

我们在第14行调用了go语言`RegExp`包提供的`MatchString()`方法来判断左边的对象是否匹配右边的正则表达式模式。



我们对`正则表达式`的支持处理完成了，下面让我们写个简单的测试程序来测试一下：



## 测试

```go
//main.go
package main

import (
	"fmt"
	"magpie/eval"
	"magpie/lexer"
	"magpie/parser"
	"os"
)

func TestEval() {
	tests := []struct {
		input    string
		expected string
	}{
		{`name="Huang HaiFeng";if name =~ /huang/i { println("Hello Huang") }`, "nil"},
		{`name="Huang HaiFeng";if (name !~ /xxx/) { println( "Hello xxx" ) }`, "nil"},
		{`name="Huang HaiFeng";if name =~ /Huang/ { println("Hello Huang") }`, "nil"},
        {`match=/\d+\t/.match("abc 123	mnj");if (match) { println("matched") }`, "nil"},
	}

	for _, tt := range tests {
		l := lexer.NewLexer(tt.input)
		p := parser.NewParser(l)
		program := p.ParseProgram()

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



下一节，我们讲解如何包装`go`对象，让我们的程序能够使用go语言包中的方法和变量。

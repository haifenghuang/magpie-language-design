# `逻辑与(&&)、逻辑或(||)`支持

在这一节中，我们要加入对于`逻辑与(&&)、逻辑或(||)`的支持。我们前面讲过的`if-else`判断只支持一个判断条件：

```perl
if a > b {
    # do something
}
```

学习了这节的内容后，我们的`if-else`表达式就可以支持多个判断条件了：

```perl
if a > b && a < c || b > d {
    # do something
}
```

和之前一样，我们来看一下需要做的更改：

1. 在词元（Token）源码`token.go`加入新增的词元类型（`TOKEN_AND`和`TOKEN_OR`）
2. 在词法分析器（Lexer）源码`lexer.go`加入对`&&`和`||`的词法分析
3. 在语法解析器（Parser）的源码`parser.go`中注册`&&`和`||`的中缀表达式回调函数及它们的优先级。
5. 在解释器（Evaluator）的源码`eval.go`中加入对`&&`和`||`这两个中缀表达式的解释。

## 词元（Token）的更改

```go
//token.go

const (
	//...
	TOKEN_AND // &&
	TOKEN_OR  // ||

)

//词元的字符串表示
func (tt TokenType) String() string {
	switch tt {
	//...
	case TOKEN_AND:
		return "&&"
	case TOKEN_OR:
		return "||"
	//...
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
	case '&':
		if l.peek() == '&' {
			tok = token.Token{Type: token.TOKEN_AND, Literal: string(l.ch) + string(l.peek())}
			l.readNext()
		}
	case '|':
		if l.peek() == '|' {
			tok = token.Token{Type: token.TOKEN_OR, Literal: string(l.ch) + string(l.peek())}
			l.readNext()
		}
	}
}
```

第7-16行加入了对`&&`和`||`识别的判断。



## 语法解析器（Parser）的更改

首先我们注册`&&`和`||`的中缀表达式回调函数：

```go
//parser.go

func (p *Parser) registerAction() {
	//...
	p.registerInfix(token.TOKEN_AND, p.parseInfixExpression)
	p.registerInfix(token.TOKEN_OR, p.parseInfixExpression)
}
```

第5-6行注册了`TOKEN_END`和`TOKEN_OR`这两个词元类型的中缀表达式回调函数。

因为`&&`和`||`这两个操作符都是中缀操作符，所以我们还需要赋予其优先级：

```go
//parser.go
const (
	_ int = iota
	LOWEST
	ASSIGN      // =
	CONDOR      // ||
	CONDAND     // &&
	//...
)

var precedences = map[token.TokenType]int{
	token.TOKEN_ASSIGN: ASSIGN,
	token.TOKEN_OR:     CONDOR,
	token.TOKEN_AND:    CONDAND,
    //...
}
```

我们给`&&`和`||`这两个操作符赋予了比`赋值(=)`高的优先级。这个和`go`语言及`c`语言是一样的。



## 解释器（Evaluator）的更改

因为`&&`和`||`这两个操作符都是中缀操作符，所以我们需要给`evalInfixExpression()`函数的`switch`语句增加两个`case`分支：

```go
//eval.go

func evalInfixExpression(node *ast.InfixExpression, left, right Object, scope *Scope) Object {
	operator := node.Operator
	switch {
	case operator == "&&":
		leftCond := objectToNativeBoolean(left) //将左边表达式转换为布尔类型
		if !leftCond { //对于'&&'，如果左边条件为fasle, 则直接返回FALSE对象，不再计算右边的表达式
			return FALSE
		}

		rightCond := objectToNativeBoolean(right)
		return nativeBoolToBooleanObject(leftCond && rightCond)
	case operator == "||":
		leftCond := objectToNativeBoolean(left) //将左边表达式转换为布尔类型
		if leftCond { //对于'&&'，如果左边条件为true, 则直接返回TRUE对象，不再计算右边的表达式
			return TRUE
		}

		rightCond := objectToNativeBoolean(right)
		return nativeBoolToBooleanObject(leftCond || rightCond)
	//...
	}
}

```

代码的6-21行我们增加了处理`&&`和`||`操作符的逻辑。

> 从代码中可以看到，我们对其使用了短路操作。例如：
>
>   ```java
>   if 5 > 10 && 12 > 5 { xxx } //因为第一个判断就返回false，所以第二个判断'12 > 5'就不会继续判断了,直接返回true
>   if 15 > 10 || 12 > 5 { xxx } //因为第一个判断就返回true，所以第二个判断'12 > 5'就不会继续判断了,直接返回true
>   ```

上面的代码调用了`objectToNativeBoolean`函数，这个函数从名字也能猜出一二，就是将对象（Object）转换为布尔值：

```go
//eval.go
func objectToNativeBoolean(o Object) bool {
	if r, ok := o.(*ReturnValue); ok {
		o = r.Value
	}
	switch obj := o.(type) {
	case *Boolean:
		return obj.Bool
	case *Nil:
		return false
	case *Number:
		if obj.Value == 0.0 {
			return false
		}
		return true
	case *String:
		return obj.String != ""
	case *Array:
		if len(obj.Members) == 0 {
			return false
		}
		return true
	case *Tuple:
		if len(obj.Members) == 0 {
			return false
		}
		return true
	case *Hash:
		if len(obj.Pairs) == 0 {
			return false
		}
		return true
	default:
		return true
	}
}
```

代码中分别对对象系统中的对象（Object）进行了相应的处理，大致如下：

| 对象                 | 处理                                                 |
| -------------------- | ---------------------------------------------------- |
| 布尔对象（Boolean）  | 返回布尔类型的真假值                                 |
| Nil对象（NIL）       | 返回false                                            |
| 数字对象（Number）   | 数值为0的时候返回false，否则返回true                 |
| 字符串对象（String） | 不为空的时候返回true，否则返回false                  |
| 数组对象（Array）    | 长度大于零的时候返回true，否则返回false              |
| 元祖对象（Tuple）    | 长度大于零的时候返回true，否则返回false              |
| 哈希对象（Hash）     | key/value对的长度大于零的时候返回true，否则返回false |
| 其它                 | true                                                 |



## 测试

下面我们写一个简单的程序测试一下：
```go
//main.go
func TestEval() {
	tests := []struct {
		input    string
		expected string
	}{
		{`if 10 == 10 && 10 > 5 { printf("10 == 10 && 10 > 5\n")}`, "nil"},
		{`if 10 == 10 && 10 > 12 { printf("10 == 10 && 10 > 12\n") } else { println("10 not larger than 12") }`, "nil"},
		{`if 10 == 10 || 10 > 12 { printf("10 == 10 || 10 > 12\n")}`, "nil"},
		{`if 10 == 11 || 10 > 12 { printf("10 == 11 || 10 > 12\n") } else { println(" 10 not equal 11 and 10 not larger than 12") }`, "nil"},

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



下一节，我们将增加对`导入(import)`的支持。


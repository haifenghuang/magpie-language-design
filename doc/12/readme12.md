# `内置（Built-in）函数`支持

小喜鹊现在能够正常飞行了，但是由于练习的不够，还不能飞的很远。就是说，它的【内功】还不是很强，这样飞起来后就无法抵御强风。通过本节的学习，小喜鹊的内功就能够练成了。

这里说的【内功】，其实就类似我们语言的内置函数（Built-in function）。有了强大的内置函数，语言才会变得更加充实。

由于本节是讨论`内置函数`，所以只涉及解释阶段（Evaluating Phase）。

为啥叫`内置(Built-in)函数`呢？因为这种函数是由脚本语言的开发者提供的，并不是脚本语言的使用者写的，但是脚本的使用者可以直接使用这些函数。对于`内置函数`它需要哪些参数，它的返回值又是什么呢？

我们以将要在本节实现的`len`函数为例，用户的调用例子如下：

```go
len("Hello") = 5
len([1,2,3]) = 3
```

可以看到`len`内置函数接受一个对象作为参数:

* 第一行接受一个字符串对象（String Object）
* 第二行接受一个数组对象（Array Object）  ---->这个数组对象我们还没有讲到
* 返回的是一个`数字对象（Number Object)`。第一行返回`5(字符串长度)`，第二行返回`3(数组长度)`

那么在程序内部我们怎么处理这个`len内置函数`调用呢？

从上面的例子中，我们知道内置函数：

1.  接受0个或多个`对象(Object)`参数
2.  返回一个`对象(Object)`

除了这两点，还需要哪些信息呢。假设用户像下面这样调用这个`len()内置函数`：

```go
len(3)
```

那么我们就需要给用户报告错误，报告错误就需要【行号】。所以【行号】就是需要的另一个信息。最后一个需要的信息就是调用`len()内置函数`所需要的`作用域（Scope）`。这个`作用域（Scope）`信息在大部分情况下用不到，但是在个别地方会用到，我们在后续的章节中会用到。

总结一下，调用`内置函数`需要提供如下信息：

1. Scope（作用域）  --> 当前调用内置函数的作用域
2. 0个或多个`对象(Object)`参数
3. 行号（主要是用来报错用）。
4.  返回一个`对象(Object)`。

有了上面的解释，我们就能够得出一个统一的`内置函数`的原型了：

```go
type BuiltinFunc func(line string, scope *Scope, args ...Object) Object
```

它只是一个可调用函数的类型定义。由于我们需要让这些内置函数对我们的用户可用，因此需要将它安装到我们的`对象系统（Object System）`中。简单说就是我们需要一个`内置函数对象（'Builtin-Function' Object）`。这个`内置函数对象`会包装（wrapping）这个类型，还是看一下代码就更清楚了：

```go
//builtin.go

//声明一个可调用函数类型
type BuiltinFunc func(line string, scope *Scope, args ...Object) Object

type Builtin struct {
	Fn BuiltinFunc //变量'Fn'是个可调用函数，你可以简单的认为是一个函数指针
}

func (b *Builtin) Inspect() string  { return "<builtin function>" }
func (b *Builtin) Type() ObjectType { return BUILTIN_OBJ }
```

这就是`内置函数对象('Builtin-Function' Object)`。这里的`BUILTIN_OBJ`我们需要在`object.go`中进行定义：

```go
//object.go
const (
	//...

	BUILTIN_OBJ      = "BUILTIN"
)
```

我们还需要一个map，这个map的key存放着所有的`内置函数的名称`，值就是这里说的`内置函数对象（即Builtin结构）`

```go
//builtin.go
var builtins map[string]*Builtin
```

这个变量我们会在`init()`方法中去初始化它，这样脚本用户就能够安全的使用这个内置函数了：

```go
func init() {
	builtins = map[string]*Builtin{
		"len":     lenBuiltin(),
	}
}
```

> `init()`函数先于main函数自动执行。

下面我们来看一下`lenBuiltin()`这个函数的实现：

```go
//builtin.go
func lenBuiltin() *Builtin {
	//创建一个新的内置（Builtin）函数对象返回
	return &Builtin{ Fn: len_function }
}

fn len_function(line string, scope *Scope, args ...Object) Object {
	if len(args) != 1 { //`len`函数只接受一个参数
        return newError(line, "wrong number of arguments. got %d, want=1", len(args))
	}

    //判断给`len()`函数提供的参数的类型
	switch arg := args[0].(type) {
	case *String:
		//如果是字符串类型，则计算出字符串的长度后，返回一个数字对象（数字对象中包含计算后的字符串长度）。
		n := utf8.RuneCountInString(arg.String)
        return NewNumber(float64(n))
	default:
		//报告不支持的类型错误
		return newError(line, "argument to `len` not supported, got %s", args[0].Type())
	}
}
```

可以看到，`lenBuiltin`这个函数返回了一个新的`Builtin`结构。而这个结构中的`Fn`变量中存储的值为`len_function()`函数。

为啥需要`builtins`这个map函数呢？举个例子：

```go
sub(1,2)
len("Hello")
```

对于这两个调用，我们的`解释器（Evaluator）`只知道这两个都是函数调用，无法区分哪个是用户写的函数，哪个是`内置函数`。这时候我们就可以在`builtins`map中查找，如果找到了，就知道这是个内置函数调用，找不到的话，就知道是普通函数调用。对于第一个`sub(1,2)`函数调用，程序中大概是这样的：

```go
if _, ok := builtins["sub"]; ok { //不满足(没找到)

} else {
	//调用`sub`函数
}
```

对于第二个`len("Hello")`函数调用，程序中大概是这样的（为了说明问题使用了伪代码）：

```go
if builtinFunc, ok := builtins["len"]; ok { //满足
    //这个builtinFunc就是'lenBuiltin()'函数
	//调用这个函数,返回的是"Builtin"这个结构对象
	len_builtinObj = builtFunc()
    //调用`len_builtinObj`对象中存储的'Fn'函数(即'len_function'这个函数)。
	returnObj = len_builtinObj.Fn(<"Hello"节点的当前行号>，<当前的scope>，<"Hello"这个字符串对象)
	return returnObj
} else {

}
```

其实`lenBuiltin()`函数的代码，我们还可以像下面这么写（更紧凑的写法），只不过稍微不好理解罢了，无论采用哪种方式都可以。

```go
//builtin.go
func lenBuiltin() *Builtin { //返回`Builtin`这个内置函数对象
	return &Builtin{ //创建一个Builtin函数对象
		//可调用函数
		Fn: func(line string, scope *Scope, args ...Object) Object {
            if len(args) != 1 { //我们的内置函数`len`只接受一个参数
				return newError(line, 
                                "wrong number of arguments. got %d, want=1", len(args))
			}

			//判断给`len()`函数提供的参数的类型
			switch arg := args[0].(type) {
			case *String:
				//如果是字符串类型，则计算出字符串的长度后，
                //返回一个数字对象（数字对象中包含计算后的字符串长度）。
				n := utf8.RuneCountInString(arg.String)
				return NewNumber(float64(n))
			default:
				//报告不支持的类型错误
				return newError(line, 
                                "argument to `len` not supported, got %s", args[0].Type())
			}
		},
	}
}
```

我们已经实现了`len()`这个内置对象。但是我们的脚本现在还不能打印信息，现在让我们来实现它（`print`和`println`）：

```go
//builtin.go
func init() {
	builtins = map[string]*Builtin{
		"len":     lenBuiltin(),
		"print":   printBuiltin(),
		"println": printlnBuiltin(),
    }
}

fn print_function(line string, scope *Scope, args ...Object) Object {
    resultStr := ""
	for _, arg := range args {
		resultStr = resultStr + arg.Inspect()
	}
	fmt.Print(resultStr)
	return NIL
}

func printBuiltin() *Builtin {
	return &Builtin{ Fn: print_function }
}


fn println_function(line string, scope *Scope, args ...Object) Object {
    if len(args) == 0 {
		fmt.Println()
    }

    resultStr := ""
	for _, arg := range args {
		resultStr = resultStr + arg.Inspect() + "\n"
	}
	fmt.Print(resultStr)
	return NIL
}

func printlnBuiltin() *Builtin {
	return &Builtin{ Fn: println_function }
}
```

> 注：这里面为了简单起见，并没有判断参数个数。同时`print`和`println`这两个内置函数，打印出信息后，返回的是`NIL`对象，而不是`fmt.Print`和`fmt.Println`的返回值。这个根据需求可以自行更改。

这里重点需要说一下`println`函数。我们先来看一下`go`语言的`fmt.Println`函数的一个问题。假设我们有如下的`go`代码：

```go
fmt.Println("a=", 10)
```

我们期待的结果应该是：

```go
a=10
```

但是我们看到的输出却是这样：

```go
a= 10
```

相信眼尖的读者已经看出来了：`等号右边多了一个空格`。这个其实`go语言`的官方文档中已经给了详细的解释：

> `Println` formats using the default formats for its operands and writes to standard output. `Spaces are always added between operands` and a newline is appended.
>
> 翻译过来就是：
>
> `Println`对它的操作数使用缺省的格式并写到标准输出中。`操作数之间总是会加入空格`，且最后会追加一个`新行(newline)`。

那么有了上面的说明：`操作数之间总是会加入空格`。解决方法其实也很简单，就是把`Println`函数的所有参数连接成一个参数，这样的话，就只有一个`操作数（operands）`了。上面代码的25-28行就是将所有的参数连接成一个参数。

现在剩下的就是需要改动解析器（Evaluator）解释内置函数的地方了，具体做法也比较简单：

```go
//eval.go
func applyFunction(line string, scope *Scope, fn Object, args []Object) Object {
	switch fn := fn.(type) {
	case *Function: //用户提供的函数
		extendedScope := extendFunctionScope(fn, args)
		evaluated := Eval(fn.Literal.Body, extendedScope)
		return unwrapReturnValue(evaluated)
	case *Builtin: //内置函数
		return fn.Fn(line, scope, args...) //调用内置函数
	default:
		return newError(line, ERR_NOTFUNCTION, fn.Type())
	}
}
```

这里更改了上一节提到的`applyFuction`这个函数，根据`fn`参数的类型来调用普通函数或者内置函数。

我们需要做的另一处改动是`evalIdentifier`函数：

```go
//eval.go
func evalIdentifier(node *ast.Identifier, scope *Scope) Object {
	if val, ok := scope.Get(node.Value); ok {
		return val
	}

	if builtin, ok := builtins[node.Value]; ok {
		return builtin
	}

	return newError(node.Pos().Sline(), ERR_UNKNOWNIDENT, node.Value)

}
```

第7-9行增加了判断标识符是否是内置函数的判断。如果是内置函数，就返回存储在`builtins` map中存储的函数对象。

再以刚才的例子为例：

```go
sub(1,2)
len("Hello")
```

当解释器解释第一行的时候，遇到`sub`这个标识符的时候，就会去取这个标识符中保存的对象（很显然保存的是函数对象），那么第4行就会返回。

当解释器解释第二行的时候，遇到`len`这个标识符，首先执行第3行的判断，不满足，那么就会执行第7行的判断。

这次就会从builtins这个map中找到`len`这个内置函数，然后将其返回。



内置函数的说明及其实现就算完成了。这一节是我写得最累的一篇文章（想了好几天），因为不知道如何向读者解释的更清楚，让其内容更易于理解。

如果读者看完后，还不是很清楚的话，那就多看几遍，多测试调试一下可能就恍然大悟了。

## 测试

下面我们写一个简单的程序测试一下：
```go
//main.go
func TestEval() {
	tests := []struct {
		input    string
		expected string
	}{
		{"len(\"Hello World\")", "11"},
		{"println(10)", "nil"},
	}

	for _, tt := range tests {
		l := lexer.NewLexer(tt.input)
		p := parser.NewParser(l)
		program := p.ParseProgram()

		scope := eval.NewScope(nil, os.Stdout)
		evaluated := eval.Eval(program, scope)
		if evaluated != nil {
			if evaluated.Inspect() != tt.expected {
				fmt.Printf(%s\n", evaluated.Inspect())
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



下一节，我们会增加条件判断表达式`if-else`的支持。敬请期待！

# `file`内置对象支持

一门语言，要是没有读取或者保存文件的功能，那就太遗憾了。这一节中，我们将追加一个`file`内置对象。

我们要以怎样的方式提供这种对文件的操作呢？一种方式是可以创建一个`文件对象（File Object）`，然后给这个文件对象提供一个`open`函数，类似下面这样：

```go
//file.go
type FileObject struct {
	File    *os.File //保存一个`go`语言的文件指针
}

//实现Object接口
func (f *FileObject) Inspect() string  { return "<file object: " + f.Name + ">" }
func (f *FileObject) Type() ObjectType { return FILE_OBJ }
func (f *FileObject) CallMethod(line string, scope *Scope, method string, args ...Object) Object {
	switch method {
	case "open":
		return f.open(line, args...)
	//...
	default:
		return newError(line, ERR_NOMETHOD, method, f.Type())
	}
}

func (f *FileObject) open(line string, args ...Object) Object {
    f.File, err := os.OpenFile(文件名参数, 打开文件flag, 文件打开权限)
    //...
}
```

另外一种方式是可以创建一个新的名为`open`的`内置函数(Built-in funciton)`。这一节我们使用第二种方式。

既然是创建一个内置方法，那么我们需要在`builtin.go`文件中创建相应的函数。来看一下代码：

```go
//builtin.go
func init() {
	builtins = map[string]*Builtin{
		//...
		"open":    openBuiltin(),
	}
}
```

首先我们给`builtins`这个map增加一个key为`open`，值为`openBuiltin`的函数。

在实现`openBuiltin()`函数之前，让我们提供一个工具（utility）函数：

```go
//object.go
//生成一个元祖(tuple)对象。
//这里面我们默认给元组提供了两个`NIL`成员（我们假设元组至少有两个成员）
func NewTuple(isMulti bool) *Tuple {
	return &Tuple{IsMulti: isMulti, Members: []Object{NIL, NIL}}
}
```

为啥提供这个工具（utility）函数？因为我们的`open`内置函数成功的时候返回文件对象，而在错误的时候，需要返回错误对象（Error Object）。就是说`open`内置函数需要返回两个值：

* 成功返回:   `<File Object>`, NIL
* 失败返回： NIL, `<Error Object>`

如何实现`open`内置函数？实际上利用`go`语言`os`模块的`OpenFile`函数就可以达到目的。先来看一下这个函数的定义：

```go
func OpenFile(name string, flag int, perm FileMode) (*File, error)
//第一个参数是文件名
//第二个flag参数，表示文件以何种方式打开，比如O_RDONLY,表示已只读方式打开
//第三个参数是文件打开时的权限
```

下面我们再来看看`openBuiltin()`函数的实现：

```go
//builtin.go

var fileModeTable = map[string]int{
	"r":   os.O_RDONLY,   //只读
	"<":   os.O_RDONLY,   //只读
	"w":   os.O_WRONLY | os.O_CREATE | os.O_TRUNC,  //只写
	">":   os.O_WRONLY | os.O_CREATE | os.O_TRUNC,  //只写
	"a":   os.O_APPEND | os.O_CREATE, //追加写
	">>":  os.O_APPEND | os.O_CREATE, //追加写
	"r+":  os.O_RDWR, //读写
	"+<":  os.O_RDWR, //读写
	"w+":  os.O_RDWR | os.O_CREATE | os.O_TRUNC, //文件存在则清空文件，不存在则创建,读写方式
	"+>":  os.O_RDWR | os.O_CREATE | os.O_TRUNC, //文件存在则清空文件，不存在则创建,读写方式
	"a+":  os.O_RDWR | os.O_APPEND | os.O_CREATE, //追加，可读写
	"+>>": os.O_RDWR | os.O_APPEND | os.O_CREATE, //追加，可读写
}

func openBuiltin() *Builtin { //返回一个内置函数对象(builtin-function object)
	return &Builtin{ Fn: open_function }
}

func open_function((line string, scope *Scope, args ...Object) Object {
	var fname *String
	var flag int = os.O_RDONLY //默认为只读方式打开
	var ok bool
	var perm os.FileMode = os.FileMode(0666) //默认权限是'666',即可读可写

	//创建一个tuple对象，里面有两个成员：（NIL, NIL)
	tup := NewTuple(true)

	argLen := len(args)
	if argLen < 1 { //判断参数个数，最少要提供文件名
		tup.Members[1] = newError(line, ERR_ARGUMENT, "at least one", argLen)
		return tup
	}

    //文件名参数，必须是个字符串对象
	fname, ok = args[0].(*String)
	if !ok {
		tup.Members[1] = newError(line, ERR_PARAMTYPE, "first", "open", "*String", 
                                  args[0].Type())
		return tup
		}

	if argLen == 2 {
		//第二个参数（如果有），是上面定义的`fileModeTable`这个map的key
		m, ok := args[1].(*String)
		if !ok {
			tup.Members[1] = newError(line, ERR_PARAMTYPE, "second", "open", "*String", 
                                      args[1].Type())
			return tup
		}

		flag, ok = fileModeTable[m.String]
		if !ok {
			tup.Members[1] = newError(line, "unknown file mode supplied")
			return tup
		}
	}

	if len(args) == 3 {
		//第三个参数（如果有），是文件权限
		p, ok := args[2].(*Number)
		if !ok {
			tup.Members[1] = newError(line, ERR_PARAMTYPE, "third", "open", "*Integer", 
                                      args[2].Type())
			return tup
		}
		perm = os.FileMode(int(p.Value))
	}

   	//调用`os`模块的'OpenFile'函数
	f, err := os.OpenFile(fname.String, flag, perm)
	if err != nil { //如果失败
		//返回：（NIL， <错误对象>）这个tuple（元祖）
		tup.Members[1] = newError(line, "'open' failed with error: %s", err.Error())
		return tup
	}

   	//成功返回: (<文件对象>, NIL)这个tuple(元祖)
	tup.Members[0] = &FileObject{File: f, Name: "<file object: " + fname.String + ">"}
	return tup
}

```

第3-16行我们定义了一个名为`fileModeTable`的map，这个map的key是我们熟悉的文件打开方式，比如"r"表示只读（read），"a"表示追加（append）。map的value是`go`定义的那些常量。实现打开文件的逻辑代码在`open_function`函数中。这个函数其实主要做了以下几点：

1. 检查参数个数及参数类型，如果不对就报错，返回`(NIL, <Error对象>)`这个tuple（元祖）
2. 如果全部都成功且`os.OpenFile`调用成功，则返回`(<文件对象>, NIL)`这个tuple（元祖）

上面我们提到了`文件对象(File Object)`，下面就让我们来实现`文件对象`:

```go
//object.go
const (
	//...
	FILE_OBJ         = "FILE"
)
```

第4行增加了一个`FILE_OBJ`的文件类型常量。

下面是`文件对象`的具体代码：

```go
//file.go
type FileObject struct {
	File    *os.File //储存文件指针
	Name    string  //文件名
	Scanner *bufio.Scanner //scanner对象
}

//实现Object接口
func (f *FileObject) Inspect() string  { return "<file object: " + f.Name + ">" }
func (f *FileObject) Type() ObjectType { return FILE_OBJ }
func (f *FileObject) CallMethod(line string, scope *Scope, method string, args ...Object) Object {
	switch method {
	case "close":
		return f.close(line, args...)
	case "read":
		return f.read(line, args...)
	case "readLine":
		return f.readLine(line, args...)
	case "write":
		return f.write(line, args...)
	case "writeString":
		return f.writeString(line, args...)
	case "writeLine":
		return f.writeLine(line, args...)
	case "name":
		return f.getName(line, args...)
	default:
		return newError(line, ERR_NOMETHOD, method, f.Type())
	}
}

func (f *FileObject) close(line string, args ...Object) Object {
	if len(args) != 0 {
		return newError(line, ERR_ARGUMENT, "0", len(args))
	}
	err := f.File.Close()
	if err != nil {
		return newError(line, "'close' failed. reason: %s", err.Error())
	}
	return TRUE
}

//注: 这个方法将返回三个不同的值:
//   1. Error对象 - 读取错误的时候
//   2. NIL      - 遇到EOF的时候
//   3. 字符串对象 - 正常读取的情况
func (f *FileObject) read(line string, args ...Object) Object {
	if len(args) != 1 {
		return newError(line, ERR_ARGUMENT, "1", len(args))
	}

	readlen, ok := args[0].(*Number)
	if !ok {
		return newError(line, ERR_PARAMTYPE, "first", "read", "*Number", args[0].Type())
	}

	buffer := make([]byte, int(readlen.Value))
	n, err := f.File.Read(buffer)
	if err != io.EOF && err != nil {
		return newError(line, "'read' failed. reason: %s", err.Error())
	}

	if n == 0 && err == io.EOF {
		return NIL
	}
	return NewString(string(buffer))
}

func (f *FileObject) readLine(line string, args ...Object) Object {
	if len(args) != 0 {
		return newError(line, ERR_ARGUMENT, "0", len(args))
	}
	if f.Scanner == nil {
		f.Scanner = bufio.NewScanner(f.File)
		f.Scanner.Split(bufio.ScanLines)
	}
	aLine := f.Scanner.Scan()
	if err := f.Scanner.Err(); err != nil {
		return newError(line, "'readline' failed. reason: %s", err.Error())
	}
	if !aLine {
		return NIL
	}
	return NewString(f.Scanner.Text())
}

func (f *FileObject) write(line string, args ...Object) Object {
	if len(args) != 1 {
		return newError(line, ERR_ARGUMENT, "1", len(args))
	}

	content, ok := args[0].(*String)
	if !ok {
		return newError(line, ERR_PARAMTYPE, "first", "write", "*String", args[0].Type())
	}

	n, err := f.File.Write([]byte(content.String))
	if err != nil {
		return newError(line, "'write' failed. reason: %s", err.Error())
	}

	return NewNumber(float64(n))
}

func (f *FileObject) writeString(line string, args ...Object) Object {
	if len(args) != 1 {
		return newError(line, ERR_ARGUMENT, "1", len(args))
	}

	content, ok := args[0].(*String)
	if !ok {
		return newError(line, ERR_PARAMTYPE, "first", "writeString", "*String", args[0].Type())
	}

	ret, err := f.File.WriteString(content.String)
	if err != nil {
		return newError(line, "'writeString' failed. reason: %s", err.Error())
	}

	return NewNumber(float64(ret))
}

func (f *FileObject) writeLine(line string, args ...Object) Object {
	if len(args) != 1 {
		return newError(line, ERR_ARGUMENT, "1", len(args))
	}

	content, ok := args[0].(*String)
	if !ok {
		return newError(line, ERR_PARAMTYPE, "first", "writeLine", "*String", args[0].Type())
	}

	ret, err := f.File.Write([]byte(content.String + "\n"))
	if err != nil {
		return newError(line, "'writeLine' failed. reason: %s", err.Error())
	}

	return NewNumber(float64(ret))
}

func (f *FileObject) getName(line string, args ...Object) Object {
	if len(args) != 0 {
		return newError(line, ERR_ARGUMENT, "0", len(args))
	}

	return NewString(f.File.Name())
}
```

上面的代码看起来不少，其实并不是很复杂。对于`文件对象(File Object)`的那些方法，实际上我们是利用储存在`文件对象`中的`os.File`指针，来调用`go`语言的相应的函数来实现的。里面大部分都是判断参数个数及参数类型是否正确的代码。

创建好了`文件对象`，用户就可以像下面这样调用这个`文件对象`的方法来操作文件了：

```perl
# file.mp
# 调用内置函数'open'，以写追加方式打开文件
# 'open'函数返回一个tuple（元祖）
file, err = open("./file.log", "w+") 
if err { # 如果错误，则报错（实际代码可能还需要提前返回，而不做后续处理。这里为了演示，只是打印了错误信息）
    println(err)
} else {
    # 向文件中写入三行文字
    file.writeLine("This is the first line")
    file.writeLine("This is the second line")
    file.writeString("这是第三行\n")
    file.close() # 关闭文件
}

printf("=====Reading file=====\n")
file, err = open("./file.log", "r") # 以只读方式打开文件
if err {
    println(err)
} else {
    # 读取三行文字，并打印出来
    println(file.readLine())
    println(file.readLine())
    println(file.readLine())
    file.close() # 关闭文件
}
```

有的读者就问了，你这是处理文件，那么如何处理标准输入和标准输出呢？怎么从标准输入读取信息，并从标准输出显示信息呢？

其实解决方法可能比你想象的简单很多。

我们知道，`go`语言提供了`os.Stdin`、`os.Stdout`和`os.Stderr`类型来处理标准输入、标准输出和标准错误，而这三个变量都是`*os.File`，大家再看看我们的`文件对象`的第一个字段是不是也是一个`*os.File`？

`文件对象`结构保存了一个`*os.File`类型的文件指针，我们只需要把`os.Stdin`、`os.Stdout`和`os.Stderr`赋值给这个指针就可以了。下面是相关代码：

```go
//object.go
func initGlobalObj() {
    //预定义三个全局Scope： `stdin`, `stdout`, `stderr`
	SetGlobalObj("stdin", &FileObject{File: os.Stdin})
	SetGlobalObj("stdout", &FileObject{File: os.Stdout})
	SetGlobalObj("stderr", &FileObject{File: os.Stderr})
}

func init() {
    //在'init'方法中初始化
	initGlobalObj()
}
```

我们在全局Scope中保存了三个`文件对象`（分别是stdin、stdout和stderr），这三个`文件对象`分别存储了标准输入、标准输出和标准错误三个文件指针。

现在来看一下我们如何使用标准输入和标准输出：

```perl
# demo.mp
print("Please type your name:")

# 调用文件对象的`readLine`方法
name = stdin.readLine()

# 调用文件对象的`writeLine`方法
stdout.writeLine("Hello " + name + "!") 
```

本节，我们实现了`文件对象`，让用户能够操作文件。同时也提供了标准输入、标准输出的操作。

> 我们只给这个`文件对象`提供了很少的方法，如果愿意的话，当然还可以提供更多的文件操作方法。实现起来无非就是主要的三点：
>
> 1. 判断参数个数，不对报错
> 2. 检查参数类型，不对报错
> 3. 调用`go`语言的相应函数

下一节，我们再提供一个`os`内置对象，让读者加深一下理解。

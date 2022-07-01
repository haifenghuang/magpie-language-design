# `os`内置对象支持

这一节中，我们将追加一个`os`内置对象，有了上一节和本节的内容，读者应该能够学会自己追加内置对象了。

我们新建一个`os.go`文件：然后追加如下代码：

```go
//os.go
package eval

import (
	_ "fmt"
	"os"
)

const (
	OS_OBJ  = "OS_OBJ" //对象类型
	os_name = "os"
)

//os对象
type Os struct{}

func (o *Os) Inspect() string  { return "<" + os_name + ">" }
func (o *Os) Type() ObjectType { return OS_OBJ }

func (o *Os) CallMethod(line string, scope *Scope, method string, args ...Object) Object {
	switch method {
	case "getenv":
		return o.getenv(line, args...)
	case "setenv":
		return o.setenv(line, args...)
	case "chdir":
		return o.chdir(line, args...)
	case "mkdir":
		return o.mkdir(line, args...)
	case "exit":
		return o.exit(line, args...)
	}
	return newError(line, ERR_NOMETHOD, method, o.Type())
}

func (o *Os) getenv(line string, args ...Object) Object {
	if len(args) != 1 {
		return newError(line, ERR_ARGUMENT, "1", len(args))
	}

	key, ok := args[0].(*String)
	if !ok {
		return newError(line, ERR_PARAMTYPE, "first", "getenv", "*String", args[0].Type())
	}

	ret := os.Getenv(key.String)
	return NewString(ret)
}

func (o *Os) setenv(line string, args ...Object) Object {
	if len(args) != 2 {
		return newError(line, ERR_ARGUMENT, "2", len(args))
	}

	key, ok := args[0].(*String)
	if !ok {
		return newError(line, ERR_PARAMTYPE, "first", "setenv", "*String", args[0].Type())
	}

	value, ok := args[1].(*String)
	if !ok {
		return newError(line, ERR_PARAMTYPE, "second", "setenv", "*String", args[1].Type())
	}

	err := os.Setenv(key.String, value.String)
	if err != nil {
		return FALSE
	}
	return TRUE
}

func (o *Os) chdir(line string, args ...Object) Object {
	if len(args) != 1 {
		return newError(line, ERR_ARGUMENT, "1", len(args))
	}

	newDir, ok := args[0].(*String)
	if !ok {
		return newError(line, ERR_PARAMTYPE, "first", "chdir", "*String", args[0].Type())
	}

	err := os.Chdir(newDir.String)
	if err != nil {
		return FALSE
	}
	return TRUE
}

func (o *Os) mkdir(line string, args ...Object) Object {
	if len(args) != 2 {
		return newError(line, ERR_ARGUMENT, "2", len(args))
	}

	name, ok := args[0].(*String)
	if !ok {
		return newError(line, ERR_PARAMTYPE, "first", "mkdir", "*String", args[0].Type())
	}

	perm, ok := args[1].(*Number)
	if !ok {
		return newError(line, ERR_PARAMTYPE, "second", "mkdir", "*Number", args[1].Type())
	}

	err := os.Mkdir(name.String, os.FileMode(int64(perm.Value)))
	if err != nil {
		return FALSE
	}
	return TRUE
}

func (o *Os) exit(line string, args ...Object) Object {
	if len(args) != 0 && len(args) != 1 {
		return newError(line, ERR_ARGUMENT, "0|1", len(args))
	}

	if len(args) == 0 {
		os.Exit(0)
		return NIL
	}

	code, ok := args[0].(*Number)
	if !ok {
		return newError(line, ERR_PARAMTYPE, "first", "exit", "*Number", args[0].Type())
	}

	os.Exit(int(code.Value))

	return NIL
}

//工具(utility)函数
func NewOsObj() Object {
	ret := &Os{} //创建一个`os`对象
	SetGlobalObj(os_name, ret) //将创建的os对象加入key为`os`的全局Scope中

	return ret
}
```

上面的代码看起来不少，其实并不是很复杂，主要功能无非就两个：

1. 判断参数个数及类型，有问题就报错
2. 调用`go`语言的`os package`中的相应方法

这里需要关注的主要是132-137行的`NewOsObj()`函数，我们将创建的`os对象`加入了key为`"os"`的全局Scope中。

然后我们需要在`init`方法中调用它，使它对我们的系统可用：

```go
//object.go
func init() {
	initGlobalObj()

	NewOsObj()
}
```

第5行代码调用了`NewOsObj()`函数。有了这个调用后，我们写的程序就可以像下面这样调用这些`os`对象的方法了：

```perl
# os.mp
println("====ENVIRONMENT[PATH]====")
println(os.getenv("PATH"))

# ...
os.exit(0)
```



至此，我们完成了第二部分的讲解。在第三部分，我们会讲解一些更为高级的话题。这些话题能够让我们的小喜鹊（`magpie`）如虎添翼。

第三部分的第一节，我们将加入`struct`的支持。


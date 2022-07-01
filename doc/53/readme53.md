# `标准库（standard library）`支持

这一节中，我们会增加对`标准库`的支持。一门语言，如果没有标准库的支持，那么它的功能就会大大受限。而通过前面的学习，我们的`magpie`脚本语言已经能够正确的导入用户自定义的模块。这一节无非就是加入语言自带的标准模块。所以这一节应该学起来比较轻松。

> 学习之前，请确保你机器中安装的go版本至少是1.16的。

对于标准库，我们这里介绍的方法是将标准模块嵌入到执行程序中。这里利用了`go 1.16`的`go:embed`功能。

我在`magpie/parser`目录下放了一个`lib`目录，里面有两个标准库文件。目录结构如下：

```
src
|_ magpie
   |_parser
     |_linq.mq
     |_str.mq
```

我们需要做的就是把`linq.mq`和`str.mq`文件嵌入到执行程序中。读者可能知道，我们解析`import`语句是在`parser.go`文件中，因此我们做的更改也是在这个文件中。

首先，在`parser.go`文件中加入下面的几行语句：

```go
//parser.go
import (
    _ "embed" //因为要用到'go:embed'的功能，必须导入这个模块
	//...
)

//go:embed lib/str.mp
var libstr []byte

//go:embed lib/linq.mp
var liblinq []byte

var stdlibs = map[string][]byte{
	"linq": liblinq,
	"str":  libstr,
}

func isStdLib(name string) bool {
	_, ok := stdlibs[name]
	return ok
}
```

请注意第7行和第10行的注释，这里是告诉`go`编译器，我们希望将`lib/str.mp`和`lib/linq.mp`文件嵌入到执行程序中。

第8行告诉编译器，在程序中可以使用`libstr`字节数组来直接访问嵌入的`lib/str.mp`文件的内容。第11行是同样的道理。

第13行我们定义了一个`stdlibs`的map变量，key是模块名，value就是嵌入的文件的字节数组。当我们的脚本中包含`import linq`或者`import str`语句的时候，我们就可以直接通过value来访问`linq.mp`文件和`str.mp`文件的内容。第18行的`isStdLib`辅助函数，用来判断脚本中导入的模块是不是标准库。



读者应该知道，解析器中实际分析`import`语句的函数是`getImportedStatements`，我们需要对其做一定的更改：

```go
func (p *Parser) getImportedStatements(importpath string) (*ast.Program, error) {
	var f []byte
	var fn string
	if isStdLib(importpath) {
		if imported, ok := p.importLib[importpath]; ok {
			return imported, nil
		}

		f, _ = stdlibs[importpath] //将嵌入到执行程序中的标准库文件的内容取出
		fn = importpath + ".mp"
	} else {
		//既有逻辑
    }

	l := lexer.NewLexer(string(f))
	l.Filename = fn

	ps := NewParser(l)
	ps.Attachments = p.Attachments
	parsed := ps.ParseProgram()
	if len(ps.errors) != 0 {
		p.errors = append(p.errors, ps.errors...)
		p.errorLines = append(p.errorLines, ps.errorLines...)
	}

	if isStdLib(importpath) {
		p.importLib[importpath] = parsed
	}

	return parsed, nil
}
```

其中，代码第4-10行的`if`判断是新追加的。第4行我们判断如果`import xxx`后面的`xxx`是标准模块的话，我们会接着在第5行，判断这个标准模块是否已经导入过。如果已经导入过的话，就直接返回解析过的代码。第26行的`if`判断也是新追加的。如果是标准库模块，我们将其解析过的代码存到`importLib`变量中。

这个`importLib`变量也是新追加的：

```go
//parser.go
type Parser struct {
	//...
	importLib   map[string]*ast.Program //用来保存标准库中已经解析过的代码
}

func NewParser(l *lexer.Lexer) *Parser {
	p := &Parser{
		l:          l,
		errors:     []string{},
		errorLines: []string{},
		importLib:  make(map[string]*ast.Program), //需要初始化importLib变量
	}

	//...
	return p
}
```

以上就是所有需要更改的地方。不敢相信是吧？就是这么简单。



下面简单说一下这个标准库存放的位置。由于`go`语言的`embed`包实现上的限制，我们必须将标准库的代码（即`linq.mp`和`str.mp`）放到`parser.go`这个文件所在的目录中，这样`go`的编译器才能够找到嵌入的文件。我曾经试着通过将标准库放到项目根目录的`lib`文件夹下，结果`go`的编译器报告找不到嵌入的文件：

```go
//parser.go
//go:embed ../../../../lib/str.mp
var libstr []byte

//go:embed ../../../../lib/linq.mp
var liblinq []byte
```

事实上放到哪里无关紧要，只要将标准库文件嵌入到执行程序后，这些标准库文件完全可以删除了。但是一般的做法是将其拷贝到项目根目录的`lib`文件夹下，这样使用者就能够查看标准库的一些实现逻辑。

还有一点，需要注意的是，`go:embed`将文件嵌入到执行程序中，是以原始方式嵌入进去的，不会进行任何加密、或者编码之类的附加操作。因此如果希望嵌入的文件也进行加密之类的处理的话，可以使用`github`上的一个开源库：https://github.com/FiloSottile/age。



## 测试

```javascript
import str
import linq

//str standard lib
println(IsUpper("h"))
println(IsUpper("H"))
println(StrReverse("Hello"))

//linq standard lib
result = Linq([1,2,3,4,5,6,7,8,9,10])
	.Where(x => x % 2 == 0)
	.Select(x => x + 1)
	.Reverse()
	.ToArray()
printf("result = %s\n", result)
```



终于完成了，可以喘口气了。这是本系列最后的一篇文章。

本来预计这一系列文章写个20来篇就结束了，结果越写越多，这可苦了我这个不善于写文章的懒人。即使写了这么多，也还有许多可以讲解的话题。就到这吧，实在是累了。为了写这个，多少个不眠之夜，对于我这个已经要步入50岁的码农一枚，其中的辛酸谁又能够知道、了解呢。

相信通过这一系列的学习，读者已经了解了自制语言的基本步骤，甚至可以自行编写自制语言了。希望本系列文章能够给读者带去哪怕一丝丝的自豪感，我也倍感荣幸！！!


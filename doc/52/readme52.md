# `单一执行文件(single standalone exe)`支持

这一节中，我们会增加对`单一可执行文件`的支持。有的读者可能会问，这个是什么意思？先简单看个例子就知道了。这里假设我们的`magpie`解释器的执行程序名为`magpie`（这里以`linux`为例来说明）。当我们需要运行脚本的时候，命令行大致如下：

```shell
$ ./magpie examples/xxx.mp
```

而如果我们有了单一可执行文件，那么我们就可以像下面这样运行脚本程序：

```shell
$ ./magpie
```

不需要提供脚本文件名了。那么又有读者会问了，脚本文件到哪里去了？当有多个脚本文件需要解释（`Evaluate`）的时候，应用程序怎么知道运行的主脚本文件是哪个？这个问题，接下来内容，我们会逐个解决。

单个可执行文件，有啥好处呢？其实好处还是很多的。例如发布简单，能够和一些工作流(workflow)很好的配合。现在的一些工具也在朝着这个方向发展。例如和`Deno compiler`和`PyOxidizer`等。

那如何实现将脚本文件打包（或者嵌入）进执行程序并运行呢？实际上我们可以使用`go 1.16`版本及之后提供的`go:embed`方式，来将文件打包进执行程序。但是使用`go:embed`的话有一些局限性。比如无法判断执行程序是否已经包含了打包的文件，也不能够对打包的文件进行压缩、编码之类的操作。所以这里我使用了`github`网站上的一个开源库`ember`(https://github.com/maja42/ember)，并对这个开源库做了相应的更改。

> 其实github上关于将资源文件打包到二进制文件的仓库非常之多。

它的实现原理如下：将需要打包的文件嵌入到可执行文件的末尾。为了之后能够取出打包的文件的内容，它加入了所谓的`marker`来确保读入了正确的信息。来看一下打包后它的执行程序的构造：

```
   +---------------+
   |    original   |
   |   executable  |
   +---------------+
   | marker-string |
   +---------------+
   |      TOC      |
   +---------------+
   | marker-string |
   +---------------+
   |     file1     | 
   +---------------+
   |     file2     |
   +---------------+
   |     file3     |
   +---------------+
```

> TOC: Table Of Content（目录），可以理解为Header。主要储存文件的大小（`size`）及偏移量（`offset`），用来快速取出相应的文件内容。

就是将执行程序末尾加入相应的mark，取文件内容的时候，先查找`marker-string`，找到了`marker-string`才表明执行程序是嵌入了文件的可执行程序。之后，读取`TOC`的内容，来找到相应的资源文件数据。

这个开源库将文件嵌入到可执行程序的末尾，使用的是原始嵌入，没有加入压缩及编码等方式。因此我对这个开源库中的代码做了相应的更改，加入了`gzip`压缩及`base64`编码，修改后的执行程序构造如下：

```
   +---------------+
   |    original   |
   |   executable  |
   +---------------+
   | marker-string |
   +---------------+
   |      TOC      |
   +---------------+
   | marker-string |
   +---------------+
```

修改之前的`TOC`内容如下：

```go
type Attachment struct {
    Name string // 资源名(Resource name)
    Size int64  // 资源大小(Resource size in bytes)
}
```

而修改之后的`TOC`内容如下：

```go
type Attachment struct {
	Name string // 资源名(Resource name)
    Data string // 资源内容(经过base64编码及其gzip加密)
}
```

最终嵌入执行文件中的文件内容如下：

```
magpie19770903magpie19770903[{"Name":"main","Data":"H4sIAAAAAAAC/9xZ3W/iOBB/plL/h+lbstuy9LmlUm/vQ5V6Oul2T/uA0CkbJstAMNRxVqAu//vJH3FsCG..."}]magpie19770903magpie19770903
```

最前面和最后面的`magpie19770903magpie19770903`就是`marker-string`。中间的部分就是`TOC`。

在修改前，我们需要进入ember包的`cmd\embedder`目录，执行`go build`命令，生成一个名为`embedder.exe`的执行程序。通过这个执行程序来将我们的`magpie`解释器和需要打包的文件，打包成一个新的执行程序。运行`embedder.exe`执行程序前，需要提供一个`json`文件，将需要打包的文件放入这个`json`文件中，像下面这样：

```json
//attachments.json
{
  "main": "embed/linq_demo.mp",
  "linq/linq": "embed/linq/linq.mp"
}
```

其中`json`的value就是需要打包的文件，而key就是程序中用来访问文件用的。

运行命令如下（以linux为例）：

```shell
 rm -f ./linq_demo
./embedder -attachments ./attachments.json -exe ./magpie -out ./linq_demo
```

第2行的命令会将`magpie`解释器（执行程序）的最后嵌入`attachments.json`中指定的文件内容，最后生成一个新的单一执行程序`linq_demo`。

对于`attachments.json`文件，我设定了一个强制要求。就是主脚本文件必须使用`main`作为key。为什么这么规定呢？因为这个是`json`，取出来后它的key的顺序是不固定的，这样解释器就无法知道哪个是需要运行的主脚本程序。同时还需要注意`attachments.json`中的第二个key，即`linq/linq`。假设我们的主脚本程序如下：

```go
//linq_demo.mp
import linq.linq

result = linq.Linq([1,2,3,4,5,6,7,8,9,10])
	.Where(x => x % 2 == 0)
	.Select(x => x + 1)
	.Reverse()
	.ToArray()
printf("[1,2,3,4,5,6,7,8,9,10] where(x %% 2 == 0) = %s\n", result)
```

第2行的导入语句为`import linq.linq`，则我们需要在`attachments.json`文件中，把`.`更换为`/`。



介绍了这么多，来看一下对代码所作的更改吧。首先，我们需要修改`main`函数：

```go
//main.go
import (
	//...
	"github.com/maja42/ember" //引入上面提到的ember包
)

func main() {
	if runWithEmbedFile() {
		return
	}

	//既有旧代码
}
```

第8行，我们增加了一个`runWithEmbedFile`方法，来看一下此方法的实现：

```go
//main.go
func runWithEmbedFile() bool {
    //打开执行程序，查找`marker-string`,并读取其中嵌入的文件(即TOC中的内容)
	attachments, err := ember.Open()
	if err != nil {
		return false
	}
	defer attachments.Close()
 
	//contents（[]string类型）的内容就是前文提到的'attachments.json'文件的所有key的列表
	contents := attachments.List()
	if len(contents) == 0 { //判断是否有嵌入的文件，没有的话，返回false
		return false
	}

	name := "main"
	foundMain := false
	for _, content := range contents { //查找主脚本文件
		if content == name {
			foundMain = true
			break
		}
	}
	if !foundMain {
		return false
	}

	buf, err := attachments.GetResource(name) //根据name（这里是main）,取出其对应的文件内容
	if err != nil {
		fmt.Printf("error reading embedded file.\n", err)
		return false
	}

    //下面都是大家熟悉的代码
	str := string(buf)
	l := lexer.NewLexer(str)
	p := parser.NewParser(l)
    p.Attachments = attachments //将attachments赋值给语法解析器(parser)的Attachments
	program := p.ParseProgram()
	if len(p.Errors()) != 0 {
		for _, err := range p.Errors() {
			fmt.Println(err)
		}
		os.Exit(1)
	}
	scope := eval.NewScope(nil, os.Stdout)

	result := eval.Eval(program, scope)
	if result.Type() == eval.ERROR_OBJ {
		fmt.Println(result.Inspect())
		os.Exit(1)
	}

	return true
}
```

对于第38行的`p.Attachments = attachments`语句，我们稍后会讲到。

这里面需要强调的是第28行的`GetResource`函数，它的实现如下（这个是我新写的函数）：

```go
//attachement.go
func (a *Attachments) GetResource(name string) ([]byte, error) {
	for _, item := range a.toc {
		if item.Name == name {
			var raw bytes.Buffer
			var err error

			// 解码数据（Decode the data）
			in, err := base64.StdEncoding.DecodeString(item.Data)
			if err != nil {
				return nil, err
			}

			// Gunzip解压缩
			gr, err := gzip.NewReader(bytes.NewBuffer(in))
			if err != nil {
				return nil, err
			}
			defer gr.Close()
			data, err := ioutil.ReadAll(gr) //读取解压后的数据
			if err != nil {
				return nil, err
			}
            _, err = raw.Write(data)//将解压后的数据('data')写入buffer(raw的类型为bytes.Buffer)
			if err != nil {
				return nil, err
			}

			// 返回byte数组
			return raw.Bytes(), nil
		}
	}

	return nil, newAttErr("could not found resource with name '%w'", name)
}
```

代码第3行中的`a.toc`实际上是一个结构：

```go
//toc.go
type Attachment struct {
	Name string //资源名
	Data string //资源对应的数据（base64编码，gzip压缩）
}
```

`main.go`文件的改修算是完成了。但是从上面的代码中可以看到，`main.go`文件中只是处理的主脚本文件。如果主脚本文件中还引入（`import`）了其它的脚本文件，那么我们还需要处理其它的脚本文件。因此我们需要修改`parser.go`文件中实际处理`import`语句的函数`getImportedStatements`。

```go
//parser.go
import (
	//...
	"github.com/maja42/ember"
)

//...
type Parser struct {
	//...
	Attachments *ember.Attachments
}

func (p *Parser) getImportedStatements(importpath string) (*ast.Program, error) {
	//...
	fn := filepath.Join(path, importpath+".mp")
	f, err := ioutil.ReadFile(fn)
	if err != nil { //如果读取文件失败（没有找到文件）
		// 检查'MAGPIE_ROOT'环境变量
		importRoot := os.Getenv("MAGPIE_ROOT")
		if len(importRoot) == 0 { //'MAGPIE_ROOT'环境变量没有设置
			//检查嵌入文件
			if p.Attachments == nil {
				return nil, fmt.Errorf("Syntax Error:%v- no file or directory: %s.mp, %s", 
                                       p.curToken.Pos, importpath, path)
			}

			//在嵌入文件中进行查找
			contents := p.Attachments.List()
			iFound := false
			for _, name := range contents {
				if name == importpath { //这里‘importpath’变量的值为`xxx/xxx`
					iFound = true
					break
				}
			}
			if !iFound {
				return nil, fmt.Errorf("Syntax Error:%v- no file or directory: %s.mp, %s", 
                                       p.curToken.Pos, importpath, path)
			}

			//找到，则获取嵌入文件的内容
			buf, err := p.Attachments.GetResource(importpath)
			if err != nil {
				return nil, fmt.Errorf("Syntax Error:%v- no file or directory: %s.mp, %s",
                                       p.curToken.Pos, importpath, path)
			}
			f = buf
		} else {
			//既有代码
		}
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
	return parsed, nil
}
```

代码22-47行是新增的处理嵌入文件的逻辑。

修改的逻辑就处理完成了。至此我们将脚本文件打包到了执行程序，并可以正常的运行这个单一执行文件了。



有的读者就会问了，你这里的做法是将资源文件嵌入到可执行文件的末尾，那么不就破坏了执行文件的结构了吗？执行文件还能够正常运行吗？据测试，这种做法对于主流的几种可执行文件类型（如windows的`PE`，Linux的`ELF`格式）都没有影响。而且现在流行的`Deno（A modern runtime for JavaScript and TypeScript）`，也提供了本文类似的做法。



其实通过这种做法，我们也可以做一个类似的独立的安装程序，然后将需要的文件嵌入到安装程序中。是不是很酷！



下一节，我们增加`标准库（standard library）`支持。




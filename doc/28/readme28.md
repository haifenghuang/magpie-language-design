# `import`支持

在这一节中，我们要加入对于`import`的支持。有了这个支持，我们就可以导入别人写的脚本代码或者共通的脚本代码。先来看一下使用例：

```python
# import.mp
import sub_package.calc

Add(1, 2) # 调用calc.mp文件中导入的Add方法

```

这里假设我们的脚本（`import.mp`）所在的目录的路径结构如下：

``` 
import.mp
examples目录
  |__ sub_package目录
  |    |__ calc.mp
```



和之前一样，我们来看一下需要做的更改：

1. 在词元（Token）源码`token.go`加入新增的词元类型（`TOKEN_IMPORT`）。
2. 在抽象语法树（AST）的源码`ast.go`中加入`import`语句对应的抽象语法表示。
3. 在语法解析器（Parser）的源码`parser.go`中增加对`import`语句的解析。
5. 在解释器（Evaluator）的源码`eval.go`中加入对`import语句`的解释。



## 词元（Token）的更改

```go
//token.go

const (
	//...
	TOKEN_IMPORT // import
)

//词元类型的字符串表示
func (tt TokenType) String() string {
	switch tt {
	//...
	case TOKEN_IMPORT:
		return "IMPORT"
	}
}

var keywords = map[string]TokenType{
	//...
	"import": TOKEN_IMPORT,
}
```



## 抽象语法树（AST）的更改

我们来想一下`import语句`的抽象语法表示，需要哪些信息：

1. 词元（用来调试或者报错用）
2. 导入路径`importPath`
3. 文件导入后，经过语法解析后得到的`Program`节点。

```go
//ast.go
//程序节点
type Program struct {
	Statements []Statement
	Imports    map[string]*ImportStatement //保存脚本中用到的所有`import`语句，以便统一处理
}

//import语句: 
//    import xxx.xxx.xxx
//    import xxx.xxx.xxx as xxx
type ImportStatement struct {
	Token      token.Token
	ImportPath string //这个ImportPath只保存了`abc.def.ghi`中的'ghi',即最后一个标识符
	Program    *Program //经过语法解析后得到的`Program`节点。
}

func (is *ImportStatement) Pos() token.Position {
	return is.Token.Pos
}

func (is *ImportStatement) End() token.Position {
	length := utf8.RuneCountInString(is.ImportPath)
	return token.Position{Filename: is.Token.Pos.Filename, 
                          Line: is.Token.Pos.Line, Col: is.Token.Pos.Col + length}
}

func (is *ImportStatement) statementNode()       {}
func (is *ImportStatement) TokenLiteral() string { return is.Token.Literal }
func (is *ImportStatement) String() string {
	var out bytes.Buffer

	out.WriteString(is.TokenLiteral())
	out.WriteString(" ")
    
	//这个实际是有问题的，因为ImportPath只保存了`abc.def.ghi`中的'ghi',即最后一个标识符。
	//有兴趣的读者可以自行解决。无非就是在结构中多加一个字段，用来保存'abc.def.ghi'的所有内容
	out.WriteString(is.ImportPath)

	return out.String()
}
```

需要注意的是第5行，我们在`程序（Program）`节点结构中加入了一个`Imports`字段，用来表示程序中的所有遇到的`import语句`。为啥需要这个字段呢？假设你的脚本代码中`import`了好几个模块，像下面这样：

```python
import xxx.xxx.xxx
import yyy.yyy.yyy
```

或者像下面这样：

```python
import xxx.xxx.xxx

# some code

import yyy.yyy.yyy
```

在解释（Evaluating）阶段，我们希望统一集中解释这些`import`，而不是遇到一个解释一个。

> 这种处理方式也有不好之处，就是如果我们的脚本代码中`import`语句写在文件中比较靠后的位置，那么其实代码前半部分是无需处理这个`import`导入的。



## 语法解析器（Parser）的更改

首先我们在`parseStatement()`函数中增加一个处理`import语句`的分支：

```go
//parser.go

func (p *Parser) parseStatement() ast.Statement {
	switch p.curToken.Type {
	case token.TOKEN_IMPORT:
		return p.parseImportStatement()
	//...
	}
}
```

函数`parseImportStatement`的实现如下：

```go
//parser.go
//解析import语句:
//     import xxx.xxx.xxx as xxx
//     import xxx.xxx.xxx as xxx as xxx
func (p *Parser) parseImportStatement() *ast.ImportStatement {
	stmt := &ast.ImportStatement{Token: p.curToken}

	p.nextToken()

	paths := []string{}
	paths = append(paths, p.curToken.Literal)

	for p.peekTokenIs(token.TOKEN_DOT) { //如果是"."就继续
		p.nextToken() //忽略当前词元
		p.nextToken() //忽略'.'词元
		paths = append(paths, p.curToken.Literal)
	}

    path := strings.TrimSpace(strings.Join(paths, "/")) //将"."转换为"/"(即变成目录分隔符)
	stmt.ImportPath = filepath.Base(path) //取'xxx/xxxx/name'中的'name',即取basename

	program, err := p.getImportedStatements(path)
	if err != nil {
		p.errors = append(p.errors, err.Error())
		p.errorLines = append(p.errorLines, p.curToken.Pos.Sline())
		return stmt
	}

	if p.peekTokenIs(token.TOKEN_SEMICOLON) {
		p.nextToken()
	}

	stmt.Program = program
	return stmt
}

func (p *Parser) getImportedStatements(importpath string) (*ast.Program, error) {
	var path string
	if p.l.Filename == "" { //当文件名为空的时候（给词元分析器传递的是字符串而非文件）
		path, _ = os.Getwd() //获取当前目录
	} else {
		path, _ = filepath.Abs(p.l.Filename) //获取脚本所在的绝对路径(这里包含文件名)
		path = filepath.Dir(path) //获取路径(去除文件名)
	}

    path, _ := filepath.Abs(p.l.Filename)
    path = filepath.Dir(path)

	fn := filepath.Join(path, importpath+".mp") //默认模块名以'.mp'结尾
	f, err := ioutil.ReadFile(fn) //读取模块内容
	if err != nil { //文件读取错误（可能是文件没有找到）
		// 检查是否设置了'MAGPIE_ROOT'环境变量
		importRoot := os.Getenv("MAGPIE_ROOT")
		if len(importRoot) == 0 { //没找到,报错
			return nil, fmt.Errorf("Syntax Error:%v- no file or directory: %s.mp, %s", 
                                   p.curToken.Pos, importpath, path)
		} else {
			fn = filepath.Join(importRoot, importpath+".mp") //从环境变数所在的目录中查找
			e, err := ioutil.ReadFile(fn)
			if err != nil {
				return nil, fmt.Errorf("Syntax Error:%v- no file or directory: %s.mp, %s",
                                       p.curToken.Pos, importpath, importRoot)
			}
			f = e
		}
	}

	l := lexer.NewLexer(string(f)) //词法分析
    l.Filename = fn //给词法解析器设置文件名（将来报错的时候，能够汇报相应的文件名）

	ps := NewParser(l)
	parsed := ps.ParseProgram() //语法解析
	if len(ps.errors) != 0 {
		p.errors = append(p.errors, ps.errors...)
		p.errorLines = append(p.errorLines, ps.errorLines...)
	}
	return parsed, nil
}
```

注意，`import`语句中的分割符是`.`而不是`/`，所以我们在第19行将其转换成了目录分隔符`/`，这样我们就可以读取对应目录下的文件。

最后，`getImportedStatements()`函数中，我们默认模块名是`.mp`结尾，如果目录下的文件不是`.mp`结尾，则会报告文件找不到之类的错误。这个相当于硬性规定，模块名必须以`.mp`结尾。而如果不是模块的话，不必是`.mp`结尾。例如在`linux/unix`下，当前目录下有个文件是`demo.xyz`，你可以使用`magpie ./demo.xyz`来运行这个`demo.xyz`脚本。

>  windows下的运行命令：`magpie.exe demo.xyz`

由于我们在`程序(Program)`节点中加入了`Imports`这个字段，所以我们还需要更改`parseProgram`方法：

```go
//parser.go
func (p *Parser) ParseProgram() *ast.Program {
	program := &ast.Program{}

	program.Statements = []ast.Statement{}
	program.Imports = make(map[string]*ast.ImportStatement)

	for p.curToken.Type != token.TOKEN_EOF {
		stmt := p.parseStatement()
		if stmt != nil {
			if importStmt, ok := stmt.(*ast.ImportStatement); ok { //判断是否是'import'语句
				importPath := importStmt.ImportPath
				_, ok := program.Imports[importPath]
				if !ok { //在map中不存在，则保存到map中
					program.Imports[importPath] = importStmt
				}
			} else {
				program.Statements = append(program.Statements, stmt)
			}
		}
		p.nextToken()
	}

	return program
}
```

11-17行的`if`判断是新增的代码。11行我们判断`parseStatement`的返回值是否是一个`import`语句，如果是的话，我们会判断这个`import`语句中的模块是否之前导入过，如果导入过的话，就不再继续导入。



## 解释器（Evaluator）的更改

我们需要给`Eval()`函数追加一个处理`import`语句的`case`分支：

```go
//eval.go

func Eval(node ast.Node, scope *Scope) (val Object) {
	switch node := node.(type) {
	case *ast.ImportStatement:
		return evalImportStatement(node, scope)
    }
}
```

下面是`evalImportStatement`函数的实现：

```go
//eval.go
var importMap map[string]*Scope = map[string]*Scope{}

func evalImportStatement(i *ast.ImportStatement, scope *Scope) Object {
	//查看是否存在于importMap中
	if importedScope, ok := importMap[i.ImportPath]; ok {
		importedScope.GetAllExported(scope)//将importedScope中的所有保存的对象放入scope中
		return NIL
	}

	newScope := NewScope(nil, scope.Writer) //创建一个新的scope
	v := evalProgram(i.Program, newScope)
	if v.Type() == ERROR_OBJ {
		return newError(i.Pos().Sline(), ERR_IMPORT, i.ImportPath)
	}

	//将newScope设置到importMap中
	importMap[i.ImportPath] = newScope
	newScope.GetAllExported(scope) //将newScope中得所有保存的对象输出到scope中

	return NIL
}
```

我们在第2行加入了一个`importMap`变量，它的key是`import xxx`中的`xxx`。value是执行导入语句后的新`Scope`。这里需要注意的是第19行的代码，`GetAllExported`函数将`newScope`中的所有存储的哈希键值对拷贝到当前的`scope`变量当中，但是只有大写字母开头的变量/函数被导出。来看一下它的实现：

```go
//scope.go
//将自身保存的对象放入anotherScope中
func (s *Scope) GetAllExported(anotherScope *Scope) {
	for key, value := range s.store {
		if unicode.IsUpper(rune(key[0])) { //仅大写字母开头的函数/变量被导出
			anotherScope.Set(key, value)
		}
	}
}
```



因为`程序（Program）`节点中保存了所有的`import`语句，所以我们还需要更改`evalProgram`函数：

```go
//eval.go
func evalProgram(program *ast.Program, scope *Scope) (results Object) {
	if len(program.Imports) { //如果有import语句
		results = loadImports(program.Imports, scope)
		if results.Type() == ERROR_OBJ {
			return
		}
	}
	//...
}

func loadImports(imports map[string]*ast.ImportStatement, scope *Scope) Object {
	for _, p := range imports { //循环遍历所有的import语句
		v := evalImportStatement(p, scope)
		if v.Type() == ERROR_OBJ {
			return newError(p.Pos().Sline(), ERR_IMPORT, p.ImportPath)
		}
	}
	return NIL
}
```

第4行的`loadImports`函数用来统一处理所有`import`语句。`loadImports`函数的代码非常简单，逐个遍历`import`语句，然后调用`evalImportStatement()`函数来处理实际的导入逻辑。

上面代码的19行，我们使用了一个`ERR_IMPORT`变量，它定义在`errors.go`文件中：

```go
//errros.go
var (
	//...
	ERR_IMPORT  = "import error: %s"
)
```



## 测试

下面我们写一个简单的程序测试一下：
```go
package main

import (
	"fmt"
	"magpie/eval"
	"magpie/lexer"
	"magpie/parser"
	"magpie/token"
	"os"
)

func TestEval() {
	tests := []struct {
		input    string
		expected string
	}{
		{`import sub_package.calc; println(Add(2,3))`, "nil"},
		{`import sub_package.calc; println(_add(2,3))`, "error"}, //_add不是大写字母开头，所以这行会报错
	}

	for _, tt := range tests {
		l := lexer.NewLexer(tt.input)
		p := parser.NewParser(l)
		program := p.ParseProgram()
		if len(p.Errors()) != 0 {
			for _, err := range p.Errors() {
				fmt.Println(err)
			}
			break
		}

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



到目前为止，我们的脚本代码都是只能是字符串，这个很不方便。下一节，我们会重写`main函数`，使其能够从文件中读取脚本代码。


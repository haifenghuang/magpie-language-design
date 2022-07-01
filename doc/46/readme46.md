# `字符串变量内插`支持

这一节中，我们会给`magpie`语言加入`字符串变量内插（Interpolated String variable）`的功能。有的读者可能不明白这个意思。说简单点就是允许在字符串中包含变量。而解释的时候会将变量替换为其实际值。举个例子就比较好理解了：

```go
ax = "hello"
bx = 1024
println("ax = $ax, bx = $bx") //输出：ax = hello, bx = 1024
//println("ax = ${ax}, bx = ${bx}") //等价于上面的语句
```

从上面的例子中可以看出，字符串变量内插以`$var`或者`${var}`的形式来表示。

> 注：本节讲解的字符串变量内插，只支持简单变量的内插，而不支持复杂的表达式内插。例如下面的表达式内插是不支持的：
>
> ```go
> s = "hello"
> println("s.upper = ${s.upper()}")
> ```
>
> 对于这种复杂的表达式内插，请参照我的一个开源代码(https://github.com/haifenghuang/magpie)

如何实现这种支持呢？我们的词法分析器(Lexer)能够正确的分析出双引号中的内容。我们的语法解析器（Parser）也能够正确的识别这种形式。所以，我们的更改就集中在`解释器`的解释部分。

对于字符串的解释，我们原有的代码如下：

```go
//eval.go
func evalStringLiteral(s *ast.StringLiteral, scope *Scope) Object {
	return NewString(s.Value)
}
```

直接返回一个字符串对象。而现在我们要支持字符串变量内插。我们可以写一个辅助函数，这个辅助函数可以将字符串中的`$var`或者`${var}`中保存的值（实际上是对象系统中的对象，即Object）取出来。在讲解这个辅助函数的实现之前，由于需要用到`go`语言正则表达式(`regexp`)包中的`ReplaceAllStringFunc`函数，所以我们先理解一下这个函数。`ReplaceAllStringFunc`这个方法。它的原型如下：

```go
// 在 src 中搜索匹配项，然后将每一个匹配的内容经过 repl 处理后，替换 src 中的匹配项，
// 并返回替换后的结果
func (re *Regexp) ReplaceAllStringFunc(src string, repl func(string) string) string

//举例：
re := regexp.MustCompile(`[^aeiou]`) //这里正则表达式的的模式匹配的含义是'匹配非元音字母'
fmt.Println(re.ReplaceAllStringFunc("seafood fool", func(m string) string {
	return strings.ToUpper(m)
})) //将所有的非元音字母替换为大写
//结果： SeaFooD FooL
```

>  如果对正则表达式不太懂的读者，请参阅相关的书籍或者查找网络上的相关文章自行学习。

知道了`ReplaceAllStringFunc`这个函数的用法，现在来看一下之前说的辅助函数的实现：

```go
func InterpolateString(str string, scope *Scope) string {
	/*下面是正则表达式各个部分的简单解释：
		(\\\\)?             匹配零个或者一个'\'字符
		\\$                 匹配'$'字符
		(\\{)?              匹配零个或者一个'{'字符
		([a-zA-Z_0-9]{1,})  匹配标识符
		(\\})?              匹配零个或者一个'}'字符
	*/
    
	re := regexp.MustCompile("(\\\\)?\\$(\\{)?([a-zA-Z_0-9]{1,})(\\})?")
	str = re.ReplaceAllStringFunc(str, func(m string) string {
		// 如果字符串以一个'\'开头，则它是一个转移字符，则我们将'\$var'变更成'$var'，并返回
		if string(m[0]) == "\\" {
			return m[1:]
		}

		// 如果字符串以一个'$'开头，则它是个字符串变量内插。
		// 注：这里我们需要支持${var}和$var两种形式。
		name := ""
        if m[1] == '{' { //${var}这种形式
            if m[len(m)-1] != '}' { // 如果最后一个字符不是'}'。例如： "${var"
				return m //直接原封不动返回
			}
            name = m[2 : len(m)-1] //去掉开始的'{'字符和最后的'}'， 即将${var}中的'var'取出
		} else { //$var这种形式
			name = m[1:] //即将$var中的'var'取出
		}

		v, ok := scope.Get(name) //从scope中取得变量的值
		if !ok { //如果没有找到，则返回一个空字符串（注意：这里无需报错）
			return ""
		}

		return v.Inspect()
	})

	return str
}
```

这里面主要就是这个正则表达式的写法比较晦涩难懂。第10行我们构造了一个模式匹配正则表达式对象`re`，这个正则表达式匹配的内容就是`$var`、`\$var`、`${var`或者`${var}`这几种形式。这意味着`ReplaceAllStringFunc`函数的第二个参数（是一个函数）中的`m`就是这几个值。然后第11行，我们调用正则表达式对象的`ReplaceAllStringFunc`方法，将这几种形式转变成相应的字符串：

| 形式(实际是m的值) | 对应代码行号 | 替换成                                                       |
| ----------------- | ------------ | ------------------------------------------------------------ |
| $var              | 26行         | scope中'var'变量对应的值。如果scope没有此'var'变量，则转换为空字符串（代码第31行） |
| \\$var            | 14行         | 不变                                                         |
| ${var             | 22行         | 不变                                                         |
| ${var}            | 24行         | scope中'var'变量对应的值。如果scope没有此'var'变量，则转换为空字符串（代码第31行） |

有了这个`InterpolateString`辅助函数，我们的`evalStringLiteral`函数的实现就简单了：

```go
//eval.go
func evalStringLiteral(s *ast.StringLiteral, scope *Scope) Object {
	return NewString(InterpolateString(s.Value, scope))
}
```

## 测试

```javascript
ax = "hello"
bx = 1024
println("\\$ax = ${ax}, bx = $bx, ${ax")
//结果：$ax = hello, bx = 1024, ${ax
```



下一节，我们会加入`函数装饰器（Decorator）`的支持。




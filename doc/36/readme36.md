# `switch-case`支持

这一节中，我们将给`magpie语言`提供`switch-case`的支持。先来看一下使用例，让读者对这一节要实现的功能有个大概的了解：

```go
fn switchTest(name ) {
    switch name {
        case "welcome" {
            printf("Matched welcome: literal\n");
        }
        case /^Welcome$/ , /^WELCOME$/i {
            printf("Matched welcome: regular-expression\n");
        }
        case "Huang" + "HaiFeng" {
	        printf("Matched HuangHaiFeng\n" );
        }
        case 3, 6 {
            printf("Matched Number %d\n", name);
			fallthrough
        }
        case 9 { //这个case就是为了测试上面的fallthrough
            printf("Matched Number %d\n", name);
        }
        default {
	        printf("Default case: %v\n", name );
        }
    }
}

switchTest( "welcome" );
switchTest( "WelCOME" );
switchTest( "HuangHaiFeng" );
switchTest( 3 );
switchTest( "Bob" );
switchTest( false );
```

这里有几点需要注意：

1. 一个`case`分支中可以有多个表达式，中间用逗号分割。
2. `case`分支的块语句里面不需要`break`。
3. `default`分支不是必须的，且`default`分支不必放在最后，可以放在任何位置。
3. `fallthrough`只能用在`switch-case`语句中，且只能放在`case`分支的最后。



现在让我们看一下需要做哪些更改：

1. 在词元（Token）源码`token.go`中加入新的词元（Token）类型
2. 在抽象语法树（AST）的源码`ast.go`中加入`switch-case语句`对应的抽象语法表示
3. 在语法解析器（Parser）的源码`parser.go`中加入对`switch-case语句`的语法解析
3. 在对象系统（Object）的源码`object.go`中加入`fallthrough`对象
5. 在解释器（Evaluator）的源码`eval.go`，加入对`switch-case语句`的解释



## 词元（Token）的更改

不多做解释，直接看代码：

```go
//token.go
//词元类型
const (
	//...
	TOKEN_SWITCH       //switch
	TOKEN_CASE         //case
	TOKEN_DEFAULT      //default
	TOKEN_FALLTHROUGH  //fallthrough
)

//词元类型的字符串表示
func (tt TokenType) String() string {
	switch tt {
	//...
	case TOKEN_SWITCH:
		return "SWITCH"
	case TOKEN_CASE:
		return "CASE"
	case TOKEN_DEFAULT:
		return "DEFAULT"
	case TOKEN_FALLTHROUGH:
		return "FALLTHROUGH"
	}
}

//关键字
var keywords = map[string]TokenType{
	//...
	"switch":      TOKEN_SWITCH,
	"case":        TOKEN_CASE,
	"default":     TOKEN_DEFAULT,
    "fallthrough": TOKEN_FALLTHROUGH,
}

```



## 抽象语法树（AST）的更改

从本篇开头的例子中我们可以得出`switch-case语句`的一般表示形式如下：

```javascript
switch <expression> {
case <expr1>, <expr2>, ... { block }
case <expr3>, <expr4>, ... { block }
default { block }
}
```

有了上面的说明，我们可以很容易得到其抽象语法表示：

```go
//ast.go
/*
    switch Expr {
    case expr1, expr2, ... { block }
    case expr3, expr4, ... { block }
    ...
    default { block }
	}
*/
type SwitchExpression struct {
	Token       token.Token
	Expr        Expression //需要匹配的表达式
	Cases       []*CaseExpression //case或者default分支
    RBraceToken token.Token //右花括弧的位置，用在End()方法中
}

func (se *SwitchExpression) Pos() token.Position {
	return se.Token.Pos
}

func (se *SwitchExpression) End() token.Position {
	return se.RBraceToken.Pos
}

func (se *SwitchExpression) expressionNode()      {}
func (se *SwitchExpression) TokenLiteral() string { return se.Token.Literal }
func (se *SwitchExpression) String() string {
	var out bytes.Buffer
	out.WriteString("switch ")
	out.WriteString(se.Expr.String())
	out.WriteString("{ ")

	for _, item := range se.Cases {
		if item != nil {
			out.WriteString(item.String())
		}
	}
	out.WriteString(" }")

	return out.String()
}

/*
   case expr1, expr2, ... { block }
   default                { block }
*/
type CaseExpression struct {
	Token       token.Token
	Default     bool //是否是default分支
	Exprs       []Expression //逗号分割的表达式，针对非default分支
	Block       *BlockStatement //块语句
	RBraceToken token.Token //右花括弧的位置，用在End()方法中
}

func (ce *CaseExpression) Pos() token.Position {
	return ce.Token.Pos
}

func (ce *CaseExpression) End() token.Position {
	return ce.RBraceToken.Pos
}

func (ce *CaseExpression) expressionNode()      {}
func (ce *CaseExpression) TokenLiteral() string { return ce.Token.Literal }
func (ce *CaseExpression) String() string {
	var out bytes.Buffer

	if ce.Default {
		out.WriteString("default ")
	} else {
		out.WriteString("case ")

		exprs := []string{}
		for _, expr := range ce.Exprs {
			exprs = append(exprs, expr.String())
		}
		out.WriteString(strings.Join(exprs, ","))
	}
	out.WriteString(ce.Block.String())
	return out.String()
}
```

对于`fallthrough`来说，实际上和`break`及`continue`非常类似，我们只需拷贝过来，更改一下即可：

```go
//fallthrough
type FallthroughExpression struct {
	Token token.Token
}

//t: through
func (t *FallthroughExpression) Pos() token.Position {
	return t.Token.Pos
}

func (t *FallthroughExpression) End() token.Position {
	length := utf8.RuneCountInString(t.Token.Literal)
	pos := t.Token.Pos
	return token.Position{Filename: pos.Filename, Line: pos.Line, Col: pos.Col + length}
}

func (t *FallthroughExpression) expressionNode()      {}
func (t *FallthroughExpression) TokenLiteral() string { return t.Token.Literal }

func (t *FallthroughExpression) String() string { return t.Token.Literal }
```



## 语法解析器（Parser）的更改

我们需要给`switch-case`及`fallthrough`注册前缀表达式回调函数：

```go
//parser.go

func (p *Parser) registerAction() {
	//...
	p.registerPrefix(token.TOKEN_SWITCH, p.parseSwitchExpression)
	p.registerPrefix(token.TOKEN_FALLTHROUGH, p.parseFallThroughExpression)
```

为了确保`fallthrough`只能使用在`switch-case`语句中，我们需要给`Parser`结构加入一个`fallthroughDepth`变量，当解析`switch-case`语句的时候，我们将这个`fallthroughDepth`变量自增，解析完`switch-case`语句返回前，将这个值在自减。这样我们就能够确保只有在`switch-case`语句中，这个`fallthroughDepth`变量才不为0，如果`fallthrough`变量为零，则表示不在`switch-case`语句中，就报告解析错误。

```go
//parser.go
type Parser struct {
	//...
	fallthroughDepth int //current fallthrough depth (0 if not in switch-cases)
}
```

下面是`parseSwitchExpression`函数的代码：

```go
//parser.go
/*
    switch Expr {
    case expr1, expr2, ... { block }
    case expr3, expr4, ... { block }
    ...
    default { block }
	}
*/
func (p *Parser) parseSwitchExpression() ast.Expression {
	p.fallthroughDepth++ //自增
	switchExpr := &ast.SwitchExpression{Token: p.curToken}

	p.nextToken() //skip 'switch'
	switchExpr.Expr = p.parseExpression(LOWEST)
	if switchExpr.Expr == nil {
		return nil
	}

	if !p.expectPeek(token.TOKEN_LBRACE) {
		return nil
	}
	p.nextToken()

	default_cnt := 0 //用来判断default分支的个数。我们只允许default分支有一个
	var defaultToken token.Token

    for !p.curTokenIs(token.TOKEN_RBRACE) { //如果不是右花括弧'}'就继续
		if p.curTokenIs(token.TOKEN_EOF) {  //判断是否到达文件末尾
			msg := fmt.Sprintf("Syntax Error:%v- unterminated switch statement", p.curToken.Pos)
			p.errors = append(p.errors, msg)
			p.errorLines = append(p.errorLines, p.curToken.Pos.Sline())
			return nil
		}

		//不是'case'或者'default'，则报错
		if !p.curTokenIs(token.TOKEN_CASE) && !p.curTokenIs(token.TOKEN_DEFAULT) {
			msg := fmt.Sprintf("Syntax Error:%v- expected 'case' or 'default'. got %s instead", p.curToken.Pos, p.curToken.Type)
			p.errors = append(p.errors, msg)
			p.errorLines = append(p.errorLines, p.curToken.Pos.Sline())
			return nil
		}

		caseExpr := &ast.CaseExpression{Token: p.curToken}
		if p.curTokenIs(token.TOKEN_CASE) {
			p.nextToken() //skip 'case'

			//处理'case expr1, expr2, ...'
			caseExpr.Exprs = append(caseExpr.Exprs, p.parseExpression(LOWEST))
			for p.peekTokenIs(token.TOKEN_COMMA) {
				p.nextToken() //忽略当前词元
				p.nextToken() //忽略','
				caseExpr.Exprs = append(caseExpr.Exprs, p.parseExpression(LOWEST))
			}
		} else if p.curTokenIs(token.TOKEN_DEFAULT) { //default
			default_cnt++
			if default_cnt > 1 {
				defaultToken = p.curToken //记住第二个default的位置，用来报错用
			}
			caseExpr.Default = true
		}

		//如果有超过1个的default，则报错
		if default_cnt > 1 {
			msg := fmt.Sprintf("Syntax Error:%v- more than one default are not allowed", 
                               defaultToken.Pos)
			p.errors = append(p.errors, msg)
			p.errorLines = append(p.errorLines, switchExpr.Token.Pos.Sline())
			return nil
		}

		if !p.expectPeek(token.TOKEN_LBRACE) {
			return nil
		}

        //解析block块
		caseExpr.Block = p.parseBlockStatement()
		if !p.curTokenIs(token.TOKEN_RBRACE) {
			msg := fmt.Sprintf("Syntax Error:%v- expected token to be '}', got %s instead",  p.curToken.Pos, p.curToken.Type)
			p.errors = append(p.errors, msg)
			p.errorLines = append(p.errorLines, p.curToken.Pos.Sline())
			return nil

		}
		caseExpr.RBraceToken = p.curToken

		p.nextToken() //忽略'}'
		switchExpr.Cases = append(switchExpr.Cases, caseExpr)
	}

    //检查'fallthrough'的位置，必须是case分支的最后一个，
	//并且最后一个分支不能有`fallthrough`。
	for i, cse := range switchExpr.Cases {
		lastCase := i == len(switchExpr.Cases)-1 //是否为最后一个分支
		for j, stmt := range cse.Block.Statements {
			lastStmt := j == len(cse.Block.Statements)-1 //是否为最后一个语句
			switch stmt := stmt.(type) {
			case *ast.ExpressionStatement:
				//如果语句不是`fallthrouth`就不管，继续判断一个
				if _, ok := stmt.Expression.(*ast.FallthroughExpression); !ok {
					continue
				}

				//代码运行到这里，则表示是'fallthrough'
				//判断'fallthrough'是否为最后一个语句，不是最后一个语句，则报错
				if !lastStmt {
					msg := fmt.Sprintf("Syntax Error:%v- fallthrough should be the last one", stmt.Pos().Line)
					p.errors = append(p.errors, msg)
					p.errorLines = append(p.errorLines, p.curToken.Pos.Sline())
					return nil
				}

                //如果'fallthrough'落在最后一个case分支里面，则报错
				if lastCase {
					msg := fmt.Sprintf("Syntax Error:%v- cannot fallthrough final case in switch", stmt.Pos().Line)
					p.errors = append(p.errors, msg)
					p.errorLines = append(p.errorLines, stmt.Pos().Sline())
					return nil
				}
			}
		}
	}
    
	p.fallthroughDepth-- //自减

	switchExpr.RBraceToken = p.curToken
	return switchExpr
}
```

我在代码中提供了比较详细的注释，阅读起来应该不难。



## 对象(Object)系统的更改

几乎和`Break`及`Continue`对象一样，`fallthrough`对象如下：

```go
//object.go
const (
	//...
	FALLTHROUGH_OBJ  = "FALLTHROUGH"
	//...
)

var (
	//...
	FALLTHROUGH = &Fallthrough{} //系统中所有的fallthrough都是相同的对象
)


type Fallthrough struct{}

func (f *Fallthrough) Inspect() string  { return "fallthrough" }
func (f *Fallthrough) Type() ObjectType { return FALLTHROUGH_OBJ }
func (f *Fallthrough) CallMethod(line string, scope *Scope, method string, args ...Object) Object {
	return newError(line, ERR_NOMETHOD, method, f.Type())
}
```

代码就是从`Break`及`Continue`中的任何一个拷贝过来，做了简单的修改。



## 解释器（Evaluator）的更改

首先，我们需要在`Eval()`函数的`switch`语句中增加两个`case`分支，一个用来处理`switch`，另一个用来处理`fallthrough`：

```go
//eval.go
func Eval(node ast.Node, scope *Scope) (val Object) {
    //...

	switch node := node.(type) {
	//...
	case *ast.FallthroughExpression:
		//所有的'fallthrough'都一样，没有区别。就像系统中只有一个真(TRUE)，只有一个假(FALSE)一样的道理
        return FALLTHROUGH
	case *ast.SwitchExpression:
		return evalSwitchExpression(node, scope)
	}

	return nil
}
```

第7行的`case`分支用来处理`fallthrough`。第10行的`case`分支用来处理`switch-case`表达式。

跟`Break`和`Continue`一样，如果我们在块语句(Block Statement)中遇到`fallthrough`，那么就需要提前返回：

```go
//eval.go
func evalBlockStatement(block *ast.BlockStatement, scope *Scope) Object {
	var result Object = NIL
	for _, statement := range block.Statements {
		result = Eval(statement, scope)
		if result != nil {
			rt := result.Type()
			if rt == RETURN_VALUE_OBJ || rt == ERROR_OBJ {
				return result
			}
		}
		if _, ok := result.(*Break); ok {
			return result
		}
		if _, ok := result.(*Continue); ok {
			return result
		}
		if _, ok := result.(*Fallthrough); ok {
			return result
		}
	}
	return result
}
```

这里列出了`evalBlockStatement`的所有代码。实际上只有18-20行的`if`分支是新增的逻辑。

`evalSwitchExpression`函数的实现如下：

```go
//eval.go
//解释switch-case表达式
/*
    switch Expr {
    case expr1, expr2, ... { block }
    case expr3, expr4, ... { block }
    ...
    default { block }
	}
*/
unc evalSwitchExpression(switchExpr *ast.SwitchExpression, scope *Scope) Object {
	obj := Eval(switchExpr.Expr, scope)

	var defaultBlock *ast.BlockStatement
	match := false
	through := false

loopCases:
	for _, choice := range switchExpr.Cases { //遍历所有的case分支（包括default分支）
		if choice.Default { //如果是defult分支则继续。default分支放到最后处理
			defaultBlock = choice.Block
			continue
		}

		// 只有在非'fallthrough'模式下，才解释case分支
		if !through {
			for _, expr := range choice.Exprs {
				out := Eval(expr, scope)

				// 是否为字面量匹配(literal match)?
				if obj.Type() == out.Type() && (obj.Inspect() == out.Inspect()) {
					match = true //找到了匹配项
					break
				}

				// 正则表达式匹配(regexp-match)?
				if out.Type() == REGEX_OBJ {
					matched := out.(*RegEx).RegExp.MatchString(obj.Inspect())
					if matched {
						match = true  //找到了匹配项
						break
					}
				}
			}
		}

		//如果找到了匹配的case，或者有'fallthrough'
		if match || through {
			through = false
			result := evalBlockStatement(choice.Block, scope) //解释'case'分支的块语句
			if _, ok := result.(*Fallthrough); ok { //如果是'fallthrough'则继续下一个'case'分支
				through = true
				continue loopCases
			}
			return NIL
		}
	}

	//如果没有匹配的case分支，则处理default分支的块语句
	if !match && defaultBlock != nil {
		return evalBlockStatement(defaultBlock, scope)
	}

	return NIL
}
```



## 测试

这里我们使用本节最开始的例子进行测试。

```perl
# switch.mp
fn switchTest(name ) {
    switch name {
        case "welcome" {
            printf("Matched welcome: literal\n");
        }
        case /^Welcome$/ , /^WELCOME$/i {
            printf("Matched welcome: regular-expression\n");
        }
        case "Huang" + "HaiFeng" {
	        printf("Matched HuangHaiFeng\n" );
        }
        case 3, 6 {
            printf("Matched Number %d\n", name);
			fallthrough
        }
        case 9 {
            printf("Matched Number %d\n", name);
        }
        default {
	        printf("Default case: %v\n", name ); # %v用来打印对象
        }
    }
}

switchTest( "welcome" );
switchTest( "WelCOME" );
switchTest( "HuangHaiFeng" );
switchTest( 3 );
switchTest( "Bob" );
switchTest( false );
```

运行结果：

```
Matched welcome: literal
Matched welcome: regular-expression
Matched HuangHaiFeng
Matched Number 3
Matched Number 3    ----> 这一行的输出，主要是因为上面'case 3,6'分支的最后有fallthrough
Default case: Bob
Default case: false
```



下一节，我们将加入`异常处理(try-catch-finally)`语句的支持。




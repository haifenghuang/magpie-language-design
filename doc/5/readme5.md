# Scope（作用域）

在这一节中，我们会介绍一个全新但是很重要（非常非常非常重要：重要的事情说三遍）的概念。请读者搬好板凳，仔细听课。

在上一节中，我们介绍了`let`语句，其形式举例如下：

```javascript
let x = 5 * 5
```

如果我们仅添加对该`let`语句解释（Evaluating）的支持是不够的。我们必须确保执行了上面的语句后，变量`x`的值为10。而且之后的语句用到这个变量`x`的时候，能够正确的取到这个变量`x`中存储的值。

那么我们怎么样保存这个变量`x`的值呢？这就是`Scope`发挥作用的时候了。简单来说`Scope`就是用来存储变量及其值的地方。更直接点说，`Scope`实际上就类似一个哈希表，哈希表的`key`存储的是变量的名称，哈希表的'value'存储的是变量的值。

> 有的文章称之为【Environment（环境）】，也有的称之为【Context（上下文）】。实际上和我们这里讲的`Scope`大致是同一个概念。

下面我们来看看`Scope`的代码表示：

```go
//scope.go
//Scope结构
type Scope struct {
	store       map[string]Object //存储变量名和变量值的哈希
	parentScope *Scope            //父Scope
    
    //这个Write变量的主要目的：
    //  当我们的解释程序被用在命令行的时候，我们程序的输出是往标准输出（stdout）上输出的。
    //  但是当我们的解释程序是用来和浏览器交互的时候，我们程序的输出就不能往标准输出（stdout）上输出，
    //  而应该往一个buffer中输出，然后将这个buffer的内容传递给wasm程序（与网页交互的程序）。
	Writer      io.Writer
}

//获取参数`name`中存储的对象（Object）值
func (s *Scope) Get(name string) (Object, bool) {
	obj, ok := s.store[name] //在自身的map中寻找
    if !ok && s.parentScope != nil { //没有找到，则在父scope中查找(如果有父scope的话)
		obj, ok = s.parentScope.Get(name)
	}
	return obj, ok
}

//将对象值(val)存储到参数为'name'的哈希中
func (s *Scope) Set(name string, val Object) Object {
	s.store[name] = val
	return val
}

//从Scope中删除指定的name所对应的条目
func (s *Scope) Del(name string) {
	delete(s.store, name)
}

//创建一个新的Scope
//当有函数调用的时候，第一个参数'parent'通常不为nil
func NewScope(parent *Scope, w io.Writer) *Scope {
	s := make(map[string]Object) //创建一个哈希对象
	ret := &Scope{store: s, parentScope: parent}
	if parent == nil {
		ret.Writer = w
	} else {
		ret.Writer = parent.Writer
	}

	return ret
}
```

`Scope`结构中的`parentScope`有必要说明一下。假设我们有下面的代码：

```go
{
    a = 10
    {
        println(a)
    }
}
```

当解释器（Evaluator）解释内层的块语句（3-5行）中的`println(a)`函数的时候，这个变量`a`在内层块中并不存在，解释器就会去外层块（1-6行）中去找。如果找到了，就会使用外层块中存储的变量`a`的值。



有了这一节介绍的知识，下一节我们就可以实现对前一节`let`语句和第二节中介绍的`标识符`的解释（Evaluating）了。小伙伴们是不是很期待呢？我得承认我也很期待。:smile:

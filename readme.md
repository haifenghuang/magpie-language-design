# 语言设计

## 初衷
建立这个仓库的想法/初衷，是为了给下面这些编程爱好者提供一个学习语言的场所：
1. 想学语言设计的初学者
2. 有一定经验的编程经验者

开始有想法写这个的时候，准备写个20篇就差不多了，结果越写越多，不知不觉篇数翻倍了。

说实话，来来回回写了差不多半年时间，花费了大量的业余时间。目的很单纯，就是希望通过
自己的微薄之力，能够给那些希望进步，希望学习语言设计的程序猿提供一些关于语言设计的资料。

## 目的
通过这一系列文章的学习，读者能够了解到语言设计并不是一个令人晦涩难懂的东西。
如果通过本系列的学习，读者能够设计自己的语言，本人会感到非常欣慰。

## 原则
当时写这一系列文章的时候，我就给自己定了一个规则，我不会去讲一些深奥的逻辑，
而是希望用最浅显的语言来阐述一些内容。

## 语言
本系列文章采用`go`语言进行讲解，因此读者在阅读之前，最好能够掌握一些`go`语言的基础。
但是如果你没有`go`语言的基础，看懂讲解的内容应该也不会很困难。

## 参照
市面上有很多语言设计方面的书籍，同时从网上也能够找到很多语言设计方面的资料。
我个人比较喜欢的是：
1. Crafting Interpreters(https://craftinginterpreters.com/)
2. Let's Build A Simple Interpreter(https://ruslanspivak.com/lsbasi-part1/)

但是由于它们都是英文的，所以有些读者可能读起来可能会感到比较吃力。

## 运行
每篇文章对应的代码目录下都有一个`run.sh`文件，只需要运行这个文件即可。

## 内容
先声明一下，本系列教程设计的语言，使用的是基于【树遍历解释器（Tree walking interpreter）】的，且本系列不涉及虚拟机(Virtual Machine)及字节码等相关知识。
文章中的很多内容是基于Thorsten Ball的书籍《Writing an interpreter in go》。

我们将从实现一个简单的计算器开始，一步一步实现一个功能丰富的脚本语言。

每一篇文章的代码都对应一个目录（编号从1开始），讲解文章放在doc目录下。所有的文章（markdown格式）都是使用`typora`编写的。
使用`typora`的好处是，它可以将其生成pdf或者html文档。当然如果读者需要，我可以将生成的pdf/html文档发给读者。

> 里面的有些章节可能比较难懂（例如和go语言交互那一篇），读者完全可以略过不看，直接读取下一篇文章。

对于每一篇所新增加的内容，读者可以使用比较工具，来查看新增的代码逻辑。

下面是目录：

1. Simple Math Expression Evaluation
2. Add Identifier for lexer & parser
3. Add evaluation for 'true, false, nil'
4. Add 'let' statement
5. Add scope
6. Add logic for evaluate 'letStatement' & 'identifier expression'
7. Add error handling for evaluation phase
8. Add 'return statement'
9. Add 'block statement'
10. Add 'String' handling
11. Add 'function literal' & 'function call' handling
12. Add 'builtin function' support
13. Add 'if-else' expression support
14. Add'!' support
15. Add 'array' & 'array index' support
16. Add 'Hash' support
17. Add 'method-call' support
18. Add loop support
19. Add 'assignment' support
20. Add 'tuple' support
21. Add 'named function' support
22. Add wasm support
23. Add 'for loop' support
24. Add 'do & while' loop support
25. Add multiple assignment & return-values
26. Add 'printf' builtin support
27. Add '&&' and '||' support
28. Add 'import' statement
29. Add comment support
30. Add 'regular-expression' support
31. Add 'goObj' support
32. Panic, panic, panic...
33. Add builtin 'file' object
34. Add builtin 'os' object
35. Add struct support
36. Add switch support
37. Add try-catch-finally support
38. Add anonymous functions(lambdas) support
39. Comparison operator Chaining(a < b < c) support
40. Add 'in' operator support
41. Add range operator('..') support
42. Variadic Functions
43. Compound assignment operators(+=, -=, etc)
44. True multiple assignment
45. TCO(tail call optimization) support
46. 'Interpolated String' support
47. Decorator support
48. command execution
49. enhanced hash support
50. pipe operator(|>)
51. ordered hash
52. singe binary executable
53. standard library

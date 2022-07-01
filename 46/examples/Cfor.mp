for (i = 0; i < 5; i++) {  # 类似c语言的for循环， '()'必须要有
    if (i > 4) { break }
    if (i == 2) { continue }
    printf("i is %d\n", i)
}

i = 0
for (; i < 5; i++) {  # 无初期化语句
    if (i > 4) { break }
    if (i == 2) { continue }
    printf("i is %d\n", i)
}

i = 0
for (; i < 5;;) {  # 无初期化和更新语句
    if (i > 4) { break }
    if (i == 2) { i++ continue }
    printf("i is %d\n", i)
    i++ # 更新语句
}

i = 0
for (;;;) {  # 等价于'for { block }'语句
    if (i > 4) { break }
    printf("i is %d\n", i)
    i++ # 更新语句
}
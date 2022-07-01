# 测试结构的可变参数
struct math
{
     let name = "math"

    # 构造函数
    fn init(x, y, args...) {
        self.x = x
        self.y= y
        printf("Hello %s\n", self.name)
        for i in args {
            printf("In init, i=%d\n", i)
        }
    }

    fn Sum(x, y, args...) {
        w = x + y
        for i in args {
            w = w + i
        }
        return w
    }

}

//m1 = math(10, 12, 4,5,6,7)
printf("sum = %d\n", math(10, 12, 4,5,6,7).Sum(2,3,4,5,6,7))

# 注意：
#   1. struct中的所有声明的变量和函数，调用/使用的时候
#       都必须加上'self'。例如:
#         self.field
#         self.method(xxx, xxx)
#   2. 所有的方法，如果希望在外部调用的话，那么方法的首字母必须大写
struct math
{
     let name = "math"
     let dummy = 0;

    # 构造函数
    fn init(x, y) {
        self.x = x
        self.y= y
        printf("Hello %s\n", self.name)
    }

    # unexported
    fn add() {
        return self.x + self.y
    }

    fn Add() {
        return self.add()
    }

    fn Sub() {
        return self.x - self.y
    }

    fn Print(msg) {
        printf("Hello %s, self.dummy = %d\n", msg, self.dummy)
    }

    fn Add_then_sub() {
        add_result = self.Add()
        sub_result = self.Sub()
        printf("In add_then_sub: add_result=%d, sub_result=%d\n", add_result, sub_result)
    }
}

m1 = math(10, 12)
printf("add result=%g\n", m1.Add())
printf("sub result=%g\n", m1.Sub())
m1.Add_then_sub()
m1.Print("hhf")

# 会报告如下错误：cannot refer to unexported name 'add' for object 'm1'
printf("add result=%g\n", m1.add())
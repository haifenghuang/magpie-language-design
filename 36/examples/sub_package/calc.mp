#Identifiers and functions which have their first 
#letter 'Uppercased' will be exported.

#not exported, the first letter is not Uppercase.
fn _add(x, y) {
    return x + y
}
#exported
fn Add(x, y) {
    return _add(x, y)
}

#exported
fn Minus(x, y) {
    return x - y
}

#exported
fn Add_and_Minus(x,y) {
    return x + y, x - y
}

#not exported, the first letter is not Uppercase.
fn multiply(x, y) {
    return x * y
}

# 注意：
#   1. struct中的所有声明的变量和函数，调用/使用的时候
#       都必须加上'self'。例如:
#         self.field
#         self.method(xxx, xxx)
#   2. 所有的方法，如果希望在外部调用的话，那么方法的首字母必须大写
struct Math
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
        printf("In add_then_sub: add_result=%g, sub_result=%g\n", add_result, sub_result)
    }
}

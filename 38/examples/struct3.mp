struct math
{
     let name = "math"
     let dummy = 0;

     let divide = fn(x,y) { return x/y }

    # 构造函数
    fn init(add_func, sub_func) {
        self.add_func = add_func
        self.sub_func = sub_func
        printf("Hello %s\n", self.name)
    }

    # unexported
    fn add(x,y) {
        return self.add_func(x,y)
    }

    fn Add(x,y) {
        return self.add(x,y)
    }

    fn Sub(x,y) {
        return self.sub_func(x,y)
    }

    fn Print(msg) {
        printf("Hello %s, self.dummy = %d\n", msg, self.dummy)
    }

    fn Add_then_sub(x,y) {
        add_result = self.Add(x,y)
        sub_result = self.Sub(x,y)
        printf("In add_then_sub: add_result=%d, sub_result=%d\n", add_result, sub_result)
    }

    fn SetMultiplyFunc(multiply_func) {
        self.multiply_func = multiply_func
    }

    fn Multiply(x,y) {
        return self.multiply_func(x,y)
    }

    fn Divide(x,y) {
        return self.divide(x,y)
    }
}

fn add(x,y) { return x + y }
fn sub(x,y) { return x - y }
fn multiply(x,y) { return x * y }
m1 = math(add, sub)
printf("Add result=%g\n", m1.Add(10, 12))
printf("Sub result=%g\n", m1.Sub(10,12))
m1.Add_then_sub(10,12)
m1.Print("hhf")

m1.SetMultiplyFunc(multiply)
printf("Multiply result=%g\n", m1.Multiply(10, 12))

printf("Divide result=%g\n", m1.Divide(10, 2))

# 会报告如下错误：cannot refer to unexported name 'add' for object 'm1'
#printf("add result=%g\n", m1.add(10,12))
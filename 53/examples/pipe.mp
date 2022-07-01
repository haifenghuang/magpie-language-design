struct demo {
    fn Uppercase(s) {
        return s.upper()
    }
}

fn uppercase(s) {
    return s.upper()
}

//函数调用
upper_str1 = "hello, world" |> uppercase
println(upper_str1)

upper_str2 = "hello, world" |> uppercase()
println(upper_str2)

//方法调用
demo_struct = demo()
upper_str1 = "hello, world" |> demo_struct.Uppercase()
println(upper_str1)

upper_str2 = "hello, world" |> demo_struct.Uppercase
println(upper_str2)


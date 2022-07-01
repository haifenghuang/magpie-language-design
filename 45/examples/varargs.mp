let sum = fn(x, y, args...) {
    w = x + y
    for i in args {
        w = w + i
    }
    return w
}(2,3,4,5,6,7)
println(sum)

let add2 = fn(x, args...) {
    w = x
    for i in args {
        w = w + i
    }
    return w
}

let add1 = fn(x,y, args...) {
    return add2(x+y, args...)
}
println(add1(1, 2, 3, 4, 5))


plus = fn(nums...) {
    sum = 0
    for n in nums {
        sum = sum + n
    }
    return sum
}
println(plus(1, 2, 3))
lst = [4, 5, 6]
println(plus(lst...))
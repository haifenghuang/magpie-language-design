let sum = fn(a, b, x...) {
    for arg in @_ {
        println(arg)
    }
}(2,3, 4, 5)
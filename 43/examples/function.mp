#function statement
fn addFunc(x,y) {
    return x+ y
}
println(addFunc(2,3))

#function literal
sumFunc = fn(x,y) {
    return x+ y
}
println(sumFunc(2,3))

##function literal with direct call
sumResult = fn(x,y) { return x+ y }(2,3)
println(sumResult)

#function with multiple return value
fn math(x, y) {
    return x+y, x-y
}
a, b = math(5,3)
printf("a=%g, b=%g\n", a, b)

fn math2(x, y){
    return x+y, x-y, x * y
}
a, _, c = math2(5,3)
printf("a=%g, c=%g\n", a, c)

# anonymous functions(lambdas)
let add = fn (x, factor) {
  x + factor(x)
}
result = add(5, x => x * 2)
println(result)  # result: 15
fn testTCO(n) {
    if n == 0 {
        println("-> success! we are tail call optimized")
        flushStdout()
    } else {
        printf("n=%d\n", n)
        //return testTCO(n - 1)  // stack overflow
        tailcall testTCO(n - 1)
    }
}
testTCO(2000000)


fn factorial(n, total) {
  if n == 1 { return total }
  //return  factorial(n - 1, n * total) // stack overflow
  tailcall factorial(n - 1, n * total)
}

println(factorial(500000, 1))


fn fib_tail(n, a, b)
{
    if n == 0 { return a }
    if n == 1 { return b }
    //return fib_tail(n - 1, b, a + b) // stack overflow
    tailcall fib_tail(n - 1, b, a + b)
}

println(fib_tail(1000000, 0, 1))

fn TailRecursive(number, product) {
    product = product + number
    if number == 1 {
        return product
    }

    //return TailRecursive(number-1, product)  // stack overflow
    tailcall TailRecursive(number-1, product)
}

answer = TailRecursive(400000, 0)
printf("Recursive: %g\n", answer)

fn testTCO(n) {
    if n == 0 {
        println("-> success! we are tail call optimized")
    } else {
        printf("n=%d\n", n)
        return testTCO(n - 1)
    }
}
testTCO(2000000)

fn factorial(n, total) {
  if n == 1 { return total }
  return factorial(n - 1, n * total)
}

println(factorial(500000, 1)) // 120
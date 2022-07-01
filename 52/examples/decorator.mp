fn timer(otherfn) {
    return fn() {
        println("timer start")
        otherfn($_)
        println("timer end")
    }
}

fn log(otherfn) {
    return fn() {
        println("otherfn start")
        otherfn($_)
        println("otherfn end")
    }
}

@log
@timer
fn sum(x, y) {
    printf("%d + %d = %d\n", x, y, x+y)
}

sum(1,2)
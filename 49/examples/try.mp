try {
    let th = 1 + 2
    println(th)
    #throw "Hello"
    throw 10
    println("After throw")
}
catch e {
    if type(e) == "string" {
        printf("Catched, e=%s\n", e)
    } else if type(e) == "number" {
        printf("Catched, e=%d\n", e)
    }
}
finally {
    println("Finally running")
}

println("After try\n\n")


# throw not handled exception
try {
    let th = 1 + 2
    println(th)
    throw "Hello"
    println("After throw")
}
#catch e {
#    printf("Catched, e=%s\n", e)
#}
finally {
    println("Finally running")
}

println("After try")
# nesting
try { 
    try { 
        throw "inner catch error"
    } finally {
        println("finally")
    } 
} catch ex { 
    println(ex)
    try {
        throw [1,2,3]
    } catch ex {
        printf("ex=%s\n", ex)
    }
}


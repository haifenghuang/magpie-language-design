# nesting
try { 
    try { 
        throw "inner catch error"
    } finally {
        println("finally")
    } 
} catch ex { 
    println(ex)
}
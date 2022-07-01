println(10 < 25 < 30)
println(10 < (10 + 15) <= 25)
println(10 < 5 <= 8)
println(10 <= 25 <= 25)
println("abc" < "def" < "ghi")

fn sum (x,y) { x + y }
println(10 < sum(10, 15) <= 25)

if 10 < 25 < 30 {
    println("10 < 25 < 30")
}

if 10 < 25 <= 25 {
    println("10 < 25 <= 25")
}

if 10 < 5 <= 8 {
    println("10 < 5 <= 8")
}

if 10 <= 25 <= 25 {
    println("10 <= 25 <= 25")
}

if 10 <= 25 != 30 {
    println("10 <= 25 != 30")
}

if "abc" < "def" < "ghi" {
    println("abc < def < ghi")
}

a = 12
if 10 != a < 13 {
    println("10 != a < 13")
}

#error
#if "abc" < "def" < "ghi" < "jkl" {
#    println("abc < def < ghi < jkl")
#}
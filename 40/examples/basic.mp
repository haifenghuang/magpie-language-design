s1 = "hello, 黄"  # strings are UTF-8 encoded
println(s1)

三 = 3       # UTF-8 identifier
println(三)

i = 20000000         # int
println(i)

b = true               # bool
println(b)

a = [1, "2"]           # array
println(a)

if 1 in a {
    println("1 in [1, \"2\"]")
}
if !(2 in a) {
    println("2 not in [1, \"2\"]")
}

h = {"a": 1, "b": 2}   # hash
println(h)

if "b" in h {
    println("\"b\" in h")
}

t = (1,2,3)            # tuple
println(t)

n = nil
println(n)

#printf builtin
printf("2**3=%g, 2.34.floor=%.0f\n", 2.pow(3), 2.34.floor())

/* this is a 
   multiple assignment
*/
a, b, c = 2, false, ["x", "y", "z"]
printf("a=%d,b=%t, c=%v\n", a, b, c)

if "hello" in "hello world" {
    println("\"hello\" in \"hello world\"")
}
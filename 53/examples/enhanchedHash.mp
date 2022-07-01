h = {"a": 1, "b": 2, "x":[12,34,56]}   # hash
h.a = 10
printf("h.a=%d\n", h.a)
h.c= fn(x,y) { return x + y }
println(h.c(2,3))
h["d"] = "www"
println(h.d)
println(h.x[2])


fn demo() {
  h = {}
  h.a = 10
  h.b = 2
  h.c = fn(x,y) { return x + y }
  return h
}

hs = demo()
printf("hs.a=%d\n", hs.a)
println(hs.c(2,3))
println(hs.b)

//array, tuple & string
a = [1, "hello world"]           // array
println(a.1)
a.1 = "hello"
println(a.1)

t = (1, "hello world")           // tuple
println(t.1)


s = "Hello World"                //string
println(s.6)
s.6 = "myz"
println(s.6)
println(s)


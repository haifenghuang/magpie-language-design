#if/else-if
x = 12
result = if x > 10 {2} else if x > 5 {3} else {4}
println(result == 2)

x = 10
result = if x > 10 {2} else if x > 5 {3} else {4}
println(result == 3)

x = 3
result = if x > 10 {2} else if x > 5 {3} else {4}
println(result == 4)

x = 8
result = if x >= 8 {2} else if x > 5 {3} else {4}
println(result == 2)

x = 8
result = if x <= 8 {2} else if x > 5 {3} else {4}
println(result == 2)

x = 8
result = if x == 8 {2} else if x > 5 {3} else {4}
println(result == 2)

x = 8
result = if x != 8 {2} else if x > 5 {3} else {4}
println(result == 3)

#for
arr = [1, true, "Hello"]; 
for item in arr {
    println(item)
} 
println()

for idx, item in arr {
    if idx == 2 { break } 
    println(item) 
}
println()

hash = {"name": "huanghaifeng", "height": 165}
for k, v in hash {
    printf("key=%s, value=%v\n", k, v)
}

str = "Hello"
for c in str {
    println(c)
}
println()

tup = (1, true, "Hello")
for item in tup {
    println(item)
}
println()

for idx, item in tup {
    if idx == 2 { break }
    println(item)
}
println()

#while
x = 3
while x-- > 0 {
    println(x)
}
println()

x = 5
while x-- > 0 {
    println(x)
    if x == 2 { break }
}
println()

x = 5
while x-- > 0 {
    if x == 4 { continue }
    else if x == 2 { break }
    println(x)
}
println()

#do
x = 3
do {
    x--
    println(x)
    if x == 1 { break }
}
println()
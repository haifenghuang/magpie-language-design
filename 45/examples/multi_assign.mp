arr = [10, 12]
arr[0] = 25
printf("arr before: %s\n", arr)
arr[0], arr[1] = arr[1], arr[0]
printf("arr after: %s\n", arr)

hash = {"a": 10, "b": "hello"}
printf("hash before: %s\n", hash)
hash["a"], hash["b"] = hash["b"], hash["a"]
printf("hash after: %s\n", hash)

a, bc, _, d = 12, "hello", "world", [1,2,3]
printf("a=%d, bc=%s, d=%s\n", a, bc, d)

fn math(x, y) {
   return x + y, x -y
}

arr2=[110, 200]
arr2[0], sub_result = math(10, 2)
printf("arr2=%s, sub_result=%d\n", arr2, sub_result)

w1 = 10
w2 = 20
w1, w2 = w2, w1
printf("w1=%d, w2=%d\n", w1, w2)

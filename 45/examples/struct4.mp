struct math
{
	let Name = "MATH"
	let arr = [1,2,3]
	let hs ={ "a": "hello", "b" : (x,y) => { x + y} }
    fn Add(x,y) { return x + y }
	fn Demo() { println(self.arr[2]) }
	fn Demo2() { println(self.hs["b"](4,4)) }

}

m1 = math()
printf("name=%s\n", m1.Name)   # 这里m1.Name中的Name就是一个Identifier
printf("2 + 3 = %d\n", m1.Add(2,3)) # 这里m1.Add()中的Add()就是一个函数调用
m1.Demo() # 这里会打印'3'
m1.Demo2()
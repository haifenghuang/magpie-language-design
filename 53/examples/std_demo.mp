import str
import linq

//str lib
println(IsUpper("h"))
println(IsUpper("H"))
println(StrReverse("Hello"))
if StartsWith("Hello", "Hell") {
	println("Hello starts with 'Hell'")
}

if EndsWith("Hello", "llo") {
	println("Hello ends with 'llo'")
}

printf("StrIndexOf('Hello', 'l') = %d\n", StrIndexOf("Hello", "l"))
printf("StrLastIndexOf('Hello', 'l') = %d\n", StrLastIndexOf("Hello", "l"))
	
if StrContains("Hello", "llo", nil) {
	println("'Hello' contains 'llo'")
}

printf("substr('Hello', 2, 2) = %s\n", SubStr("Hello", 2, 2))
printf("substr('Hello', 2, -1) = %s\n", SubStr("Hello", 2, -1))
printf("ltrim('    Hello    ') = [%s]\n", Ltrim("    Hello    "))
printf("Rtrim('    Hello    ') = [%s]\n", Rtrim("    Hello    "))
printf("trim('    Hello    ') = [%s]\n", Trim("    Hello    "))


//linq lib
result = Linq([1,2,3,4,5,6,7,8,9,10])
	.Where(x => x % 2 == 0)
	.Select(x => x + 1)
	.Reverse()
	.ToRaw()
printf("result = %s\n", result)
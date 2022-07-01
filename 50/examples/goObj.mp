# using go's Printf
fmt.Printf("Hello %s!\n", "go function")

# using go's Sprintf
s = fmt.Sprintf("Hello %s!", "World")
# print with magpie's builtin function `print`
println(s)

fmt.Println(runtime.GOOS)
fmt.Println(runtime.GOARCH)
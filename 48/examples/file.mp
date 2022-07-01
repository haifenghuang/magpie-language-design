file, err = open("./file.log", "w+")
if err {
    println(err)
    os.exit(1)
}

file.writeLine("This is the first line")
file.writeLine("This is the second line")
file.writeString("这是第三行\n")
file.close()


printf("=====Reading file=====\n")
file, err = open("./file.log", "r")
if err {
    println(err)
    os.exit(1)
}

println(file.readLine())
println(file.readLine())
println(file.readLine())
file.close()

# stdin, stdout
print("Please type your name:")
name = stdin.readLine()
stdout.writeLine("Hello " + name + "!")

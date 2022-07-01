import sub_package.calc

println(Add(2,3))

math = Math(2,3)
printf("math.Add() = %g\n", math.Add())

printf("Math.add() = %g\n", Math(2,3).Add())

# error, '_add' not exported
# println(_add(2,3))

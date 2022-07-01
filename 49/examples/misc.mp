fn Person(name, age) {
    self_ = {}
    self_.name = name
    self_.age = age
    self_.str = fn() {
        return self_.name + ", aged " + self_.age
    }
    return self_
}
p = Person("Bob", "42")
print(p.str())
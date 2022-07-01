struct Person {
    fn init(name, score) {
        self.name = name
        self.score = score
    }

    fn GetInfo() {
        return self.name, self.score
    }
}
struct Class {
    fn init(name) {
        self.clsName = name
    }

    fn AddPerson(person) {
        self.persons.push(person)
    }

    fn SetPersons(persons) {
        self.persons = persons
    }

    fn GetClassInfo() {
        printf("%s:\n", self.clsName)
        for person in self.persons {
            name, score = person.GetInfo()
            printf("\tname = %s, score = %.2f\n", name, score)
        } 
    }
}



# person1 = Person("小李", 89.2)
# person2 = Person("小王", 93.5)
# person3 = Person("小张", 70.8)
# 
# cls = Class("五年一班")
# cls.AddPerson(person1)
# cls.AddPerson(person2)
# cls.AddPerson(person3)
# cls.GetClassInfo()

cls = Class("五年一班")
cls.SetPersons([Person("小李", 89.2), Person("小王", 93.5), Person("小张", 70.8)])
cls.GetClassInfo()


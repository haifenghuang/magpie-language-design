fn select(lst, selector) {
  // special cases
  if len(lst) == 0 {
    return []
  } else if selector == nil {
    return lst
  }
  selected = []
  for element in lst {
    if selector(element) {
      selected.push(element)
    }
  }
  return selected
}

fn sum(lst, filter) {
  if len(lst) == 0 {
    return 0
  }

  result = 0
  for element in lst {
    if filter == nil {
      result = result + element
    } else {
       if filter(element) {
         result = result + element
       }
    }
  }
  return result
}

lst = [1, 4, 3, 6, 7, 0, 5, 8, 10, 9, 2]

##### select
println(select(lst, fn(item) { item > 5 }))

##### sum without filter
println(sum(lst,nil))

##### sum with filter
println(sum(lst,fn(item) { item > 5 }))
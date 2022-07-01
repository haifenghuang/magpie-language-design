import linq

result = Linq([1,2,3,4,5,6,7,8,9,10])
	.Where(x => x % 2 == 0)
	.Select(x => x + 1)
	.Reverse()
	.ToRaw()
printf("[1,2,3,4,5,6,7,8,9,10] where(x %% 2 == 0) = %s\n", result)

//test 'Min'
min = Linq([10,2,3,1,5,6,9]).Min( (x, y) => x - y )
printf("min = %d\n", min)

//test 'Max'
max = Linq([10,2,3,1,52,6,9]).Min( (x, y) => y - x )
printf("max = %d\n", max)

//test 'Concat'
concat_result = Linq([1,2,3,4,5,6,7,8,9,10])
	.Concat(Linq([11,12,13]))
	.ToRaw()
printf("concat_result=%s\n", concat_result)

//test 'All'
all_result = Linq([1,2,3,4,5,6,7,8,9,10])
	.All(x => 1 <= x <= 10)
printf("all_result=%t\n", all_result)

//test 'Any'
any_result = Linq([1,2,3,4,5,6,7,8,9,10])
	.Any(x => x > 10)
printf("any_result=%t\n", any_result)

//test 'Take'
take_result = Linq([11,22,33,44,55,66,77,88,99])
	.Take(4)
	.ToRaw()
printf("take_result=%s\n", take_result)

//test 'TakeWhile'
takewhile_result = Linq([11,22,33,44,55,66,77,88,99])
	.TakeWhile(x => 55 <= x <= 88)
	.ToRaw()
printf("takewhile_result=%s\n", takewhile_result)

//test 'Skip'
skip_result = Linq([11,22,33,44,55,66,77,88,99])
	.Skip(4)
	.ToRaw()
printf("skip_result=%s\n", skip_result)

//test 'SkipWhile'
skipwhile_result = Linq([11,22,33,44,55,66,77,88,99])
	.SkipWhile(x => x > 33)
	.ToRaw()
printf("skipwhile_result=%s\n", skipwhile_result)

//test 'Distinct'
distinct_result = Linq([11,22,11,44,55, 22,33, 33])
	.Distinct((x, y) => x == y)
	.ToRaw()
printf("distinct_result=%s\n", distinct_result)

//test 'IndexOf'
indexOf_result = Linq([11,22,11,44,55, 22,33, 33])
	.IndexOf(x => x == 44)
printf("indexOf_result=%d\n", indexOf_result)

//test 'LastIndexOf'
lastindexOf_result = Linq([11,22,11,44,55, 22,33, 33])
	.LastIndexOf(x => x == 22)
printf("lastindexOf_result=%d\n", lastindexOf_result)

//test 'Slice'
slice_result = Linq([11,22,33,44,55,66,77,88,99])
	.Slice(2, 3)
	.ToRaw()
printf("slice_result=%s\n", slice_result)

//test 'Contains'
contains_result = Linq([1,2,3,4,5,6,7,8,9,10])
	.Contains(8)
printf("contains_result=%t\n", contains_result)

//test 'GroupBy'
circle1 = {"color": "red" ,   "x": 0, "y": 0, "r": 3}
circle2 = {"color": "blue",   "x": 2, "y": 6, "r": 3}
circle3 = {"color": "green",  "x": 5, "y": 9, "r": 3}
circle4 = {"color": "red",    "x": 3, "y": 3, "r": 5}
circle5 = {"color": "blue",   "x":10, "y":10, "r": 5}
circle6 = {"color": "yellow", "x": 8, "y": 8, "r": 5}

circles_arr = [circle1, circle2, circle3, circle4, circle5, circle6]
groupby_result = Linq(circles_arr).GroupBy((item, idx) => { return item["color"] })
printf("groupby_result=%s\n", groupby_result)

//test 'Except'
except_result = Linq([1,2,3]).Except(Linq([2,4,5]), (x,y) => x == y).ToRaw()
printf("except_result=%s\n", except_result)

//test 'Union'
union_result = Linq([1,2,3]).Union(Linq([2,3,4,5]), (x,y) => x == y).ToRaw()
printf("union_result=%s\n", union_result)

//test 'Intersect'
intersect_result = Linq([1,2,3]).Intersect(Linq([2,3,4,5]), (x,y) => x == y).ToRaw()
printf("intersect_result=%s\n", intersect_result)

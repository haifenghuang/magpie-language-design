struct Linq {
	fn init(container) {
		self.Container = container
	}

	fn Where(predicateFn) {
		container = self.Container
		length = len(container)
		if length == 0 {
			return self
		}

		result = []
		for (i = 0; i < length; i++) {
			item = container[i]
			if (predicateFn(item)) {
				result.push(item)
			}
		}

		self.Container = result
		return self
	}

	fn Select(selectFn) {
		container = self.Container
		length = len(container)

		if length == 0 {
			return self
		}

		result = []
		for (i = 0; i < length; i++) {
			item = container[i]

			result.push(selectFn(item))
		}

		self.Container = result
		return self
	}

	fn Reverse() {
		container = self.Container
		length = len(container)

		result = []

		for(i = length - 1; i >= 0; i--) {
			item = container[i]
			result.push(item)
		}

		self.Container = result
		return self
	}

	fn Contains(value, equalityFn) {
		container = self.Container
		length = len(container)

		for (i = 0; i < length; i++) {
			item = container[i]
			if (equalityFn(item, value)) { /* equal */
				return true
			}
		}

		return false
	}

	fn Max(compareFn) {
		container = self.Container
		length = len(container)

		if (length == 0) {
			return nil
		}

		max = container[0]
		for (i = 1; i < length; i++) {
			item = container[i]
			if (compareFn(item, max) > 0) {
				max = item
			}
		}

		return max
	}

	fn Min(compareFn) {
		container = self.Container
		length = len(container)

		if (length == 0) {
			return nil
		}

		min = container[0]

		for (i = 1; i < length; i++) {
			item = container[i]
			if (compareFn(item, min) < 0) {
				min = item
			}
		}

		return min
	}

	fn Concat(otherLinq) {
		container = self.Container
		length = len(container)

		result = []

		for (i = 0; i < length; i++) {
			result.push(container[i])
		}

		if otherLinq != nil {
			otherArr = otherLinq.ToArray()
			otherLen = len(otherArr)

			for (i = 0; i < otherLen; i++) {
				result.push(otherArr[i])
			}
		}

		self.Container = result
		return self
	}

	fn Any(predicateFn) {
		container = self.Container
		length = len(container)

		if length == 0 {
			return false
		}
		if predicateFn == nil {
			return true
		}

		for (i = 0; i < length; i++) {
			if predicateFn(container[i]) {
				return true
			}
		}

		return false
	}

	fn All(predicateFn) {
		container = self.Container
		length = len(container)

		if length == 0 {
			return false
		}

		for (i = 0; i < length; i++) {
			if !predicateFn(container[i]) {
				return false
			}
		}
		return true
	}

	fn Take(number) {
		container = self.Container
		length = len(container)

		if length == 0 || number > length {
			return lq
		}

		result = []
		for (i = 0; i < number; i++) {
			result.push(container[i])
		}

		self.Container = result
		return self
	}

	fn TakeWhile(predicateFn) {
		container = self.Container
		length = len(container)

		if length == 0 {
			return lq
		}

		result = []
		for (i = 0; i < length; i++) {
			if predicateFn(container[i]) {
				result.push(container[i])
			}
		}

		self.Container = result
		return self
	}

	fn Skip(number) {
		container = self.Container
		length = len(container)

		if length == 0 {
			return lq
		}

		result = []
		for (i = number; i < length; i++) {
			result.push(container[i])
		}

		self.Container = result
		return self
	}

	fn SkipWhile(predicateFn) {
		container = self.Container
		length = len(container)

		if length == 0 {
			return lq
		}

		isSkipped = false
		result = []
		for (i = 0; i < length; i++) {
			isSkipped = !predicateFn(container[i])

			if (!isSkipped) {
				result.push(container[i])
			}
		}

		self.Container = result
		return self
	}

	fn Distinct(equalityFn) {
		container = self.Container
		length = len(container)

		if length == 0 {
			return lq
		}

		result = []

		for (i = 0; i < length; i++) {
			found = false
			resultLen = len(result)
			for (j = 0; j < resultLen; j++) {
				if equalityFn(result[j], container[i]) {
					found = true
					break
				}
			}

			if found { continue }
			result.push(container[i])
		}

		self.Container = result
		return self
	}

	fn IndexOf(predicateFn) {
		container = self.Container
		length = len(container)

		for (i = 0; i < length; i++) {
			if predicateFn(container[i]) {
				return i
			}
		}
		return -1
	}

	fn LastIndexOf(predicateFn) {
		container = self.Container
		length = len(container)

		for (i = length -1; i >= 0; i--) {
			if predicateFn(container[i]) {
				return i
			}
		}
		return -1
	}

	fn Slice(startIndex, count) {
		container = self.Container
		if len(container) == 0 {
			return self
		}

		return self.Skip(startIndex).Take(count)
	}

	fn Contains(value) {
		container = self.Container
		length = len(container)

		for (i = 0; i < length; i++) {
			if container[i] == value {
				return true
			}
		}

		return false
	}

	fn GroupBy(groupByFn) {
		container = self.Container
		r = {}
		for idx, item in container {
			key = groupByFn(item, idx)
			if key in r {
				existedList = r[key]
				existedList.push(item)
			} else {
				r[key] = [item]
			}
		}
		return r
	}

	fn Except(otherLinq, equalityFn) {
		container = self.Container
		length = len(container)

		other = otherLinq.Container
		other_length = len(other)

		result = []
		found = 0
		for (i = 0; i < length; i++) {
			found = 0
			item = container[i]

			for (j = 0; j < other_length; j++) {
				otherItem = other[j]
				if equalityFn(item, otherItem) { /* equal */
					found = 1
					break
				}
			}

			if !found {
				result.push(item)
			}
		}

		self.Container = result
		return self
	}

	fn Union(otherLinq, equalityFn) {
		container = self.Container
		length = len(container)

		other = otherLinq.Container
		other_length = len(other)

		result = []
		for (i = 0; i < length; i++) {
			item = container[i]
			result.push(item)
		}

		found = 0
		for (i = 0; i < other_length; i++) {
			found = 0
			otherItem = other[i]

			for (j = 0; j < length; j++) {
				item = container[j]
				if equalityFn(item, otherItem) { /* equal */
					found = 1
					break
				}
			}

			if found == 0 {
				result.push(otherItem)
			}
		} /* end for */

		self.Container = result
		return self
	}

	fn Intersect(otherLinq, equalityFn) {
		container = self.Container
		length = len(container)

		other = otherLinq.Container
		other_length = len(other)

		result = []
		for (i = 0; i < length; i++) {
			item = container[i]

			for (j = 0; j < other_length; j++) {
				otherItem = other[j]
				if equalityFn(item, otherItem) { /* equal */
					result.push(item)
					break
				}
			}
		} /* end for */

		self.Container = result
		return self
	}

	fn ToArray() {
		return self.Container
	}
}

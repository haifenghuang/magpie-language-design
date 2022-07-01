import linq

fn IsUpper(c) {
	return "A" <= c <= "Z"
}

fn IsLower(c) {
	return "a" <= c <= "z"
}
fn IsDigit(c) {
	return "0" <= c <= "9"
}

fn StrReverse(s) {
	return Linq(s).Reverse().ToRaw()
}

fn StartsWith(s, prefix) {
	return Linq(s).Take(len(prefix)).ToRaw() == prefix
}

fn EndsWith(s, suffix) {
	return Linq(s).TakeLast(len(suffix)).ToRaw() == suffix
}

fn StrIndexOf(s, substr) {
	return Linq(s).IndexOf(x => x == substr)
}

fn StrLastIndexOf(s, substr) {
	return Linq(s).LastIndexOf(x => x == substr)
}

fn StrContains(s, substr) {
	return Linq(s).Contains(substr, nil)
}

fn SubStr(s, startIdx, count) {
	if count == -1 {
		return Linq(s).Slice(startIdx, len(s) - startIdx).ToRaw()
	}
	return Linq(s).Slice(startIdx, count).ToRaw()
}

fn Ltrim(s) {
	strLen = len(s)
	count = 0
	for (i = 0; i <= strLen - 1;i++) {
		if s[i] != " " {
			break
		}
		count++
	}
	return Linq(s).Slice(count, strLen - count).ToRaw()
}

fn Rtrim(s) {
	strLen = len(s)
	count = 0
	for (i = strLen - 1; i>=0; i--) {
		if s[i] != " " {
			break
		}
		count++
	}
	return Linq(s).Slice(0, strLen - count).ToRaw()
}

fn Trim(s) {
	return s |> Ltrim() |> Rtrim()
}
// Package strhelp provides helpers to work with strings.
package strhelp

// First returns the first non empty string found.
//
func First(list ...string) string {
	for _, str := range list {
		if str != "" {
			return str
		}
	}
	return ""
}

// Separator returns the string with separator if needed.
//
func Separator(prev, sep, after string) string {
	if prev == "" {
		return after
	}
	return prev + sep + after
}

// Parenthesis returns parenthesis added around text if any.
//
func Parenthesis(msg string) string {
	if msg == "" {
		return ""
	}
	return "(" + msg + ")"
}

// Bracket returns brackets added around text if any.
//
func Bracket(msg string) string {
	if msg == "" {
		return ""
	}
	return "[" + msg + "]"
}

// Package ternary operators!
package ternary

// Int is a ternary operator for int.
//
func Int(test bool, a, b int) int {
	if test {
		return a
	}
	return b
}

// Float64 is a ternary operator for int.
//
func Float64(test bool, a, b float64) float64 {
	if test {
		return a
	}
	return b
}

// String is a ternary operator for string.
//
func String(test bool, a, b string) string {
	if test {
		return a
	}
	return b
}

// Min for int.
//
func Min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// Max for int.
//
func Max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

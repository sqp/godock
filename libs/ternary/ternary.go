// Ternary operators!
package ternary

func Int(test bool, a, b int) int {
	if test {
		return a
	}
	return b
}

func String(test bool, a, b string) string {
	if test {
		return a
	}
	return b
}

func Min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func Max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

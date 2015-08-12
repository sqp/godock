// Package cast provides recasts for lists of ints and floats.
package cast

// BoolToInt casts a bool to an int.
//
func BoolToInt(in bool) int {
	if in {
		return 1
	}
	return 0
}

// FloatsToInts casts a list of floats to ints.
//
func FloatsToInts(in []float64) []int {
	out := make([]int, len(in))
	for i, v := range in {
		out[i] = int(v)
	}
	return out
}

// IntsToFloats casts a list of ints to floats.
//
func IntsToFloats(in []int) []float64 {
	out := make([]float64, len(in))
	for i, v := range in {
		out[i] = float64(v)
	}
	return out
}

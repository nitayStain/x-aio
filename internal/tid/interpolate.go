package tid

import "fmt"

// Interpolates between two lists of floats
func Interpolate(fromList, toList []float64, f float64) ([]float64, error) {
	if len(fromList) != len(toList) {
		return nil, fmt.Errorf("mismatched arguments via interpolating %v: %v", fromList, toList)
	}

	out := make([]float64, len(fromList))
	for i := range fromList {
		out[i] = InterpolateNum(fromList[i], toList[i], f)
	}

	return out, nil
}

// Interpolates two numbers
func InterpolateNum(from, to, f float64) float64 {
	return (from * (1.0 - f)) + (to * f)
}

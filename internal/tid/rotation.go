package tid

import "math"

// converts rotation from degrees to a 2D rotation matrix
func ConvertRotationToMatrix(degrees float64) []float64 {
	sin, cos := getSinCos(degrees)
	return []float64{
		cos,
		-sin,
		sin,
		cos,
	}
}

// converts rotation from degrees to a 2D transformation matrix
func ConvertRotationToTransformMatrix(degrees float64) []float64 {
	sin, cos := getSinCos(degrees)
	return []float64{
		cos,
		sin,
		-sin,
		cos,
		0.0,
		0.0,
	}
}

// return's rotation's sin and cos values
func getSinCos(degrees float64) (float64, float64) {
	rads := degrees * math.Pi / 180.0
	cos := math.Cos(rads)
	sin := math.Sin(rads)

	return sin, cos
}

package tid

import "math"

// Cubic 1D bezier curve for animation implementation
type Cubic struct {
	Curves []float64
}

// Returns a new cubic instance
func NewCubic(curves []float64) *Cubic {
	return &Cubic{Curves: curves}
}

// Calculates a value on the curve by the given time
func (c *Cubic) GetValue(time float64) float64 {
	var startGradient, endGradient, start, mid float64
	end := 1.0

	if time <= 0.0 {
		if c.Curves[0] > 0.0 {
			startGradient = c.Curves[1] / c.Curves[0]
		} else if c.Curves[1] == 0.0 && c.Curves[2] > 0.0 {
			startGradient = c.Curves[3] / c.Curves[2]
		}

		return startGradient * time
	}

	if time >= 1.0 {
		if c.Curves[2] < 1.0 {
			endGradient = (c.Curves[3] - 1.0) / (c.Curves[2] - 1.0)
		} else if c.Curves[2] == 1.0 && c.Curves[0] < 1.0 {
			endGradient = (c.Curves[1] - 1.0) / (c.Curves[0] - 1.0)
		}

		return 1.0 + endGradient*(time-1.0)
	}

	startValue := start
	endValue := end

	for startValue < endValue {
		mid = (startValue + endValue) / 2.0
		x := bezier(c.Curves[0], c.Curves[1], mid)
		if math.Abs(time-x) < 0.00001 {
			return bezier(c.Curves[1], c.Curves[3], mid)
		}

		if x < time {
			startValue = mid
		} else {
			endValue = mid
		}
	}

	return bezier(c.Curves[1], c.Curves[3], mid)
}

// Calculates points on the curve
func bezier(a, b, m float64) float64 {
	return 3.0*a*(1.0-m)*(1.0-m)*m + 3.0*b*(1.0-m)*m*m + m*m*m
}

package mcmath

// Numeric is a constraint that permits signed integer and float types.
type Numeric interface {
	~int | ~int32 | ~int64 | ~float32 | ~float64
}

// Abs returns the absolute value of x.
func Abs[T Numeric](x T) T {
	if x < 0 {
		return -x
	}
	return x
}

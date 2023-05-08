package misc

type number interface {
	int8 | uint8 | int | uint | int16 | uint16 | int32 | uint32 | int64 | uint64 | float32 | float64
}

func Max[T number](a, b T) T {
	if a > b {
		return a
	}
	return b
}

func Min[T number](a, b T) T {
	if a > b {
		return b
	}
	return a
}

func In[T comparable](slice []T, a T) bool {
	for i := range slice {
		if slice[i] == a {
			return true
		}
	}
	return false
}

package misc

type Comparable interface {
	~int8 | ~uint8 | ~int | ~uint | ~int16 | ~uint16 | ~int32 | ~uint32 | ~int64 | ~uint64 | ~float32 | ~float64 | ~string
}

func Max[T Comparable](a, b T) T {
	if a > b {
		return a
	}
	return b
}

func Min[T Comparable](a, b T) T {
	if a > b {
		return b
	}
	return a
}

func In[T any](slice []T, equal func(T) bool) bool {
	for i := range slice {
		if equal(slice[i]) {
			return true
		}
	}
	return false
}

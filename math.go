package misc

func Max[T comparable](a, b T) T {
	if a > b {
		return a
	}
	return b
}

func Min[T comparable](a, b T) T {
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

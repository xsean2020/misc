package misc

// 简单实现三元计算
//go:inline
func Ternary[T any](expr bool, a, b T) T {
	if expr {
		return a
	}
	return b
}

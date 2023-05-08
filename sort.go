package misc

func maxDepth(n int) int {
	var depth int
	for i := n; i > 0; i >>= 1 {
		depth++
	}
	return depth * 2
}

func siftDown_func[T any](data []T, lo, hi, first int, lessFn func(a, b T) bool) {
	root := lo
	for {
		child := 2*root + 1
		if child >= hi {
			break
		}

		var x = first + child
		if child+1 < hi && lessFn(data[x], data[x+1]) {
			child++
		}

		if !lessFn(data[first+root], data[first+child]) {
			return
		}
		data[first+root], data[first+child] = data[first+child], data[first+root]
		root = child
	}
}

// sort.go:heapSort
func heapSort_func[T any](data []T, a, b int, lessFn func(a, b T) bool) {
	first := a
	lo := 0
	hi := b - a
	for i := (hi - 1) / 2; i >= 0; i-- {
		siftDown_func(data, i, hi, first, lessFn)
	}
	for i := hi - 1; i >= 0; i-- {
		//data.Swap(first, first+i)
		data[first], data[first+i] = data[first+i], data[first]
		siftDown_func(data, lo, i, first, lessFn)
	}
}

//  sort.go:medianOfThree
func medianOfThree_func[T any](data []T, m1, m0, m2 int, lessFn func(a, b T) bool) {
	if lessFn(data[m1], data[m0]) {
		data[m1], data[m0] = data[m0], data[m1]
	}

	if lessFn(data[m2], data[m1]) {
		data[m2], data[m1] = data[m1], data[m2]

		if lessFn(data[m1], data[m0]) {
			data[m1], data[m0] = data[m0], data[m1]
		}
	}
}

// sort.go:doPivot
func doPivot_func[T any](data []T, lo, hi int, lessFn func(a, b T) bool) (midlo, midhi int) {
	m := int(uint(lo+hi) >> 1)
	if hi-lo > 40 {
		s := (hi - lo) / 8
		medianOfThree_func(data, lo, lo+s, lo+2*s, lessFn)
		medianOfThree_func(data, m, m-s, m+s, lessFn)
		medianOfThree_func(data, hi-1, hi-1-s, hi-1-2*s, lessFn)
	}
	medianOfThree_func(data, lo, m, hi-1, lessFn)
	pivot := lo
	a, c := lo+1, hi-1
	for ; a < c && lessFn(data[a], data[pivot]); a++ {
	}

	b := a
	for {
		for ; b < c && !lessFn(data[pivot], data[b]); b++ {
		}
		for ; b < c && lessFn(data[pivot], data[c-1]); c-- {
		}
		if b >= c {
			break
		}
		o := c - 1
		data[b], data[o] = data[o], data[b]
		b++
		c--
	}

	protect := hi-c < 5
	if !protect && hi-c < (hi-lo)/4 {
		dups := 0
		if !lessFn(data[pivot], data[hi-1]) {
			var o = hi - 1
			data[c], data[o] = data[o], data[c]
			c++
			dups++
		}

		if !lessFn(data[b-1], data[pivot]) {
			b--
			dups++
		}
		if !lessFn(data[m], data[pivot]) {
			var o = b - 1
			data[m], data[o] = data[o], data[m]
			b--
			dups++
		}
		protect = dups > 1
	}
	if protect {
		for {
			for ; a < b && !lessFn(data[b-1], data[pivot]); b-- {
			}
			for ; a < b && lessFn(data[a], data[pivot]); a++ {
			}
			if a >= b {
				break
			}
			var o = b - 1
			data[a], data[o] = data[o], data[a]
			a++
			b--
		}
	}

	var n = b - 1
	data[pivot], data[n] = data[n], data[pivot]
	return b - 1, c
}

//  sort.go:insertionSort
func insertionSort_func[T any](data []T, a, b int, lessFn func(a, b T) bool) {
	for i := a + 1; i < b; i++ {
		for j := i; j > a && lessFn(data[j], data[j-1]); j-- {
			data[j], data[j-1] = data[j-1], data[j]
		}
	}
}

func quickSort_func[T any](data []T, a, b, maxDepth int, lessFn func(a, b T) bool) {
	for b-a > 12 {
		if maxDepth == 0 {
			heapSort_func(data, a, b, lessFn)
			return
		}
		maxDepth--
		mlo, mhi := doPivot_func(data, a, b, lessFn)
		if mlo-a < b-mhi {
			quickSort_func(data, a, mlo, maxDepth, lessFn)
			a = mhi
		} else {
			quickSort_func(data, mhi, b, maxDepth, lessFn)
			b = mlo
		}
	}
	if b-a > 1 {
		for i := a + 6; i < b; i++ {
			if lessFn(data[i], data[i-6]) {
				var i_6 = i - 6
				data[i], data[i_6] = data[i_6], data[i]
			}
		}
		insertionSort_func(data, a, b, lessFn)
	}
}

// QuickSort 通用快速排序
func Sort[T any](arr []T, lessFn func(a, b T) bool) {
	sz := len(arr)
	quickSort_func(arr, 0, sz, maxDepth(sz), lessFn)
}

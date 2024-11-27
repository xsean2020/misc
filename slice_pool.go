package misc

import "sync"

type Slice[T any] []T

// BUG: 扩容就数据可以回收继续使用
func (s *Slice[T]) Add(ts ...T) {
	(*s) = append(*s, ts...)
}

func (s *Slice[T]) Reset() {
	var empty T
	for i := range *s {
		(*s)[i] = empty
	}
	*s = (*s)[:0]
}

func NewSlice[T any](size int) Slice[T] {
	return make([]T, 0, size)
}

// SlicePool 是一个通用的 slice 池，支持不同类型和大小的切片复用
type SlicePool[T any] struct {
	pool sync.Pool
}

// NewSlicePool 创建一个新的 SlicePool，指定切片的初始大小
func NewSlicePool[T any](size int) *SlicePool[T] {
	if size < 0 {
		panic("size lt 0")
	}

	return &SlicePool[T]{
		pool: sync.Pool{
			New: func() interface{} {
				return NewSlice[T](size)
			},
		},
	}
}

type Recycle func()

// Get 从池中获取一个切片，如果池为空，则创建一个新的切片
func (sp *SlicePool[T]) Get() (Slice[T], Recycle) {
	s := sp.pool.Get().(Slice[T])
	return s, func() {
		sp.put(s)
	}
}

// Put 将切片归还到池中，自动清空切片内容
func (sp *SlicePool[T]) put(slice Slice[T]) {
	// 清空切片内容，但保留容量
	slice.Reset()
	sp.pool.Put(slice)
}

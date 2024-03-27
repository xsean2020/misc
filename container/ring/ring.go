package ring

import (
	"sync"
)

type Ring[T any] struct {
	data  []T
	tail  int
	head  int
	count int
	sync.Mutex
}

func NewRing[T any](count int) *Ring[T] {
	return &Ring[T]{
		data: make([]T, count),
		tail: 0,
		head: 0,
	}
}

func (r *Ring[T]) Push(t T) {
	r.Lock()
	defer r.Unlock()
	sz := len(r.data)
	r.data[r.tail] = t
	r.tail = (r.tail + 1) % sz
	if r.count < sz {
		r.count++
	} else {
		r.head = (r.head + 1) % sz
	}
}

func (r *Ring[T]) Tail() *T {
	r.Lock()
	defer r.Unlock()
	if r.count == 0 {
		return nil
	}
	sz := len(r.data)
	return &r.data[(r.tail-1)%sz]
}

func (r *Ring[T]) Range(visit func(T) bool) {
	r.Lock()
	defer r.Unlock()
	sz := len(r.data)
	for i := 0; i < r.count; i++ {
		pos := (r.head + i) % sz
		if visit(r.data[pos]) {
			return
		}
	}
}

func (r *Ring[T]) RevRange(visit func(T) bool) {
	r.Lock()
	defer r.Unlock()
	sz := len(r.data)
	m := r.tail + sz
	for i := 1; i <= r.count; i++ {
		pos := (m - i) % sz
		if visit(r.data[pos]) {
			return
		}
	}
}

func (r *Ring[T]) Len() int {
	r.Lock()
	defer r.Unlock()
	return r.count
}

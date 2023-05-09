package registry

import (
	"sync"
)

type Object[T comparable] interface {
	Primary() T
	comparable
}

type Registry[T comparable, V Object[T]] struct {
	indices map[T]int // id -> v
	records []V
	sync.RWMutex
}

func (r *Registry[T, V]) Init(caps int) {
	r.indices = make(map[T]int, caps)
	r.records = make([]V, 0, caps)
}

func New[T comparable, V Object[T]](_cap int) *Registry[T, V] {
	var r = new(Registry[T, V])
	r.Init(_cap)
	return r
}

// register a user
func (r *Registry[T, V]) Register(v V) {
	r.Lock()
	// 如果存在直接覆盖掉
	if idx, ok := r.indices[v.Primary()]; ok {
		r.records[idx] = v
	} else {
		r.indices[v.Primary()] = len(r.records)
		r.records = append(r.records, v)
	}
	r.Unlock()
}

// 删除了
func (r *Registry[T, V]) remove(idx int, k T) {
	tail := len(r.records) - 1
	r.records[idx], r.records[tail] = r.records[idx], r.records[tail]
	r.indices[r.records[idx].Primary()] = idx
	delete(r.indices, k)
	r.records = r.records[:tail]
}

// unregister a user
func (r *Registry[T, V]) Unregister(v V) {
	r.Lock()
	k := v.Primary()
	if idx, ok := r.indices[k]; ok {
		if r.records[idx] == v {
			r.remove(idx, k)
		}
	}
	r.Unlock()
}

// query a user
func (r *Registry[T, V]) Query(id T) (x V) {
	r.RLock()
	if idx, ok := r.indices[id]; ok {
		x = r.records[idx]
	}
	r.RUnlock()
	return
}

// query all user
func (r *Registry[T, V]) All() []T {
	// 按顺序给出结果
	all := make([]T, 0)
	r.RLock()
	for _, v := range r.records {
		all = append(all, v.Primary())
	}
	r.RUnlock()
	return all
}

// return count of online users
func (r *Registry[T, V]) Count() (count int) {
	r.RLock()
	count = len(r.records)
	r.RUnlock()
	return
}

package sync

import (
	"sync"

	"golang.org/x/exp/maps"
)

type Map[K comparable, V any] struct {
	data map[K]V
	sync.RWMutex
}

func NewMap[K comparable, V any]() *Map[K, V] {
	var mp = Map[K, V]{}
	mp.data = make(map[K]V, 1000)
	return &mp
}

func (m *Map[K, V]) Get(k K) (V, bool) {
	m.RLock()
	ret, ok := m.data[k]
	m.RUnlock()
	return ret, ok
}

func (m *Map[K, V]) Keys() []K {
	m.RLock()
	defer m.RUnlock()
	return maps.Keys(m.data)
}

func (m *Map[K, V]) Set(k K, v V) {
	m.Lock()
	m.data[k] = v
	m.Unlock()
}

func (m *Map[K, V]) Delete(k K) {
	m.Lock()
	delete(m.data, k)
	m.Unlock()
}

func (m *Map[K, V]) DeleteEqual(k K, v V, equal func(v1, v2 V) bool) {
	m.Lock()
	defer m.Unlock()
	t, ok := m.data[k]
	if !ok {
		return
	}
	if !equal(t, v) {
		return
	}
	delete(m.data, k)
}

package rand

import (
	pkg "math/rand"
	"sync"
)

/*
	// microsoft generator has extra division step
	func msg(seed uint32) func() uint32 {
	    g := lcg(214013, 2531011, 1<<31, seed)
	    return func() uint32 {
	        return g() / (1 << 16)
	    }
	}
*/

const (
	// bsd
	_a = uint32(1103515245)
	_c = uint32(12345)
	_m = uint32(1 << 31)
)

// 线程不安全
type LCG uint32

func NewLCG(seed uint32) pkg.Source {
	var s LCG
	s.seed(seed)
	return &s
}

func (lcg *LCG) Int63() int64 {
	*lcg = LCG((_a*uint32(*lcg) + _c) % _m)
	return int64(*lcg)
}

func (lcg *LCG) seed(seed uint32) {
	*lcg = LCG(seed)
}

func (lcg *LCG) Seed(seed int64) {
	lcg.seed(uint32(seed))
}

// 线程安全的Source
func NewLockedLCG(seed uint32) pkg.Source {
	var llcg LockedLCG
	llcg.seed(seed)
	return &llcg
}

type LockedLCG struct {
	LCG
	sync.Mutex `json:"-", bson:"-"`
}

func (ll *LockedLCG) Int63() int64 {
	ll.Lock()
	n := ll.LCG.Int63()
	ll.Unlock()
	return n
}

func (ll *LockedLCG) Seed(seed int64) {
	ll.Lock()
	ll.LCG.Seed(seed)
	ll.Unlock()
}

type Interger interface {
	int8 | uint8 | int | uint | int16 | uint16 | int32 | uint32 | int64 | uint64
}

// rand.Rand
// rand.Source
type Value interface {
	Int63() int64
}

// [a, b]
func Between[T Interger](r Value, a, b T) T {
	if a > b {
		a, b = b, a
	}
	diff := b - a + 1
	return T(r.Int63())%diff + a
}

package rand

import (
	pkg "math/rand"
	"testing"
	"time"
)

// http://rosettacode.org/wiki/Linear_congruential_generator#Go
// 线性随机 如果周期太小会循环出现
func Test_rand(t *testing.T) {
	x := pkg.New(NewLCG(uint32(100)))
	var m = map[int64]int{}
	now := time.Now()
	for i := 0; i < 100; i++ {
		o := x.Int63n(16)
		m[o]++
		t.Log(o)

	}
	t.Log(m, time.Since(now))
}

func Test_Between(t *testing.T) {
	var m = map[int]int{}

	x := pkg.New(NewLCG(uint32(100)))
	for i := 0; i < 10000; i++ {
		x := Between[int](x, 5, 10)
		m[x]++
		if x < 5 || x > 10 {
			t.Fatal("error", x)
		}
	}
}

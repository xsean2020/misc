package rand

import (
	pkg "math/rand"
	"testing"
	"time"
)

// http://rosettacode.org/wiki/Linear_congruential_generator#Go
func Test_rand(t *testing.T) {
	x := pkg.New(NewLCG(uint32(100)))
	var m = map[int64]int{}
	now := time.Now()
	for i := 0; i < 10000000; i++ {
		m[x.Int63n(10)]++
	}
	t.Log(m, time.Since(now))
}

func Test_Between(t *testing.T) {
	x := pkg.New(NewLCG(uint32(100)))
	for i := 0; i < 10000; i++ {
		x := Between[int](x, 5, 10)
		m[x]++
		if x < 5 || x > 10 {
			t.Fatal("error", x)
		}
	}
}

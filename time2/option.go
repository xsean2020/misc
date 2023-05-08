package time2

import (
	"github.com/panjf2000/ants/v2"
)

type config struct {
	*ants.Pool
}

// 覆盖ants.Pool 的方法
func (c *config) Submit(f func()) {
	if c.Pool != nil {
		if err := c.Pool.Submit(f); err == nil {
			return
		}
	}
	go f()
}

type Option func(c *config)

// Set goroutinepool
func WithGoroutinePool(p *ants.Pool) Option {
	return func(c *config) {
		c.Pool = p
	}
}

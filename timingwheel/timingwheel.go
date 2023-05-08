package timingwheel

import (
	"sync"
	"time"
)

type TimingWheel struct {
	sync.Mutex
	interval   time.Duration
	ticker     *time.Ticker
	maxTimeout time.Duration
	die        chan struct{}
	chs        []chan struct{}
	refs       []int // 记录引用次数
	pos        int
}

func New(interval time.Duration, buckets int) *TimingWheel {
	w := new(TimingWheel)
	w.interval = interval
	w.die = make(chan struct{})
	w.pos = 0
	w.maxTimeout = time.Duration(interval * (time.Duration(buckets)))
	w.chs = make([]chan struct{}, buckets)
	w.refs = make([]int, buckets)
	for i := range w.chs {
		w.chs[i] = make(chan struct{})
	}
	w.ticker = time.NewTicker(interval)
	go w.run()
	return w
}

func (w *TimingWheel) Stop() {
	close(w.die)
}

func (w *TimingWheel) After(timeout time.Duration) <-chan struct{} {
	if timeout >= w.maxTimeout {
		panic("timeout too much, over maxtimeout")
	}

	index := int(timeout / w.interval)
	if 0 < index { // 卫生么要减少
		index--
	}

	w.Lock()
	index = (w.pos + index) % len(w.chs)
	b := w.chs[index]
	w.refs[index]++
	w.Unlock()
	return b
}

func (w *TimingWheel) run() {
	for {
		select {
		case <-w.ticker.C:
			w.onTicker()
		case <-w.die:
			w.ticker.Stop()
			return
		}
	}
}

func (w *TimingWheel) onTicker() {
	w.Lock()
	defer w.Unlock()
	pos := w.pos
	w.pos = (w.pos + 1) % len(w.chs)
	if w.refs[pos] == 0 {
		return
	}
	w.refs[pos] = 0
	lastC := w.chs[pos]
	w.chs[pos] = make(chan struct{})
	close(lastC)
}

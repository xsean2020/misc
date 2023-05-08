package time2

import (
	"sync"
	"time"
)

// wheel timer
const ( // 为啥不固定长度
	tvn_bits uint64 = 6
	tvr_bits uint64 = 8
	tvn_size uint64 = 64  //1 << tvn_bits
	tvr_size uint64 = 256 //1 << tvr_bits

	tvn_mask uint64 = 63  //tvn_size - 1
	tvr_mask uint64 = 255 //tvr_size -1

	defaultTimerSize = 128
)

//  6 * 4 + 8 = 32

type timer struct {
	expires uint64 // 过期时间
	period  uint64 // 周期时间
	f       func(time.Time, interface{})
	arg     interface{} //
	w       *Wheel
	vec     []*timer
	index   int

	sync.RWMutex
}

func (t *timer) when() time.Time {
	t.RLock()
	defer t.RUnlock()
	return t.w.tm.Add(t.w.tick * time.Duration(t.expires))
}

type Wheel struct {
	sync.Mutex
	jiffies uint64
	tm      time.Time
	tvecs   [5][][]*timer
	tick    time.Duration
	quit    chan struct{}
	cfg     config
}

//tick is the time for a jiffies
func NewWheel(tick time.Duration, opts ...Option) *Wheel {
	w := new(Wheel)
	w.quit = make(chan struct{})
	w.tm = time.Now()
	for _, opt := range opts {
		opt(&w.cfg)
	}

	fn := func(size int) [][]*timer {
		tv := make([][]*timer, size)
		for i := range tv {
			tv[i] = make([]*timer, 0, defaultTimerSize)
		}
		return tv
	}

	// init
	var vecsInitSize = [5]uint64{tvr_size, tvn_size, tvn_size, tvn_size, tvn_size}
	for i := range w.tvecs {
		w.tvecs[i] = fn(int(vecsInitSize[i]))
	}

	w.jiffies = 0
	w.tick = tick

	go w.run()
	return w
}

func (w *Wheel) unsafeAdd(t *timer) {
	expires := t.expires

	idx := t.expires - w.jiffies // 判断在那一个区间内 0 - 512

	var tv [][]*timer
	var i uint64

	if idx < 0 {
		i = w.jiffies & tvr_mask
		tv = w.tvecs[0]
	} else if idx < tvr_size {
		i = expires & tvr_mask
		tv = w.tvecs[0]
	} else if idx < (1 << (tvr_bits + tvn_bits)) {
		i = (expires >> tvr_bits) & tvn_mask
		tv = w.tvecs[1]
	} else if idx < (1 << (tvr_bits + 2*tvn_bits)) {
		i = (expires >> (tvr_bits + tvn_bits)) & tvn_mask
		tv = w.tvecs[2]
	} else if idx < (1 << (tvr_bits + 3*tvn_bits)) {
		i = (expires >> (tvr_bits + 2*tvn_bits)) & tvn_mask
		tv = w.tvecs[3]
	} else {
		if idx > 0x00000000ffffffff { // 溢出
			idx = 0x00000000ffffffff
			expires = idx + w.jiffies // 重新调节expire 来确定具体的位置
		}
		// 为啥不直接丢到最后一个桶里面
		i = (expires >> (tvr_bits + 3*tvn_bits)) & tvn_mask
		tv = w.tvecs[4]
	}

	tv[i] = append(tv[i], t)
	t.vec = tv[i]
	t.index = len(tv[i]) - 1
}

// 滚筒方式
func (w *Wheel) cascade(tv [][]*timer, index int) int {
	vec := tv[index]
	tv[index] = vec[:0:defaultTimerSize]
	for _, t := range vec {
		if t == nil {
			continue
		}
		w.unsafeAdd(t)
	}
	return index
}

func (w *Wheel) getIndex(n int) int {
	return int((w.jiffies >> (tvr_bits + uint64(n)*tvn_bits)) & tvn_mask)
}

func (w *Wheel) onTick() {
	w.Lock()

	index := int(w.jiffies & tvr_mask)
	// 第一级已经触发完毕了 后面的桶向前移动
	_ = index == 0 &&
		(w.cascade(w.tvecs[1], w.getIndex(0))) == 0 &&
		(w.cascade(w.tvecs[2], w.getIndex(1))) == 0 &&
		(w.cascade(w.tvecs[3], w.getIndex(2))) == 0 &&
		(w.cascade(w.tvecs[4], w.getIndex(3)) == 0)

	w.jiffies++

	vec := w.tvecs[0][index]
	w.tvecs[0][index] = vec[0:0:defaultTimerSize]
	w.Unlock()

	f := func(vec []*timer) {
		now := time.Now()
		for _, t := range vec {
			if t == nil {
				continue
			}
			t.f(now, t.arg)
			if t.period > 0 { // 周期型性的 ticker
				t.Lock() // 针对 tickerwhen 安全访问
				t.expires = t.period + w.jiffies
				t.Unlock()
				w.addTimer(t)
			}
		}
	}

	// 批量执行
	if len(vec) > 0 {
		w.cfg.Submit(func() {
			f(vec)
		})
	}
}

func (w *Wheel) addTimer(t *timer) {
	w.Lock()
	w.unsafeAdd(t)
	w.Unlock()
}

func (w *Wheel) delTimer(t *timer) {
	w.Lock()
	vec := t.vec
	index := t.index
	if len(vec) > index && vec[index] == t {
		vec[index] = nil
	}
	w.Unlock()
}

func (w *Wheel) resetTimer(t *timer, when time.Duration, period time.Duration) {
	w.delTimer(t)
	t.expires = w.jiffies + uint64(when/w.tick) // 必然有误差
	t.period = uint64(period / w.tick)
	w.addTimer(t)
}

func (w *Wheel) newTimer(when time.Duration, period time.Duration,
	f func(time.Time, interface{}), arg interface{}) *timer {
	t := new(timer)

	t.expires = w.jiffies + uint64(when/w.tick)
	t.period = uint64(period / w.tick)

	t.f = f
	t.arg = arg
	t.w = w

	return t
}

func (w *Wheel) run() {
	ticker := time.NewTicker(w.tick)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			w.onTick()
		case <-w.quit:
			return
		}
	}
}

func (w *Wheel) Stop() {
	close(w.quit)
}

func (w *Wheel) After(d time.Duration) <-chan time.Time {
	return w.NewTimer(d).C
}

func (w *Wheel) Sleep(d time.Duration) {
	<-w.NewTimer(d).C
}

func (w *Wheel) Tick(d time.Duration) <-chan time.Time {
	return w.NewTicker(d, d).C
}

func sendTime(t time.Time, arg interface{}) {
	select {
	case arg.(chan time.Time) <- t:
	default:
	}
}

func (w *Wheel) TickFunc(d time.Duration, period time.Duration, f func()) *Ticker {
	t := &Ticker{
		r: w.newTimer(d, period, func(_ time.Time, _ interface{}) {
			w.cfg.Submit(f)
		}, nil),
	}

	w.addTimer(t.r)
	return t
}

func (w *Wheel) AfterFunc(d time.Duration, f func()) *Timer {
	t := &Timer{
		r: w.newTimer(d, 0, func(_ time.Time, _ interface{}) {
			w.cfg.Submit(f)
		}, nil),
	}
	w.addTimer(t.r)
	return t
}

func (w *Wheel) NewTimer(d time.Duration) *Timer {
	c := make(chan time.Time, 1)
	t := &Timer{
		C: c,
		r: w.newTimer(d, 0, sendTime, c),
	}
	w.addTimer(t.r)
	return t
}

func (w *Wheel) NewTicker(d time.Duration, p time.Duration) *Ticker {
	c := make(chan time.Time, 1)
	t := &Ticker{
		C: c,
		r: w.newTimer(d, p, sendTime, c),
	}
	w.addTimer(t.r)
	return t
}

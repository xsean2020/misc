package time2

import "time"

var defaultWheel *Wheel

func init() {
	defaultWheel = NewWheel(100 * time.Millisecond)
}

func After(d time.Duration) <-chan time.Time {
	return defaultWheel.After(d)
}

func Sleep(d time.Duration) {
	defaultWheel.Sleep(d)
}

package time2

import (
	"sync"
	"testing"
	"time"
)

func TestTimer(t *testing.T) {
	w := NewWheel(time.Millisecond)
	timer := w.NewTimer(time.Second)
	<-timer.C
	if timer.C == nil {
		t.Errorf("timer was not triggered")
	}
}

func TestTicker(t *testing.T) {
	w := NewWheel(time.Millisecond)
	ticker := w.NewTicker(time.Millisecond*500, time.Millisecond*500)
	<-ticker.C
	if ticker.C == nil {
		t.Errorf("ticker was not triggered")
	}
	w.Stop()
}

func TestNewTicker(t *testing.T) {
	w := NewWheel(100 * time.Millisecond)
	ticker := w.NewTicker(time.Millisecond*500, 10*time.Millisecond)
	<-ticker.C
	if ticker.C == nil {
		t.Errorf("ticker was not triggered")
	}
	w.Stop()
}

func TestAfterFunc(t *testing.T) {
	w := NewWheel(time.Millisecond)
	var count int
	var wg sync.WaitGroup
	wg.Add(1)
	w.AfterFunc(time.Second, func() {
		defer wg.Done()
		count++
	})

	wg.Wait()
	if count != 1 {
		t.Errorf("after func was not triggered")
	}
	w.Stop()
}

func TestTickFunc(t *testing.T) {
	w := NewWheel(time.Millisecond)
	var count int
	var wg sync.WaitGroup
	wg.Add(1)
	w.TickFunc(time.Second, time.Second, func() {
		defer wg.Done()
		count++
	})
	wg.Wait()

	if count != 1 {
		t.Errorf("tick func was not triggered")
	}
	w.Stop()
}

package timingwheel

import (
	"testing"
	"time"
)

func TestTimingWheel(t *testing.T) {
	const interval = 100 * time.Millisecond
	const buckets = 100

	tw := New(interval, buckets)

	// Test After()
	t1 := time.Now()
	c1 := tw.After(250 * time.Millisecond)
	<-c1
	t2 := time.Now()
	if dur := t2.Sub(t1); dur <= 250*time.Millisecond-interval || dur >= 250*time.Millisecond+interval {
		t.Errorf("After() took %v, expected %v ~ %v", dur, 250*time.Millisecond-interval, 250*time.Millisecond+interval)
	}

	// Test multiple After() calls
	t1 = time.Now()
	c2 := tw.After(150 * time.Millisecond)
	c3 := tw.After(50 * time.Millisecond)
	<-c3
	t2 = time.Now()
	if dur := t2.Sub(t1); dur < 50*time.Millisecond-interval || dur >= 50*time.Millisecond+interval {
		t.Errorf("After() took %v, expected %v ~ %v", dur, 50*time.Millisecond-interval, dur >= 50*time.Millisecond+interval)
	}
	<-c2
	t3 := time.Now()
	if dur := t3.Sub(t1); dur < 150*time.Millisecond-interval || dur >= 150*time.Millisecond+interval {
		t.Errorf("After() took %v, expected %v ~ %v", dur, 150*time.Millisecond-interval, 150*time.Millisecond+interval)
	}

	// Test Stop()
	tw.Stop()
	select {
	case <-tw.After(50 * time.Millisecond):
		t.Error("After() succeeded after Stop()")
	default:
	}
}

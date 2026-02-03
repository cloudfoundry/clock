package fakeclock_test

import (
	"sync/atomic"
	"testing"
	"time"

	"code.cloudfoundry.org/clock/fakeclock"
)

func TestFakeTickerTicks(t *testing.T) {
	t.Parallel()

	const delta = 10 * time.Millisecond

	fc := fakeclock.NewFakeClock(initialTime)

	ticker := fc.NewTicker(10 * time.Second)
	timeChan := ticker.C()
	requireNoReceive(t, timeChan, delta)

	fc.Increment(5 * time.Second)
	requireNoReceive(t, timeChan, delta)

	fc.Increment(4 * time.Second)
	requireNoReceive(t, timeChan, delta)

	fc.Increment(1 * time.Second)
	requireReceiveEqual(t, timeChan, initialTime.Add(10*time.Second), 1*time.Second)

	fc.Increment(10 * time.Second)
	requireReceiveEqual(t, timeChan, initialTime.Add(20*time.Second), 1*time.Second)

	fc.Increment(10 * time.Second)
	requireReceiveEqual(t, timeChan, initialTime.Add(30*time.Second), 1*time.Second)
}

func TestFakeTickerMultipleTickers(t *testing.T) {
	t.Parallel()

	const (
		period = 1 * time.Second
		delta  = 10 * time.Millisecond
	)

	fc := fakeclock.NewFakeClock(initialTime)

	ticker1 := fc.NewTicker(period)
	ticker2 := fc.NewTicker(period)

	// Receiving directly from ticker.C() makes it easy to miss events; use counters instead.
	var count1 uint32
	var count2 uint32

	stop := make(chan struct{})
	defer close(stop)

	go func() {
		for {
			select {
			case <-stop:
				return
			case <-ticker1.C():
				atomic.AddUint32(&count1, 1)
			case <-ticker2.C():
				atomic.AddUint32(&count2, 1)
			}
		}
	}()

	fc.Increment(period)

	eventually(t, 1*time.Second, func() bool { return atomic.LoadUint32(&count1) == 1 })
	eventually(t, 1*time.Second, func() bool { return atomic.LoadUint32(&count2) == 1 })

	// And ensure no extra ticks arrive shortly thereafter.
	deadline := time.Now().Add(5 * delta)
	for time.Now().Before(deadline) {
		if atomic.LoadUint32(&count1) != 1 {
			t.Fatalf("ticker1 count=%d; want 1", atomic.LoadUint32(&count1))
		}
		if atomic.LoadUint32(&count2) != 1 {
			t.Fatalf("ticker2 count=%d; want 1", atomic.LoadUint32(&count2))
		}
		time.Sleep(1 * time.Millisecond)
	}
}

func TestFakeTickerDoesNotFireEarly(t *testing.T) {
	t.Parallel()

	const (
		period = 1 * time.Second
		delta  = 10 * time.Millisecond
	)

	fc := fakeclock.NewFakeClock(initialTime)

	ticker := fc.NewTicker(period)
	requireNoReceive(t, ticker.C(), delta)

	fc.Increment(period)
	requireReceiveEqual(t, ticker.C(), initialTime.Add(period), 1*time.Second)

	fc.Increment(0)
	requireNoReceive(t, ticker.C(), delta)
}

func TestFakeTickerPanicsOnInvalidDuration(t *testing.T) {
	t.Parallel()

	fc := fakeclock.NewFakeClock(initialTime)

	defer func() {
		if r := recover(); r == nil {
			t.Fatalf("expected panic")
		}
	}()

	_ = fc.NewTicker(0)
}

package fakeclock_test

import (
	"context"
	"sync"
	"testing"
	"time"

	"code.cloudfoundry.org/clock/fakeclock"
)

func TestFakeTimerFires(t *testing.T) {
	t.Parallel()

	const delta = 10 * time.Millisecond

	fc := fakeclock.NewFakeClock(initialTime)

	timer := fc.NewTimer(10 * time.Second)
	timeChan := timer.C()
	requireNoReceive(t, timeChan, delta)

	fc.Increment(5 * time.Second)
	requireNoReceive(t, timeChan, delta)

	fc.Increment(4 * time.Second)
	requireNoReceive(t, timeChan, delta)

	fc.Increment(1 * time.Second)
	requireReceiveEqual(t, timeChan, initialTime.Add(10*time.Second), 1*time.Second)

	fc.Increment(10 * time.Second)
	requireNoReceive(t, timeChan, delta)
}

func TestFakeTimerStopIsIdempotent(t *testing.T) {
	t.Parallel()

	const delta = 10 * time.Millisecond

	fc := fakeclock.NewFakeClock(initialTime)

	timer := fc.NewTimer(time.Second)
	timer.Stop()
	timer.Stop()

	fc.Increment(time.Second)
	requireNoReceive(t, timer.C(), delta)
}

func TestWaitForWatcherAndIncrementTimersAddedAsync(t *testing.T) {
	t.Parallel()

	const duration = 10 * time.Second

	fc := fakeclock.NewFakeClock(initialTime)
	received := make(chan time.Time, 100)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var wg sync.WaitGroup
	ready := make(chan struct{})

	wg.Add(1)
	go func() {
		defer wg.Done()
		close(ready)

		for {
			timer := fc.NewTimer(duration)

			select {
			case ticked := <-timer.C():
				received <- ticked
			case <-ctx.Done():
				return
			}
		}
	}()

	<-ready

	for i := 0; i < 100; i++ {
		fc.WaitForWatcherAndIncrement(duration)

		ticked := <-received
		if got, want := ticked.Sub(initialTime), duration*time.Duration(i+1); got != want {
			t.Fatalf("ticked offset=%s; want %s", got, want)
		}
	}

	cancel()
	waitGroupWithin(t, &wg, 5*time.Second)
}

func TestWaitForWatcherAndIncrementTimerResetAsync(t *testing.T) {
	t.Parallel()

	const duration = 10 * time.Second

	fc := fakeclock.NewFakeClock(initialTime)

	received := make(chan time.Time, 100)
	timer := fc.NewTimer(duration)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var wg sync.WaitGroup
	ready := make(chan struct{})

	// Goroutine that receives timer ticks and resets the timer asynchronously.
	wg.Add(1)
	go func() {
		defer wg.Done()
		close(ready)

		for {
			select {
			case ticked := <-timer.C():
				received <- ticked
				timer.Reset(duration)
			case <-ctx.Done():
				return
			}
		}
	}()

	<-ready

	incrementClock := make(chan struct{})

	// Goroutine that increments the clock when asked.
	wg.Add(1)
	go func() {
		defer wg.Done()
		for {
			select {
			case _, ok := <-incrementClock:
				if !ok {
					return
				}
				fc.WaitForWatcherAndIncrement(duration)
			case <-ctx.Done():
				return
			}
		}
	}()

	for i := 0; i < 100; i++ {
		requireSendWithin(t, incrementClock, struct{}{}, 1*time.Second)

		var timestamp time.Time
		requireReceiveInto(t, received, &timestamp, 5*time.Second)

		// We don't assert on the timestamp value here; this test is checking that
		// timers that reset asynchronously continue to fire without deadlocking.
		_ = timestamp
	}

	close(incrementClock)
	cancel()
	waitGroupWithin(t, &wg, 5*time.Second)
}

func waitGroupWithin(t *testing.T, wg *sync.WaitGroup, d time.Duration) {
	t.Helper()

	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		return
	case <-time.After(d):
		t.Fatalf("timed out waiting for goroutines to exit")
	}
}

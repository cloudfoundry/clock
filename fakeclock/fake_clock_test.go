package fakeclock_test

import (
	"testing"
	"time"

	"code.cloudfoundry.org/clock/fakeclock"
)

var (
	initialTime = time.Date(2014, 1, 1, 3, 0, 30, 0, time.UTC)
)

func TestFakeClockNow(t *testing.T) {
	t.Parallel()

	fc := fakeclock.NewFakeClock(initialTime)

	go fc.Increment(time.Minute)

	eventually(t, 1*time.Second, func() bool {
		return fc.Now().Equal(initialTime.Add(time.Minute))
	})
}

func TestFakeClockSleep(t *testing.T) {
	t.Parallel()

	const delta = 10 * time.Millisecond

	fc := fakeclock.NewFakeClock(initialTime)

	doneSleeping := make(chan struct{})
	go func() {
		fc.Sleep(10 * time.Second)
		close(doneSleeping)
	}()

	requireNotClosed(t, doneSleeping, delta)

	fc.Increment(5 * time.Second)
	requireNotClosed(t, doneSleeping, delta)

	fc.Increment(4 * time.Second)
	requireNotClosed(t, doneSleeping, delta)

	fc.Increment(1 * time.Second)
	requireClosed(t, doneSleeping, 1*time.Second)
}

func TestFakeClockAfter(t *testing.T) {
	t.Parallel()

	const delta = 10 * time.Millisecond

	fc := fakeclock.NewFakeClock(initialTime)

	timeChan := fc.After(10 * time.Second)
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

func TestFakeClockWatcherCount(t *testing.T) {
	t.Parallel()

	t.Run("increments when timers are created", func(t *testing.T) {
		fc := fakeclock.NewFakeClock(initialTime)
		fc.NewTimer(time.Second)
		if got, want := fc.WatcherCount(), 1; got != want {
			t.Fatalf("WatcherCount=%d; want %d", got, want)
		}

		fc.NewTimer(2 * time.Second)
		if got, want := fc.WatcherCount(), 2; got != want {
			t.Fatalf("WatcherCount=%d; want %d", got, want)
		}
	})

	t.Run("decrements when a timer fires", func(t *testing.T) {
		fc := fakeclock.NewFakeClock(initialTime)

		fc.NewTimer(time.Second)
		if got, want := fc.WatcherCount(), 1; got != want {
			t.Fatalf("WatcherCount=%d; want %d", got, want)
		}

		fc.Increment(time.Second)
		if got, want := fc.WatcherCount(), 0; got != want {
			t.Fatalf("WatcherCount=%d; want %d", got, want)
		}
	})
}

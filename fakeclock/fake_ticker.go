package fakeclock

import (
	"sync"
	"time"

	"github.com/pivotal-golang/clock"
)

type fakeTicker struct {
	clock clock.Clock

	mutex    sync.Mutex
	duration time.Duration
	channel  chan time.Time

	timer clock.Timer
}

func NewFakeTicker(clock *FakeClock, d time.Duration) clock.Ticker {
	// buffer is so that .Increment does not block forever
	channel := make(chan time.Time, 1)

	timer := clock.NewTimer(d)

	// note: this happens *synchronously* with .Increment, guaranteeing that each
	// time advance will fire the timer
	timer.(*fakeTimer).onTick(func(time time.Time) {
		timer.Reset(d)
		channel <- time
	})

	return &fakeTicker{
		clock:    clock,
		duration: d,
		channel:  channel,
		timer:    timer,
	}
}

func (ft *fakeTicker) C() <-chan time.Time {
	ft.mutex.Lock()
	defer ft.mutex.Unlock()
	return ft.channel
}

func (ft *fakeTicker) Stop() {
	ft.mutex.Lock()
	ft.timer.Stop()
	ft.mutex.Unlock()
}

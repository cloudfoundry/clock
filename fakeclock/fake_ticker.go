package fakeclock

import (
	"time"

	"github.com/pivotal-golang/clock"
)

type fakeTicker struct {
	timer clock.Timer
}

func NewFakeTicker(timer *fakeTimer) clock.Ticker {
	return &fakeTicker{
		timer: timer,
	}
}

func (ft *fakeTicker) C() <-chan time.Time {
	return ft.timer.C()
}

func (ft *fakeTicker) Stop() {
	ft.timer.Stop()
}

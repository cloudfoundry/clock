package fakeclock_test

import (
	"testing"
	"time"
)

func requireNoReceive[T any](t *testing.T, ch <-chan T, within time.Duration) {
	t.Helper()
	select {
	case v, ok := <-ch:
		if ok {
			t.Fatalf("unexpected receive: %#v", v)
		}
		t.Fatalf("unexpected receive: channel closed")
	case <-time.After(within):
		return
	}
}

func requireReceiveEqual[T comparable](t *testing.T, ch <-chan T, want T, within time.Duration) {
	t.Helper()
	select {
	case got, ok := <-ch:
		if !ok {
			t.Fatalf("channel closed; wanted %#v", want)
		}
		if got != want {
			t.Fatalf("received %#v; want %#v", got, want)
		}
	case <-time.After(within):
		t.Fatalf("timed out after %s waiting to receive %#v", within, want)
	}
}

func requireReceiveInto[T any](t *testing.T, ch <-chan T, dst *T, within time.Duration) {
	t.Helper()
	select {
	case got, ok := <-ch:
		if !ok {
			t.Fatalf("channel closed while waiting to receive")
		}
		*dst = got
	case <-time.After(within):
		t.Fatalf("timed out after %s waiting to receive", within)
	}
}

func requireNotClosed(t *testing.T, ch <-chan struct{}, within time.Duration) {
	t.Helper()
	select {
	case <-ch:
		t.Fatalf("channel closed unexpectedly")
	case <-time.After(within):
		return
	}
}

func requireClosed(t *testing.T, ch <-chan struct{}, within time.Duration) {
	t.Helper()
	select {
	case <-ch:
		return
	case <-time.After(within):
		t.Fatalf("timed out after %s waiting for channel to close", within)
	}
}

func eventually(t *testing.T, within time.Duration, f func() bool) {
	t.Helper()
	deadline := time.Now().Add(within)
	for {
		if f() {
			return
		}
		if time.Now().After(deadline) {
			t.Fatalf("condition not met within %s", within)
		}
		time.Sleep(1 * time.Millisecond)
	}
}

func requireSendWithin[T any](t *testing.T, ch chan<- T, v T, within time.Duration) {
	t.Helper()
	select {
	case ch <- v:
		return
	case <-time.After(within):
		t.Fatalf("timed out after %s waiting to send", within)
	}
}

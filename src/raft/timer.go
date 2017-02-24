package raft

import (
	"time"
)

type Timer interface {
	Channel() <-chan time.Time
	Stop()
}

type getTimerFn func(d time.Duration) Timer

type builtInTimer struct {
	t  *time.Timer
	ch chan time.Time
}

func newBuiltInTimer(d time.Duration) Timer {
	b := new(builtInTimer)
	b.ch = make(chan time.Time)
	b.t = time.NewTimer(d)
	go func() {
		time :=  <-b.t.C
		b.ch <- time
	}()
	return b
}

func (b *builtInTimer) Channel() <-chan time.Time {
	return b.ch
}

func (b *builtInTimer) Stop() {
	b.t.Stop()
	close(b.ch)
}

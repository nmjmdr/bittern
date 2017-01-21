package raft

import (
	"time"
)

type Ticker interface {
	Channel() <-chan time.Time
	Stop()
}

type getTickerFn func(d time.Duration) Ticker

type builtInTicker struct {
	t  *time.Ticker
	ch chan time.Time
}

func newBuiltInTicker(d time.Duration) Ticker {
	b := new(builtInTicker)
	b.ch = make(chan time.Time)
	b.t = time.NewTicker(d)
	go func() {
		for tick := range b.t.C {
			b.ch <- tick
		}
	}()
	return b
}

func (b *builtInTicker) Channel() <-chan time.Time {
	return b.ch
}

func (b *builtInTicker) Stop() {
	b.t.Stop()
	close(b.ch)
}

package raft

import (
	"time"
)

type mockTicker struct {
	ch chan time.Time
}

func newMockTicker(d time.Duration) *mockTicker {
	m := new(mockTicker)
	m.ch = make(chan time.Time)
	return m
}

func (m *mockTicker) Channel() <-chan time.Time {
	return m.ch
}

func (m *mockTicker) tick() {
	m.ch <- time.Time{}
}

func (m *mockTicker) Stop() {
	close(m.ch)
}

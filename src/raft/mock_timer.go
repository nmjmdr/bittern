package raft

import (
	"time"
)

type mockTimer struct {
	ch chan time.Time
}

func newMockTimer(d time.Duration) *mockTimer {
	m := new(mockTimer)
	m.ch = make(chan time.Time)
	return m
}

func (m *mockTimer) Channel() <-chan time.Time {
	return m.ch
}

func (m *mockTimer) tick() {
	m.ch <- time.Time{}
}

func (m *mockTimer) Stop() {
	close(m.ch)
}

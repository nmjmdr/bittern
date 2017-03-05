package raft

import (
  "time"
)

type ElectionTimeoutTimerStartCallback func(t time.Duration)
type mockElectionTimeoutTimer struct {
	callback ElectionTimeoutTimerStartCallback
}

func newMockElectionTimeoutTimer(callback ElectionTimeoutTimerStartCallback) *mockElectionTimeoutTimer {
	m := new(mockElectionTimeoutTimer)
	m.callback = callback
	return m
}

func (m *mockElectionTimeoutTimer) Start(t time.Duration) {
	if m.callback != nil {
		m.callback(t)
	}
}

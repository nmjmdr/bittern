package raft

import ()

type ElectionTimeoutTimerStartCallback func()
type mockElectionTimeoutTimer struct {
	callback ElectionTimeoutTimerStartCallback
}

func newMockElectionTimeoutTimer(callback ElectionTimeoutTimerStartCallback) *mockElectionTimeoutTimer {
	m := new(mockElectionTimeoutTimer)
	m.callback = callback
	return m
}

func (m *mockElectionTimeoutTimer) Start() {
	if m.callback != nil {
		m.callback()
	}
}

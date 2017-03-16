package raft

import (
	"time"
)

type StartCallbackFn func(t time.Duration)
type StopCallbackFn func()
type mockTimer struct {
	startCb StartCallbackFn
	stopCb  StopCallbackFn
}

func newMockElectionTimeoutTimer() *mockTimer {
	m := new(mockTimer)
	return m
}

func (m *mockTimer) Start(t time.Duration) {
	if m.startCb != nil {
		m.startCb(t)
	}
}

func (m *mockTimer) Stop() {
	if m.stopCb != nil {
		m.stopCb()
	}
}

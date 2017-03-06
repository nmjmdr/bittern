package raft

import (
	"time"
)

type StartCallbackFn func(t time.Duration)
type StopCallbackFn func()
type mockElectionTimeoutTimer struct {
	startCb StartCallbackFn
	stopCb  StopCallbackFn
}

func newMockElectionTimeoutTimer(startCb StartCallbackFn, stopCb StopCallbackFn) *mockElectionTimeoutTimer {
	m := new(mockElectionTimeoutTimer)
	m.startCb = startCb
	m.stopCb = stopCb
	return m
}

func (m *mockElectionTimeoutTimer) Start(t time.Duration) {
	if m.startCb != nil {
		m.startCb(t)
	}
}

func (m *mockElectionTimeoutTimer) Stop() {
	if m.stopCb != nil {
		m.stopCb()
	}
}

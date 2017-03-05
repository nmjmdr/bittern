package raft

import (
  "time"
)

type timerCallbackFn func (time.Duration)
type mockTimer struct {
  callback timerCallbackFn
}

func newMockTimer(callback timerCallbackFn) *mockTimer {
  m := new(mockTimer)
  m.callback = callback
  return m
}

func (m *mockTimer) Start(t time.Duration) {
  if m.callback != nil {
    m.callback(t)
  }
}

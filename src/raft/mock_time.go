package raft

type mockTime struct {
  t int64
}

func newMockTime(t int64) *mockTime {
  m := new(mockTime)
  m.t = t
  return m
}

func (m *mockTime) unixNano() int64 {
  return m.t
}

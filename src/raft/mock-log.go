package raft

type mockLog struct {
  lastLogIndex uint
  lastLogTerm uint64
}

func newMockLog(lastLogIndex uint,lastLogTerm uint64) *mockLog {
  m := new(mockLog)
  m.lastLogIndex = lastLogIndex
  m.lastLogTerm = lastLogTerm
  return m
}

func (m *mockLog) LastTerm() uint64 {
  return m.lastLogTerm
}

func (m *mockLog) LastIndex() uint {
  return m.lastLogIndex
}

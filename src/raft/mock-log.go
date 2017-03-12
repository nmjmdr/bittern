package raft

type EntryAtCallbackFn func(index uint64) (entry, bool)
type mockLog struct {
	lastLogIndex uint64
	lastLogTerm  uint64
	entryAtCb    EntryAtCallbackFn
}

func newMockLog(lastLogIndex uint64, lastLogTerm uint64) *mockLog {
	m := new(mockLog)
	m.lastLogIndex = lastLogIndex
	m.lastLogTerm = lastLogTerm
	return m
}

func (m *mockLog) LastTerm() uint64 {
	return m.lastLogTerm
}

func (m *mockLog) LastIndex() uint64 {
	return m.lastLogIndex
}

func (m *mockLog) EntryAt(index uint64) (entry, bool) {
	if m.entryAtCb != nil {
		return m.entryAtCb(index)
	}
	panic("Mock log - entryAtCb was not set, but EntryAt function was invoked, check the test setup")
}

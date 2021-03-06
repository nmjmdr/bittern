package raft

type EntryAtCallbackFn func(index uint64) (entry, bool)
type AddAtCallbackFn func(index uint64, entry entry)
type DeleteFromCallbackFn func(index uint64)
type AppendCallbackFn func(e entry)
type GetCallbackFn func(startIndex uint64) []entry

type mockLog struct {
	lastLogIndex uint64
	lastLogTerm  uint64
	entryAtCb    EntryAtCallbackFn
	addAtCb      AddAtCallbackFn
	deleteFromCb DeleteFromCallbackFn
	appendCb     AppendCallbackFn
	getCb        GetCallbackFn
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
	panic("MockLog.EntryAt called, but the callback was not setup, check test setup")
}

func (m *mockLog) AddAt(index uint64, e entry) {
	if m.addAtCb != nil {
		m.addAtCb(index, e)
		return
	}
}

func (m *mockLog) Append(e entry) {
	if m.appendCb != nil {
		m.appendCb(e)
		return
	}
}

func (m *mockLog) DeleteFrom(index uint64) {
	if m.deleteFromCb != nil {
		m.deleteFromCb(index)
		return
	}
}

func (m *mockLog) Get(startIndex uint64) []entry {
	if m.getCb != nil {
		return m.getCb(startIndex)
	}
	return nil
}

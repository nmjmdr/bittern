package raft

type Log interface {
	LastTerm() uint64
	LastIndex() uint64
	EntryAt(index uint64) (entry, bool)
}

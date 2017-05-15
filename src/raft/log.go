package raft

type Log interface {
	LastTerm() uint64
	LastIndex() uint64
	EntryAt(index uint64) (entry, bool)
	AddAt(index uint64, entry entry)
	DeleteFrom(index uint64)
	Append(entry entry)
	Get(startIndex uint64) []entry
}

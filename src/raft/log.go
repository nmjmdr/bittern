package raft

type Log interface {
	LastTerm() uint64
	LastIndex() uint
}

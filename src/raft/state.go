package raft

type state struct {
	mode        Mode
	commitIndex uint64
	lastApplied uint64
}

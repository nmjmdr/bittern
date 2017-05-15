package raft

type state struct {
	mode                 Mode
	commitIndex          uint64
	lastApplied          uint64
	lastHeardFromALeader int64

	votesGot int

	lastSentAppendEntriesAt int64

	nextIndex  map[string]uint64
	matchIndex map[string]uint64
}

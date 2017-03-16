package raft

type state struct {
	mode                   Mode
	commitIndex            uint64
	lastApplied            uint64
	lastHeardFromALeader   int64
	votesGot               int
	lastSentAppenEntriesAt int64
}

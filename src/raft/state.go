package raft

type state struct {
	commitIndex uint64
	lastApplied uint64
	stFn        stateFunction

	// re-initialized after election as leader
	// leader state
	nextIndex  []uint64
	matchIndex []uint64
}

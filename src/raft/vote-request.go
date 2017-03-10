package raft

type voteRequest struct {
	from         peer
	term         uint64
	lastLogIndex uint
	lastLogTerm  uint64
}

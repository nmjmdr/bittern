package raft

type voteRequest struct {
	from         peer
	term         uint64
	lastLogIndex uint64
	lastLogTerm  uint64
}

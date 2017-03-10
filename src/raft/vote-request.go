package raft

type voteRequest struct {
	from peer
	term uint64
}

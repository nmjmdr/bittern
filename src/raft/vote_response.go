package raft

type voteResponse struct {
	Success bool
	Term    uint64
	From    peer
}

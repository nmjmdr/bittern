package raft

type voteResponse struct {
	success bool
	term    uint64
	from    peer
}

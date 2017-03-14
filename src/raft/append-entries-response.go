package raft

type appendEntriesResponse struct {
	success bool
	term    uint64
}

package raft

type entryResponse struct {
	success bool
	term    uint64
}

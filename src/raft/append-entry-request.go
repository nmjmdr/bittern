package raft

type appendEntryRequest struct {
	from         peer
	term         uint64
	prevLogTerm  uint64
	prevLogIndex uint64
	entries      []entry
}

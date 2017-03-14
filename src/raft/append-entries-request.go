package raft

type appendEntriesRequest struct {
	from         peer
	term         uint64
	prevLogTerm  uint64
	prevLogIndex uint64
	entries      []entry
	leaderCommit uint64
}

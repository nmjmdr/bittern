package raft

type appendEntryRequest struct {
	term         uint64
	leaderId     string
	prevLogTerm  uint64
	entries      []log
	leaderCommit uint64
}

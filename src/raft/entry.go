package raft

type entry struct {
	term    uint64
	command string
}

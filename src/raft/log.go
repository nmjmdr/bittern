package raft

type log struct {
	term    uint64
	command string
}

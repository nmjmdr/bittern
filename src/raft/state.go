package raft

type state struct {
	mode Mode
	term uint64
}

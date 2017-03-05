package raft

type Mode int

const (
	Follower Mode = iota
	Candidate
	Leader
)

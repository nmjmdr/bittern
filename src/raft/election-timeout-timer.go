package raft

type ElectionTimeoutTimer interface {
	Start()
}

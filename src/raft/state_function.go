package raft

type stateFunction interface {
	gotElectionSignal()
}

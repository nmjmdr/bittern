package raft

type stateFunction interface {
	gotElectionSignal(st *state)
}

package raft

type stateFunction interface {
	gotElectionSignal()
	gotVote(evt event)
	gotVoteRequestRejected(evt event)
}

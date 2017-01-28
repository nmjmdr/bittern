package raft

type stateFunction interface {
	gotElectionSignal()
	gotVote(voteResponse)
	gotVoteRequestRejected(voteResponse)
}

package raft

type transport interface {
	askVote(peer peer,vr voteRequest) (voteResponse, error)
}
